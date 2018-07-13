package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pt "git.torproject.org/pluggable-transports/goptlib.git"
	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

type PTServerCommand struct {
	wg sync.WaitGroup
}

func NewPTServerCommand() *PTServerCommand {
	return &PTServerCommand{}
}

func (cmd *PTServerCommand) Run(args []string) error {
	fs := NewFlagSet("marionette-ptserver", flag.ContinueOnError)
	var (
		format  = fs.String("format", "", "Format name and version")
		logFile = fs.String("log-file", "", "Path to log file.")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *format == "" {
		return errors.New("format required")
	}

	if *logFile != "" {
		file, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)

		if err != nil {
			fmt.Errorf("Unable to open log file: %s", err)
		}

		log.SetOutput(file)
		defer file.Close()
	}

	// Read MAR file.
	data, err := mar.ReadFormat(*format)
	if os.IsNotExist(err) {
		return fmt.Errorf("MAR document not found: %s", *format)
	} else if err != nil {
		return err
	}

	// Parse document.
	doc, err := mar.Parse(marionette.PartyServer, data)
	if err != nil {
		return err
	}

	// We always use the production logger when running as a PT.
	config := zap.NewProductionConfig()
	config.DisableStacktrace = true
	marionette.Logger, _ = config.Build()

	// Setup the PT.
	serverInfo, err := pt.ServerSetup(nil)
	if err != nil {
		return err
	}

	listeners := make([]net.Listener, 0)
	for _, bindAddr := range serverInfo.Bindaddrs {
		if bindAddr.MethodName != "marionette" {
			pt.SmethodError(bindAddr.MethodName, "no such method")
		}

		log.Printf("Starting Marionette PT")

		// Marionette always listen on port 8081 so we ignore TOR.
		// This should probably be fixed.
		host, port, err := net.SplitHostPort(bindAddr.Addr.String())
		if err != nil {
			log.Printf("Unable to split host/port: %s", err)
			pt.SmethodError(bindAddr.MethodName, err.Error())
			break
		}

		if port != "8081" {
			log.Printf("Port wasn't 8081")
			pt.SmethodError(bindAddr.MethodName, err.Error())
			break
		}

		// Start the listener.
		listener, err := marionette.Listen(doc, host)

		if err != nil {
			log.Printf("Unable to create listener: %s", err)
			pt.SmethodError(bindAddr.MethodName, err.Error())
			break
		}

		cmd.wg.Add(1)
		go func() { defer cmd.wg.Done(); cmd.acceptLoop(listener, &serverInfo) }()

		pt.Smethod(bindAddr.MethodName, listener.Addr())
		listeners = append(listeners, listener)
	}
	pt.SmethodsDone()

	// Wait for SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	<-sigChan

	// Close listeners and wait for connections to close.
	for _, listener := range listeners {
		listener.Close()
	}
	cmd.wg.Wait()

	return nil
}

func (cmd *PTServerCommand) acceptLoop(listener net.Listener, serverInfo *pt.ServerInfo) {
	defer listener.Close()

	for {
		log.Printf("Entering listener.Accept")
		connection, err := listener.Accept()

		log.Printf("Accepted listener")
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				continue
			}
			log.Printf("Error accepting client: %s", err)
			return
		}

		cmd.wg.Add(1)
		go func() { defer cmd.wg.Done(); cmd.handleConn(connection, serverInfo) }()
	}
}

func (cmd *PTServerCommand) handleConn(connection net.Conn, serverInfo *pt.ServerInfo) {
	log.Printf("Client connected: %s", connection.RemoteAddr())
	defer log.Printf("Client disconnected: %s", connection.RemoteAddr())

	log.Printf("Connecting to Onion Router")
	or, err := pt.DialOr(serverInfo, "127.0.0.1:1234", "marionette")
	if err != nil {
		log.Printf("Unable to connect to Onion Router: %s", err)
		return
	}
	defer or.Close()

	log.Printf("Entering copy loop")
	proxyConns(connection, or)
}

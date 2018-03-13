package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"

	"go.uber.org/zap"

	pt "git.torproject.org/pluggable-transports/goptlib.git"
)

func handleServerConn(connection net.Conn, serverInfo *pt.ServerInfo) {
	log.Printf("Client connected: %s", connection.RemoteAddr())

	handlerChan <- 1
	defer func() {
		handlerChan <- -1
		log.Printf("Client disconnected: %s", connection.RemoteAddr())
	}()

	log.Printf("Connecting to Onion Router")
	or, err := pt.DialOr(serverInfo, "127.0.0.1:1234", "marionette")

	if err != nil {
		log.Printf("Unable to connect to Onion Router: %s", err)
		return
	}

	defer or.Close()

	log.Printf("Entering copy loop")
	copyLoop(connection, or)
}

func acceptServerLoop(listener net.Listener, serverInfo *pt.ServerInfo) {
	defer listener.Close()

	for {
                log.Printf("Entering listener.Accept")
		connection, err := listener.Accept()

                log.Printf("Accepted listener")
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				continue
			}

			log.Printf("Error accepting client: %s", err)
			return
		}

		go handleServerConn(connection, serverInfo)
	}
}

type PTServerCommand struct{}

func NewPTServerCommand() *PTServerCommand {
	return &PTServerCommand{}
}

func (cmd *PTServerCommand) Run(args []string) error {
	fs := flag.NewFlagSet("marionette-ptserver", flag.ContinueOnError)
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
	formatName, formatVersion := mar.SplitFormat(*format)
	data := mar.Format(formatName, formatVersion)
	if data == nil {
		return fmt.Errorf("MAR document not found: %s", formatName)
	}

	// Parse document.
	doc, err := mar.Parse(marionette.PartyServer, data)
	if err != nil {
		return err
	}

	// We always use the production logger when running
	// as a PT.
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	marionette.Logger = logger

	// Setup the PT.
	serverInfo, err := pt.ServerSetup(nil)

	if err != nil {
		return err
	}

	listeners := make([]net.Listener, 0)

	for _, bindAddr := range serverInfo.Bindaddrs {
		if bindAddr.MethodName == "marionette" {
			log.Printf("Starting Marionette PT")

			// Marionette always listen on port 8081,
			// so we ignore what Tor is telling us
			// here. This should probably be fixed.
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

			go acceptServerLoop(listener, &serverInfo)

			pt.Smethod(bindAddr.MethodName, listener.Addr())
			listeners = append(listeners, listener)
		} else {
			pt.SmethodError(bindAddr.MethodName, "no such method")
		}
	}
	pt.SmethodsDone()

	numHandlers := 0
	var sig os.Signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	sig = nil
	for sig == nil {
		select {
		case n := <-handlerChan:
			numHandlers += n
		case sig = <-sigChan:
		}
	}

	for _, listener := range listeners {
		listener.Close()
	}

	return nil
}

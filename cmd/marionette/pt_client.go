package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
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

type PTClientCommand struct {
	wg sync.WaitGroup
}

func NewPTClientCommand() *PTClientCommand {
	return &PTClientCommand{}
}

func (cmd *PTClientCommand) Run(args []string) error {
	fs := flag.NewFlagSet("marionette-pt-client", flag.ContinueOnError)
	var (
		format  = fs.String("format", "", "Format name and version")
		logFile = fs.String("log-file", "", "Path to log file.")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *logFile != "" {
		file, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Errorf("Unable to open log file: %s", err)
		}

		log.SetOutput(file)
		defer file.Close()
	}

	// Validate arguments.
	if *format == "" {
		return errors.New("format required")
	}

	// Read MAR file.
	formatName, formatVersion := mar.SplitFormat(*format)
	data := mar.Format(formatName, formatVersion)
	if data == nil {
		return fmt.Errorf("MAR document not found: %s", formatName)
	}

	// Parse document.
	doc, err := mar.Parse(marionette.PartyClient, data)
	if err != nil {
		return err
	}

	// We always use the production logger when running as a PT.
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	marionette.Logger = logger

	clientInfo, err := pt.ClientSetup(nil)
	if err != nil {
		return err
	} else if clientInfo.ProxyURL != nil {
		return errors.New("marionette: PT proxy url not provided")
	}

	listeners := make([]net.Listener, 0)
	for _, methodName := range clientInfo.MethodNames {
		if methodName != "marionette" {
			pt.CmethodError(methodName, "no such method")
		}

		listener, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			pt.CmethodError(methodName, err.Error())
			break
		}

		cmd.wg.Add(1)
		go func() { defer cmd.wg.Done(); cmd.acceptLoop(listener, doc) }()

		pt.Cmethod(methodName, listener.Version(), listener.Addr())
		listeners = append(listeners, listener)
	}
	pt.CmethodsDone()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	<-sigChan

	for _, listener := range listeners {
		listener.Close()
	}
	cmd.wg.Wait()

	return nil
}

func (cmd *PTClientCommand) acceptLoop(listener *pt.SocksListener, doc *mar.Document) {
	defer listener.Close()

	for {
		connection, err := listener.AcceptSocks()
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Temporary() {
				continue
			}
			return
		}

		cmd.wg.Add(1)
		go func() { defer cmd.wg.Done(); cmd.handleConn(connection, doc) }()
	}
}

func (cmd *PTClientCommand) handleConn(connection *pt.SocksConn, doc *mar.Document) {
	host, _, err := net.SplitHostPort(connection.Req.Target)
	if err != nil {
		log.Printf("Invalid connection request target: %s", connection.Req.Target)
		connection.Reject()
		return
	}

	log.Printf("Connecting to Marionette server: %s", host)
	defer connection.Close()
	defer log.Printf("Disconnected from Marionette host: %s", host)

	streamSet := marionette.NewStreamSet()
	defer streamSet.Close()

	// Create dialer to remote server.
	dialer, err := marionette.NewDialer(doc, host, streamSet)
	if err != nil {
		log.Printf("Unable to create dialer: %s", err)
		connection.Reject()
		return
	}
	defer dialer.Close()
	log.Printf("Connected!")

	// Create a stream through the dialer.
	stream, err := dialer.Dial()
	if err != nil {
		log.Printf("Unable to connect to server: %s", err)
		connection.Reject()
		return
	}
	defer stream.Close()

	// Allow connection.
	if err := connection.Grant(nil); err != nil {
		return
	}

	log.Printf("Proxying stream to connection")
	proxyConns(stream, connection)
}

func proxyConns(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(a, b) }()
	go func() { defer wg.Done(); io.Copy(b, a) }()
	wg.Wait()
}

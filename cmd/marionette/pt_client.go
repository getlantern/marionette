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

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"

	"go.uber.org/zap"

	pt "git.torproject.org/pluggable-transports/goptlib.git"
)

var handlerChan = make(chan int)

func copyLoop(stream net.Conn, or net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(or, stream)
		wg.Done()
	}()

	go func() {
		io.Copy(stream, or)
		wg.Done()
	}()

	wg.Wait()
}

func handleClient(connection *pt.SocksConn, doc *mar.Document) {

	host, _, err := net.SplitHostPort(connection.Req.Target)

	if err != nil {
		connection.Reject()
		return
	}

	handlerChan <- 1

	defer func() {
		handlerChan <- -1
		connection.Close()
		log.Printf("Disconnected from Marionette host: %s", host)
	}()

	streamSet := marionette.NewStreamSet()
	defer streamSet.Close()

	log.Printf("Connecting to Marionette server: %s", host)

	// Create dialer to remote server.
	dialer, err := marionette.NewDialer(doc, host, streamSet)
	if err != nil {
		log.Printf("Unable to create dialer: %s", err)
		connection.Reject()
		return
	}
	defer dialer.Close()
	log.Printf("Connected!")

	stream, err := dialer.Dial()
	if err != nil {
		log.Printf("Unable to connect to server: %s", err)
		connection.Reject()
		return
	}
	defer stream.Close()

	err = connection.Grant(nil)

	if err != nil {
		return
	}

	log.Printf("Entering copyLoop")
	copyLoop(stream, connection)
}

func acceptLoop(listener *pt.SocksListener, doc *mar.Document) {
	defer listener.Close()

	for {
		connection, err := listener.AcceptSocks()

		if err != nil {
			netErr, ok := err.(net.Error)

			if ok && netErr.Temporary() {
				continue
			}

			return
		}

		go handleClient(connection, doc)
	}
}

type PTClientCommand struct{}

func NewPTClientCommand() *PTClientCommand {
	return &PTClientCommand{}
}

func (cmd *PTClientCommand) Run(args []string) error {
	fs := flag.NewFlagSet("marionette-ptclient", flag.ContinueOnError)
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

	// We always use the production logger when running
	// as a PT.
	logger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	marionette.Logger = logger

	clientInfo, err := pt.ClientSetup(nil)

	if err != nil {
		return err
	}

	if clientInfo.ProxyURL != nil {
		// FIXME: Ugh? I wonder if this can actually
		// happen.
		return nil
	}

	listeners := make([]net.Listener, 0)

	for _, methodName := range clientInfo.MethodNames {
		if methodName == "marionette" {
			listener, err := pt.ListenSocks("tcp", "127.0.0.1:0")

			if err != nil {
				pt.CmethodError(methodName, err.Error())
				break
			}

			go acceptLoop(listener, doc)

			pt.Cmethod(methodName, listener.Version(), listener.Addr())
			listeners = append(listeners, listener)
		} else {
			pt.CmethodError(methodName, "no such method")
		}
	}
	pt.CmethodsDone()

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

	for n := range handlerChan {
		numHandlers += n
		if numHandlers == 0 {
			break
		}
	}

	return nil
}

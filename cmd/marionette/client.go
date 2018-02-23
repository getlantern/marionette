package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	_ "github.com/redjack/marionette/plugins"
	"go.uber.org/zap"
)

type ClientCommand struct{}

func NewClientCommand() *ClientCommand {
	return &ClientCommand{}
}

func (cmd *ClientCommand) Run(args []string) error {
	// Parse arguments.
	fs := flag.NewFlagSet("marionette-client", flag.ContinueOnError)
	var (
		version  = fs.Bool("version", false, "")
		bind     = fs.String("bind", "127.0.0.1:8079", "Bind address")
		serverIP = fs.String("server", "127.0.0.1", "Server IP address")
		format   = fs.String("format", "", "Format name and version")
		verbose  = fs.Bool("v", false, "Debug logging enabled")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	// If version is specified, print and exit.
	if *version {
		return printVersion()
	}

	// Validate arguments.
	if *format == "" {
		return errors.New("format required")
	}

	// Strip off format version.
	// TODO: Split version.
	formatName := mar.StripFormatVersion(*format)

	// Read MAR file.
	data := mar.Format(formatName, "")
	if data == nil {
		return fmt.Errorf("MAR document not found: %s", formatName)
	}

	// Parse document.
	doc, err := mar.Parse(marionette.PartyClient, data)
	if err != nil {
		return err
	}

	// Set logger if debug is on.
	if *verbose {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil
		}
		marionette.Logger = logger
	} else {
		logger, err := zap.NewProduction()
		if err != nil {
			return nil
		}
		marionette.Logger = logger
	}

	// Start listener.
	ln, err := net.Listen("tcp", *bind)
	if err != nil {
		return err
	}
	defer ln.Close()

	streamSet := marionette.NewStreamSet()
	defer streamSet.Close()

	// Create dialer to remote server.
	dialer, err := marionette.NewDialer(doc, *serverIP, streamSet)
	if err != nil {
		return err
	}
	defer dialer.Close()

	// Start proxy.
	proxy := marionette.NewClientProxy(ln, dialer)
	if err := proxy.Open(); err != nil {
		return err
	}
	defer proxy.Close()

	fmt.Printf("listening on %s, connected to %s\n", *bind, *serverIP)

	// Wait for signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Fprintln(os.Stderr, "received interrupt, shutting down...")

	return nil
}

// printVersion prints a list of available formats and their versions.
func printVersion() error {
	fmt.Println("Marionette proxy client.")
	fmt.Println("Available formats:")
	for _, format := range mar.Formats() {
		fmt.Printf(" %s\n", format)
	}
	fmt.Println("")
	return nil
}

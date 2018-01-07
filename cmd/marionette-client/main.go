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
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	config := marionette.DefaultConfig()

	// Parse arguments.
	fs := flag.NewFlagSet("marionette-client", flag.ContinueOnError)
	version := fs.Bool("version", false, "")
	fs.StringVar(&config.Client.Bind, "bind", config.Client.Bind, "Bind address")
	fs.StringVar(&config.Server.IP, "server", config.Server.IP, "Server IP address")
	fs.StringVar(&config.General.Format, "format", config.General.Format, "Format name and version")
	fs.BoolVar(&config.General.Debug, "debug", config.General.Debug, "Debug logging enabled")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	// If version is specified, print and exit.
	if *version {
		return printVersion()
	}

	// Validate arguments.
	if config.General.Format == "" {
		return errors.New("format required")
	}

	// Strip off format version.
	// TODO: Split version.
	format := marionette.StripFormatVersion(config.General.Format)

	// Read MAR file.
	data := mar.Format(format, "")
	if data == nil {
		return fmt.Errorf("MAR document not found: %s", format)
	}

	// Parse document.
	doc, err := mar.Parse(marionette.PartyClient, data)
	if err != nil {
		return err
	}

	// Set logger if debug is on.
	if config.General.Debug {
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
	ln, err := net.Listen("tcp", config.Client.Bind)
	if err != nil {
		return err
	}
	defer ln.Close()

	// Create dialer to remote server.
	dialer, err := marionette.NewDialer(doc, config.Server.IP)
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

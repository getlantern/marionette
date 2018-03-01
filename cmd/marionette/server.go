package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/armon/go-socks5"
	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	_ "github.com/redjack/marionette/plugins"
	"go.uber.org/zap"
)

type ServerCommand struct{}

func NewServerCommand() *ServerCommand {
	return &ServerCommand{}
}

func (cmd *ServerCommand) Run(args []string) error {
	// Parse arguments.
	fs := flag.NewFlagSet("marionette-server", flag.ContinueOnError)
	var (
		bind      = fs.String("bind", "", "Bind address")
		useSocks5 = fs.Bool("socks5", false, "Enable socks5 proxying")
		proxyAddr = fs.String("proxy", "", "Proxy IP and port")
		format    = fs.String("format", "", "Format name and version")
		verbose   = fs.Bool("v", false, "Debug logging enabled")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate arguments.
	if *format == "" {
		return errors.New("format required")
	} else if !*useSocks5 && *proxyAddr == "" {
		return errors.New("proxy address required")
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

	// Set logger if verbose.
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
	ln, err := marionette.Listen(doc, *bind)
	if err != nil {
		return err
	}
	defer ln.Close()

	// Start proxy.
	proxy := marionette.NewServerProxy(ln)
	if *useSocks5 {
		if proxy.Socks5Server, err = socks5.New(&socks5.Config{}); err != nil {
			return err
		}
	} else {
		proxy.Addr = *proxyAddr
	}
	if err := proxy.Open(); err != nil {
		return err
	}
	defer proxy.Close()

	// Notify user that proxy is ready.
	if proxy.Socks5Server != nil {
		fmt.Printf("listening on %s, proxying via socks5\n", ln.Addr().String())
	} else {
		fmt.Printf("listening on %s, proxying to %s\n", ln.Addr().String(), *proxyAddr)
	}

	// Wait for signal.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Fprintln(os.Stderr, "received interrupt, shutting down...")

	return nil
}

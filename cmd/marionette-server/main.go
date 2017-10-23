package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/redjack/marionette/assets"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Parse arguments.
	var config marionette.Config
	fs := flag.NewFlagSet("marionette-server", flag.ContinueOnError)
	version := fs.Bool("version", false, "")
	fs.StringVar(&config.Server.ServerIP, "server_ip", config.Server.ServerIP, "")
	fs.StringVar(&config.Server.ServerIP, "sip", config.Server.ServerIP, "")
	fs.StringVar(&config.Server.ProxyPort, "proxy_port", config.Server.ProxyPort, "")
	fs.StringVar(&config.Server.ProxyPort, "pport", config.Server.ProxyPort, "")
	fs.StringVar(&config.Server.ProxyIP, "proxy_ip", config.Server.ProxyIP, "")
	fs.StringVar(&config.Server.ProxyIP, "pip", config.Server.ProxyIP, "")
	fs.StringVar(&config.General.Format, "format", config.General.Format, "")
	fs.StringVar(&config.General.Format, "f", config.General.Format, "")
	fs.BoolVar(&config.General.Debug, "debug", config.General.Debug, "")
	fs.BoolVar(&config.General.Debug, "d", config.General.Debug, "")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	// If version is specified, print and exit.
	if *version {
		return printVersion()
	}

	// Strip off format version.
	format := marionette.StripFormatVersion(config.General.Format)

	if !config.General.Debug {
		log.SetOutput(ioutil.Discard)
	}

	// Start server.
	server := marionette.NewServer(format)
	server.factory = ProxyServer
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// printVersion prints a list of available formats and their versions.
func printVersion() error {
	fmt.Println("Marionette proxy server.")
	fmt.Println("Available formats:")
	for _, format := range assets.Formats() {
		fmt.Printf(" %s", format)
	}
	return nil
}

/*
class ProxyServerProtocol(protocol.Protocol):

    def connectionMade(self):
        log.msg("ProxyServerProtocol: connected to peer")
        self.cli_queue = self.factory.cli_queue
        self.cli_queue.get().addCallback(self.serverDataReceived)

    def serverDataReceived(self, chunk):
        if chunk is False:
            self.cli_queue = None
            log.msg("ProxyServerProtocol: disconnecting from peer")
            self.factory.continueTrying = False
            self.transport.loseConnection()
        elif self.cli_queue:
            log.msg(
                "ProxyServerProtocol: writing %d bytes to peer" %
                len(chunk))

            self.transport.write(chunk)
            self.cli_queue.get().addCallback(self.serverDataReceived)
        else:
            log.msg(
                "ProxyServerProtocol: (2) writing %d bytes to peer" %
                len(chunk))
            self.factory.cli_queue.put(chunk)

    def dataReceived(self, chunk):
        log.msg(
            "ProxyServerProtocol: %d bytes received from peer" %
            len(chunk))
        self.factory.srv_queue.put(chunk)

    def connectionLost(self, why):
        log.msg("ProxyServerProtocol.connectionLost: " + str(why))
        if self.cli_queue:
            self.cli_queue = None
            log.msg("ProxyServerProtocol: peer disconnected unexpectedly")


class ProxyServerFactory(protocol.ClientFactory):
    protocol = ProxyServerProtocol

    def __init__(self, srv_queue, cli_queue):
        self.srv_queue = srv_queue
        self.cli_queue = cli_queue


class ProxyServer(object):

    def __init__(self):
        self.connector = None

    def connectionMade(self, marionette_stream):
        log.msg("ProxyServer.connectionMade")
        self.cli_queue = defer.DeferredQueue()
        self.srv_queue = defer.DeferredQueue()
        self.marionette_stream = marionette_stream
        self.srv_queue.get().addCallback(self.clientDataReceived)

        self.factory = ProxyServerFactory(self.srv_queue, self.cli_queue)
        self.connector = reactor.connectTCP(
            REMOTE_IP,
            REMOTE_PORT,
            self.factory)

    def clientDataReceived(self, chunk):
        log.msg(
            "ProxyServer.clientDataReceived: writing %d bytes to original client" %
            len(chunk))
        self.marionette_stream.push(chunk)
        self.srv_queue.get().addCallback(self.clientDataReceived)

    def dataReceived(self, chunk):
        log.msg("ProxyServer.dataReceived: %s bytes" % len(chunk))
        self.cli_queue.put(chunk)

    def connectionLost(self):
        log.msg("ProxyServer.connectionLost")
        self.cli_queue.put(False)
        self.connector.disconnect()

*/

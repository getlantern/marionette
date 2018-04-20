package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/redjack/marionette/plugins/model"
)

var ErrUsage = errors.New("usage")

func main() {
	if err := run(os.Args[1:]); err == ErrUsage {
		fmt.Fprintln(os.Stderr, Usage())
		os.Exit(1)
	} else if err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	switch args[0] {
	case "client":
		return NewClientCommand().Run(args[1:])
	case "formats":
		return NewFormatsCommand().Run(args[1:])
	case "pt-client":
		return NewPTClientCommand().Run(args[1:])
	case "pt-server":
		return NewPTServerCommand().Run(args[1:])
	case "server":
		return NewServerCommand().Run(args[1:])
	default:
		return ErrUsage
	}
}

func Usage() string {
	return `
Marionette is a programmable client-server proxy that enables the user to
control network traffic features with a lightweight domain-specific language.

Usage:

	marionette command [arguments]

The commands are:

	client    runs the client proxy
	formats   show a list of available formats
	pt-client runs the client proxy as a PT
	pt-server runs the server proxy as a PT
	server    runs the server proxy
`[1:]
}

type FlagSet struct {
	*flag.FlagSet
	Pprof bool
}

func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	fs := &FlagSet{FlagSet: flag.NewFlagSet(name, errorHandling)}
	fs.Float64Var(&model.SleepFactor, "sleep-factor", model.SleepFactor, "model.sleep() multipler")
	fs.BoolVar(&fs.Pprof, "pprof", false, "enable pprof debugging")
	return fs
}

func (fs *FlagSet) Parse(arguments []string) error {
	if err := fs.FlagSet.Parse(arguments); err != nil {
		return err
	}

	// Run pprof-server in the background if requested.
	if fs.Pprof {
		fmt.Println("pprof listening on http://localhost:6060")
		go func() { http.ListenAndServe("localhost:6060", nil) }()
	}

	return nil
}

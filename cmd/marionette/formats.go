package main

import (
	"flag"
	"fmt"

	"github.com/redjack/marionette/mar"
)

type FormatsCommand struct{}

func NewFormatsCommand() *FormatsCommand {
	return &FormatsCommand{}
}

func (cmd *FormatsCommand) Run(args []string) error {
	fs := flag.NewFlagSet("marionette-formats", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	for _, format := range mar.Formats() {
		fmt.Println(format)
	}
	return nil
}

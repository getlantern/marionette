package io

import (
	"errors"

	"github.com/redjack/marionette"
)

func init() {
	marionette.RegisterPlugin("io", "puts", Puts)
}

func Puts(fsm marionette.FSM, args ...interface{}) (success bool, err error) {
	if len(args) < 1 {
		return false, errors.New("io.puts: not enough arguments")
	}

	data, ok := args[0].(string)
	if !ok {
		return false, errors.New("io.puts: invalid argument type")
	}

	// Keep attempting to send even if there are timeouts.
	for len(data) > 0 {
		n, err := fsm.Conn().Write([]byte(data))
		data = data[n:]
		if isTimeoutError(err) {
			continue
		} else if err != nil {
			return false, err
		}
	}

	return true, nil
}

// isTimeoutError returns true if the error is a timeout error.
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	} else if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
		return true
	}
	return false
}

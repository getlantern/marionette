package channel

import (
	"errors"

	"github.com/redjack/marionette"
)

func init() {
	marionette.RegisterPlugin("channel", "bind", Bind)
}

// Bind binds the variable specified in the first argument to a port.
func Bind(fsm marionette.FSM, args ...interface{}) error {
	if len(args) < 1 {
		return errors.New("channel.bind: not enough arguments")
	}

	name, ok := args[0].(string)
	if !ok {
		return errors.New("channel.bind: invalid argument type")
	}

	// Ignore if variable is already bound.
	if value := fsm.Var(name); value != nil {
		if i, _ := value.(int); i > 0 {
			return nil
		}
	}

	// Create a new connection on a random port.
	port, err := fsm.Listen()
	if err != nil {
		return err
	}

	// Save port number to variables.
	fsm.SetVar(name, port)

	return nil
}

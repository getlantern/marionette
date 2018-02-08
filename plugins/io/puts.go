package marionette

import (
	"errors"

	"github.com/redjack/marionette"
)

func init() {
	marionette.RegisterPlugin("io", "puts", Puts)
}

func Puts(fsm marionette.FSM, args []interface{}) (success bool, err error) {
	conn := fsm.Conn()
	if conn == nil {
		return false, nil
	}

	if len(args) < 1 {
		return false, errors.New("io.puts: not enough arguments")
	}

	data, ok := args[0].(string)
	if !ok {
		return false, errors.New("io.puts: invalid argument type")
	}

	// Keep attempting to send even if there are timeouts.
	for len(data) > 0 {
		n, err := conn.Write([]byte(data))
		data = data[n:]
		if e, ok := err.(interface {
			Timeout() bool
		}); ok && e.Timeout() {
			continue
		} else if err != nil {
			return false, err
		}
	}

	return true, nil
}

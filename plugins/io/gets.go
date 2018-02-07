package marionette

import (
	"bytes"
	"errors"
	"io"

	"github.com/redjack/marionette"
)

func init() {
	marionette.RegisterPlugin("io", "gets", Gets)
}

func Gets(fsm *marionette.FSM, args []interface{}) (success bool, err error) {
	conn := fsm.Conn()
	if conn == nil {
		return false, nil
	}

	if len(args) < 1 {
		return false, errors.New("io.gets: not enough arguments")
	}
	exp, ok := args[0].(string)
	if !ok {
		return false, errors.New("io.gets: invalid argument type")
	}

	// Read enough bytes to see if our expected data comes through.
	buf := make([]byte, len(exp))
	if _, err := io.ReadFull(conn, buf); err != nil {
		fsm.SetReadBuffer(buf)
		return false, err
	}

	// If bytes don't equal what we expect then shift them back on the buffer.
	if !bytes.Equal([]byte(exp), buf) {
		fsm.SetReadBuffer(buf)
		return false, nil
	}

	return true, nil
}

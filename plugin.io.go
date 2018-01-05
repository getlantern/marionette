package marionette

import (
	"bytes"
	"errors"
	"io"
)

func IOPutsPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	if fsm.conn == nil {
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
		n, err := fsm.conn.Write([]byte(data))
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

func IOGetsPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	if fsm.conn == nil {
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
	if _, err := io.ReadFull(fsm.conn, buf); err != nil {
		fsm.SetBuffer(buf)
		return false, err
	}

	// If bytes don't equal what we expect then shift them back on the buffer.
	if !bytes.Equal([]byte(exp), buf) {
		fsm.SetBuffer(buf)
		return false, nil
	}

	return true, nil
}

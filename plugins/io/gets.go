package marionette

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/redjack/marionette"
)

func init() {
	marionette.RegisterPlugin("io", "gets", Gets)
}

func Gets(fsm marionette.FSM, args []interface{}) (success bool, err error) {
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

	// Read buffer to see if our expected data comes through.
	buf, err := conn.Peek(len(exp))
	if err == bufio.ErrBufferFull {
		return false, nil
	} else if err != nil {
		return false, err
	} else if !bytes.Equal([]byte(exp), buf) {
		return false, nil
	}

	// Move buffer forward.
	if _, err := conn.Seek(int64(len(buf)), io.SeekCurrent); err != nil {
		return false, err
	}
	return true, nil
}

package io

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("io", "gets", Gets)
}

func Gets(fsm marionette.FSM, args ...interface{}) error {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()), zap.String("state", fsm.State()))

	if len(args) < 1 {
		return errors.New("io.gets: not enough arguments")
	}
	exp, ok := args[0].(string)
	if !ok {
		return errors.New("io.gets: invalid argument type")
	}

	// Read buffer to see if our expected data comes through.
	buf, err := fsm.Conn().Peek(len(exp))
	if err == bufio.ErrBufferFull {
		return nil
	} else if err != nil {
		return err
	} else if !bytes.Equal([]byte(exp), buf) {
		return fmt.Errorf("io.gets: unexpected data: %q", buf)
	}

	// Move buffer forward.
	if _, err := fsm.Conn().Seek(int64(len(buf)), io.SeekCurrent); err != nil {
		return err
	}

	logger.Debug("io.gets", zap.Int("n", len(buf)))
	return nil
}

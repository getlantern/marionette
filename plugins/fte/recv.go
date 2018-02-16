package fte

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("fte", "recv", Recv)
	marionette.RegisterPlugin("fte", "recv_async", Recv)
}

// Recv receives data from a connection.
func Recv(fsm marionette.FSM, args ...interface{}) error {
	return recv(fsm, args)
}

func recv(fsm marionette.FSM, args []interface{}) error {
	t0 := time.Now()

	logger := marionette.Logger.With(
		zap.String("plugin", "fte.recv"),
		zap.String("party", fsm.Party()),
		zap.String("state", fsm.State()),
	)

	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	regex, ok := args[0].(string)
	if !ok {
		return errors.New("invalid regex argument type")
	}
	if _, ok := args[1].(int); !ok {
		return errors.New("invalid msg_len argument type")
	}

	// Retrieve data from the connection.
	conn := fsm.Conn()
	ciphertext, err := conn.Peek(-1)
	if err != nil {
		logger.Error("cannot read from connection", zap.Error(err))
		return err
	}

	// Decode ciphertext.
	plaintext, remainder, err := fsm.Cipher(regex).Decrypt(ciphertext)
	if err != nil {
		logger.Error("cannot decrypt ciphertext", zap.Error(err))
		return err
	}

	// Unmarshal data.
	var cell marionette.Cell
	if err := cell.UnmarshalBinary(plaintext); err != nil {
		logger.Error("cannot unmarshal cell", zap.Error(err))
		return err
	}

	// Validate that the FSM & cell document UUIDs match.
	if fsm.UUID() != cell.UUID {
		logger.Error("uuid mismatch", zap.Int("local", fsm.UUID()), zap.Int("remote", cell.UUID))
		return marionette.ErrUUIDMismatch
	}

	// Set instance ID if it hasn't been set yet.
	// Validate ID if one has already been set.
	if fsm.InstanceID() == 0 {
		fsm.SetInstanceID(cell.InstanceID)
		return marionette.ErrRetryTransition
	} else if fsm.InstanceID() != cell.InstanceID {
		logger.Error("instance id mismatch", zap.Int("local", fsm.InstanceID()), zap.Int("remote", cell.InstanceID))
		return fmt.Errorf("instance id mismatch: fsm=%d, cell=%d", fsm.InstanceID(), cell.InstanceID)
	}

	// Write plaintext to a cell decoder pipe.
	if err := fsm.StreamSet().Enqueue(&cell); err != nil {
		logger.Error("cannot enqueue cell", zap.Error(err))
		return err
	}

	// Move buffer forward by bytes consumed by the cipher.
	if _, err := conn.Seek(int64(len(ciphertext)-len(remainder)), io.SeekCurrent); err != nil {
		logger.Error("cannot move buffer forward", zap.Error(err))
		return err
	}

	logger.Debug("msg received",
		zap.Int("plaintext", len(cell.Payload)),
		zap.Int("ciphertext", len(ciphertext)),
		zap.Duration("t", time.Since(t0)),
	)

	return nil
}

package fte

import (
	"errors"
	"fmt"
	"io"

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
	logger := marionette.Logger.With(zap.String("party", fsm.Party()), zap.String("state", fsm.State()))

	if len(args) < 2 {
		return errors.New("fte.recv: not enough arguments")
	}

	regex, ok := args[0].(string)
	if !ok {
		return errors.New("fte.recv: invalid regex argument type")
	}
	msgLen, ok := args[1].(int)
	if !ok {
		return errors.New("fte.recv: invalid msg_len argument type")
	}

	// Retrieve data from the connection.
	conn := fsm.Conn()
	ciphertext, err := conn.Peek(-1)
	if err != nil {
		return err
	}

	// Decode ciphertext.
	cipher, err := fsm.Cipher(regex, msgLen)
	if err != nil {
		return err
	}
	plaintext, remainder, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return err
	}

	// Unmarshal data.
	var cell marionette.Cell
	if err := cell.UnmarshalBinary(plaintext); err != nil {
		return err
	}

	// Validate that the FSM & cell document UUIDs match.
	if fsm.UUID() != cell.UUID {
		return fmt.Errorf("uuid mismatch: fsm=%d, cell=%d", fsm.UUID(), cell.UUID)
	}

	// Set instance ID if it hasn't been set yet.
	// Validate ID if one has already been set.
	if fsm.InstanceID() == 0 {
		fsm.SetInstanceID(cell.InstanceID)
		return marionette.ErrRetryTransition
	} else if fsm.InstanceID() != cell.InstanceID {
		return fmt.Errorf("instance id mismatch: fsm=%d, cell=%d", fsm.InstanceID(), cell.InstanceID)
	}

	// Write plaintext to a cell decoder pipe.
	if err := fsm.StreamSet().Enqueue(&cell); err != nil {
		return err
	}

	// Move buffer forward by bytes consumed by the cipher.
	if _, err := conn.Seek(int64(len(ciphertext)-len(remainder)), io.SeekCurrent); err != nil {
		return err
	}

	logger.Debug("fte.recv",
		zap.Int("plaintext", len(cell.Payload)),
		zap.Int("ciphertext", len(ciphertext)),
	)

	return nil
}

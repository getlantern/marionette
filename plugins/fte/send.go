package fte

import (
	"errors"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("fte", "send", Send)
	marionette.RegisterPlugin("fte", "send_async", SendAsync)
}

// Send sends data to a connection.
func Send(fsm marionette.FSM, args ...interface{}) error {
	return send(fsm, args, true)
}

// SendAsync send data to a connection without blocking.
func SendAsync(fsm marionette.FSM, args ...interface{}) error {
	return send(fsm, args, false)
}

func send(fsm marionette.FSM, args []interface{}, blocking bool) error {
	logger := marionette.Logger.With(
		zap.String("plugin", "fte.send"),
		zap.Bool("blocking", blocking),
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
	msgLen, ok := args[1].(int)
	if !ok {
		return errors.New("invalid msg_len argument type")
	}

	// Find random stream id with data.
	cipher, err := fsm.Cipher(regex, msgLen)
	if err != nil {
		return err
	}

	// If asynchronous, keep trying to read a cell until there is data.
	// If synchronous, send an empty cell if there is no data.
	var cell *marionette.Cell
	for {
		notify := fsm.StreamSet().WriteNotify()

		logger.Debug("dequeuing cell")

		capacity, err := cipher.Capacity()
		if err != nil {
			return err
		}

		cell = fsm.StreamSet().Dequeue(capacity)
		if cell != nil {
			break
		} else if !blocking {
			logger.Debug("no cell, sending empty cell")
			cell = marionette.NewCell(0, 0, 0, marionette.NORMAL)
			break
		}

		// Wait until new data is available if blocking.
		<-notify
	}

	// Assign fsm data to cell.
	cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID()

	// Encode to binary.
	plaintext, err := cell.MarshalBinary()
	if err != nil {
		return err
	}

	// Encrypt using FTE cipher.
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return err
	}

	// Write to outgoing connection.
	if _, err := fsm.Conn().Write(ciphertext); err != nil {
		return err
	}

	logger.Debug("msg sent",
		zap.Int("plaintext", len(cell.Payload)),
		zap.Int("ciphertext", len(ciphertext)),
	)
	return nil
}

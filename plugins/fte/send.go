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
func Send(fsm marionette.FSM, args ...interface{}) (success bool, err error) {
	return send(fsm, args, true)
}

// SendAsync send data to a connection without blocking.
func SendAsync(fsm marionette.FSM, args ...interface{}) (success bool, err error) {
	return send(fsm, args, false)
}

func send(fsm marionette.FSM, args []interface{}, blocking bool) (success bool, err error) {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()))

	if len(args) < 2 {
		return false, errors.New("fte.send: not enough arguments")
	}

	regex, ok := args[0].(string)
	if !ok {
		return false, errors.New("fte.send: invalid regex argument type")
	}
	msgLen, ok := args[1].(int)
	if !ok {
		return false, errors.New("fte.send: invalid msg_len argument type")
	}

	// Find random stream id with data.
	cipher, err := fsm.Cipher(regex, msgLen)
	if err != nil {
		return false, err
	}

	// If asynchronous, keep trying to read a cell until there is data.
	// If synchronous, send an empty cell if there is no data.
	var cell *marionette.Cell
	for {
		notify := fsm.StreamSet().WriteNotify()

		logger.Debug("fte.send: dequeuing cell")
		cell = fsm.StreamSet().Dequeue(cipher.Capacity())
		if cell != nil {
			break
		} else if !blocking {
			logger.Debug("fte.send: no cell, sending empty cell")
			cell = marionette.NewCell(0, 0, 0, marionette.NORMAL)
			break
		}

		// Wait until new data is available if blocking.
		<-notify
	}

	// Assign fsm data to cell.
	cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID()

	logger.Info("fte.send: marshaling cell", zap.Int("n", len(cell.Payload)))

	// Encode to binary.
	plaintext, err := cell.MarshalBinary()
	if err != nil {
		return false, err
	}

	logger.Debug("fte.send: encrypting cell")

	// Encrypt using FTE cipher.
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return false, err
	}

	logger.Debug("fte.send: writing cell data")

	// Write to outgoing connection.
	if _, err := fsm.Conn().Write(ciphertext); err != nil {
		return false, err
	}

	logger.Debug("fte.send: cell data written")
	return true, nil
}

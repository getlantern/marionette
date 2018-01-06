package marionette

import (
	"errors"
	"time"

	"go.uber.org/zap"
)

const MaxCellLength = 262144

// FTESendPlugin send data to a connection.
func FTESendPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteSendPlugin(fsm, args, true)
}

// FTESendAsyncPlugin send data to a connection without blocking.
func FTESendAsyncPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteSendPlugin(fsm, args, false)
}

func fteSendPlugin(fsm *FSM, args []interface{}, isSync bool) (success bool, err error) {
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
	var cell *Cell
	for {
		cell = fsm.streams.GenerateCell(cipher.Capacity())
		if cell != nil {
			break
		} else if isSync {
			cell = NewCell(0, 0, 0, NORMAL)
			break
		}

		// TODO: Synchronize using a channel.
		time.Sleep(100 * time.Millisecond)
	}

	// Assign fsm data to cell.
	cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID

	// Encode to binary.
	plaintext, err := cell.MarshalBinary()
	if err != nil {
		return false, err
	}

	// Encrypt using FTE cipher.
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return false, err
	}

	// Write to outgoing connection.
	if _, err := fsm.conn.Write(ciphertext); err != nil {
		return false, err
	}
	return true, nil
}

// FTERecvPlugin receives data from a connection.
func FTERecvPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteRecvPlugin(fsm, args)
}

// FTERecvAsyncPlugin receives data from a connection without blocking.
func FTERecvAsyncPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteRecvPlugin(fsm, args)
}

func fteRecvPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	logger := fsm.logger()

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

	logger.Debug("fte.recv: reading buffer")

	// Retrieve data from the connection.
	ciphertext, err := fsm.ReadBuffer()
	if err != nil {
		return false, err
	} else if len(ciphertext) == 0 {
		return false, nil
	}

	logger.Debug("fte.recv: buffer read", zap.Int("n", len(ciphertext)))

	// Decode ciphertext.
	cipher, err := fsm.Cipher(regex, msgLen)
	if err != nil {
		return false, err
	}
	plaintext, remainder, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return false, err
	}
	logger.Debug("fte.recv: buffer decoded", zap.Int("plaintext", len(plaintext)), zap.Int("remainder", len(remainder)))

	// Unmarshal data.
	var cell Cell
	if err := cell.UnmarshalBinary(plaintext); err != nil {
		return false, err
	}

	logger.Debug("fte.recv: received cell", zap.Int("payload", len(cell.Payload)))

	assert(fsm.UUID() == cell.UUID)
	fsm.InstanceID = cell.InstanceID
	if fsm.InstanceID == 0 || cell.StreamID == 0 {
		return false, nil
	}

	// Write plaintext to a cell decoder pipe.
	fsm.streams.AddCell(&cell)

	// Push any additional bytes back onto the FSM's read buffer.
	fsm.SetReadBuffer(remainder)

	return true, nil
}

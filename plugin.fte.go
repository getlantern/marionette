package marionette

import (
	"errors"
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

func fteSendPlugin(fsm *FSM, args []interface{}, blocking bool) (success bool, err error) {
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

	cell := fsm.streams.GenerateCell(cipher.Capacity() /*blocking*/)
	if cell == nil {
		fsm.logger().Debug("fte.send: no data available")
		return false, nil
	}
	cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID

	plaintext, err := cell.MarshalBinary()
	if err != nil {
		return false, err
	}

	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return false, err
	}

	if _, err := fsm.conn.Write(ciphertext); err != nil {
		return false, err
	}
	return true, nil
}

// FTERecvPlugin receives data from a connection.
func FTERecvPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteRecvPlugin(fsm, args, true)
}

// FTERecvAsyncPlugin receives data from a connection without blocking.
func FTERecvAsyncPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	return fteRecvPlugin(fsm, args, false)
}

func fteRecvPlugin(fsm *FSM, args []interface{}, blocking bool) (success bool, err error) {
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

	// Retrieve data from the connection.
	ciphertext, err := fsm.ReadBuffer()
	if err != nil {
		return false, err
	} else if len(ciphertext) == 0 {
		return false, nil
	}

	// Decode ciphertext.
	cipher, err := fsm.Cipher(regex, msgLen)
	if err != nil {
		return false, err
	}
	plaintext, remainder, err := cipher.Decrypt(ciphertext)
	if err != nil {
		return false, err
	}

	// Unmarshal data.
	var cell Cell
	if err := cell.UnmarshalBinary(plaintext); err != nil {
		return false, err
	}

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

package tg

import (
	"errors"
	"io"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("tg", "recv", Recv)
}

func Recv(fsm marionette.FSM, args ...interface{}) (success bool, err error) {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()))
	conn := fsm.Conn()

	if len(args) < 1 {
		return false, errors.New("tg.recv: not enough arguments")
	}

	name, ok := args[0].(string)
	if !ok {
		return false, errors.New("tg.recv: invalid grammar name argument type")
	}

	logger.Debug("tg.recv: reading buffer", zap.String("grammar", name))

	// Retrieve grammar by name.
	grammar := grammars[name]
	if grammar == nil {
		return false, errors.New("tg.recv: grammar not found")
	}

	// Retrieve data from the connection.
	ciphertext, err := conn.Peek(-1)
	if err != nil {
		return false, err
	}

	logger.Debug("tg.recv: buffer read", zap.Int("n", len(ciphertext)))

	// Verify incoming data can be parsed by the grammar.
	m := Parse(grammar.Name, string(ciphertext))
	if m == nil {
		logger.Debug("tg.recv: cannot parse buffer")
		return false, nil
	}

	// Execute each cipher against the data.
	var data []byte
	for _, cipher := range grammar.Ciphers {
		logger.Debug("tg.recv: handle buffer", zap.String("name", cipher.Key()))

		if buf, err := cipher.Decrypt(fsm, []byte(m[cipher.Key()])); err != nil {
			return false, err
		} else if len(buf) != 0 {
			data = append(data, buf...)
		}
	}

	// If any handlers matched and returned data then decode data as a cell.
	if len(data) > 0 {
		logger.Debug("tg.recv: decoding buffer")
		var cell marionette.Cell
		if err := cell.UnmarshalBinary(data); err != nil {
			return false, err
		}
		logger.Debug("tg.recv: buffer decoded", zap.Int("n", len(cell.Payload)))

		if cell.UUID != fsm.UUID() {
			return false, marionette.ErrUUIDMismatch
		}

		if fsm.InstanceID() == 0 {
			if cell.InstanceID == 0 {
				return false, errors.New("msg instance id required")
			}
			fsm.SetInstanceID(cell.InstanceID)
		}

		if err := fsm.StreamSet().Enqueue(&cell); err != nil {
			return false, err
		}
	}

	// Clear FSM's read buffer on success.
	if _, err := conn.Seek(int64(len(ciphertext)), io.SeekCurrent); err != nil {
		return false, err
	}

	logger.Debug("tg.recv: recv complete")

	return true, nil
}

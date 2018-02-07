package marionette

import (
	"errors"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("tg", "recv", Recv)
}

func Recv(fsm *marionette.FSM, args []interface{}) (success bool, err error) {
	logger := fsm.Logger()

	if len(args) < 1 {
		return false, errors.New("tg.send: not enough arguments")
	}

	grammar, ok := args[0].(string)
	if !ok {
		return false, errors.New("tg.send: invalid grammar argument type")
	}

	logger.Debug("tg.recv: reading buffer", zap.String("grammar", grammar))

	// Retrieve data from the connection.
	ciphertext, err := fsm.ReadBuffer()
	if err != nil {
		return false, err
	}

	logger.Debug("tg.recv: buffer read", zap.Int("n", len(ciphertext)))

	// Verify incoming data can be parsed by the grammar.
	if parse(grammar, string(ciphertext)) == nil {
		logger.Debug("tg.recv: cannot parse buffer")
		return false, nil
	}

	// Execute each handler against the data.
	var cell_str string
	for _, h := range tgConfigs[grammar].handlers {
		logger.Debug("tg.recv: handle buffer", zap.String("name", h.name))
		s, err := execute_handler_receiver(fsm, grammar, h.name, string(ciphertext))
		if err != nil {
			return false, err
		} else if s != "" {
			cell_str += s
		}
	}

	// If any handlers matched and returned data then decode data as a cell.
	if len(cell_str) > 0 {
		logger.Debug("tg.recv: decoding buffer")
		var cell marionette.Cell
		if err := cell.UnmarshalBinary([]byte(cell_str)); err != nil {
			return false, err
		}
		logger.Debug("tg.recv: buffer decoded", zap.Int("n", len(cell.Payload)))

		assert(cell.UUID == fsm.UUID())
		fsm.InstanceID = cell.InstanceID

		if fsm.InstanceID == 0 {
			return false, nil
		}

		if err := fsm.StreamSet().Enqueue(&cell); err != nil {
			return false, err
		}
	}

	// Clear FSM's read buffer on success.
	fsm.SetReadBuffer(nil)

	logger.Debug("tg.recv: recv complete")

	return true, nil
}

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

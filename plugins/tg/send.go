package tg

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("tg", "send", Send)
}

func Send(fsm marionette.FSM, args ...interface{}) (success bool, err error) {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()))

	if len(args) < 1 {
		return false, errors.New("tg.send: not enough arguments")
	}

	name, ok := args[0].(string)
	if !ok {
		return false, errors.New("tg.send: invalid grammar name argument type")
	}

	// Find grammar by name.
	grammar := grammars[name]
	if grammar == nil {
		return false, errors.New("tg.send: grammar not found")
	} else if len(grammar.Templates) == 0 {
		return false, fmt.Errorf("tg.send: grammar has no templates: %q", grammar)
	}

	// Randomly choose template and replace embedded placeholders.
	ciphertext := grammar.Templates[rand.Intn(len(grammar.Templates))]
	for _, cipher := range grammar.Ciphers {
		if ciphertext, err = execute_handler_sender(fsm, cipher, ciphertext); err != nil {
			return false, fmt.Errorf("tg.send: execute handler sender: %q", err)
		}
	}

	logger.Debug("tg.send: writing cell data")

	// Write to outgoing connection.
	if _, err := fsm.Conn().Write([]byte(ciphertext)); err != nil {
		return false, err
	}

	logger.Debug("tg.send: cell data written")
	return true, nil
}

func execute_handler_sender(fsm marionette.FSM, cipher Cipher, template string) (_ string, err error) {
	// Encode data from streams if there is capacity in the handler.
	var data []byte
	if capacity := cipher.Capacity(); capacity > 0 {
		cell := fsm.StreamSet().Dequeue(cipher.Capacity())
		if cell == nil {
			cell = marionette.NewCell(0, 0, 0, marionette.NORMAL)
		}

		// Assign ids and marshal to bytes.
		cell.UUID, cell.InstanceID = fsm.UUID(), fsm.InstanceID()
		if data, err = cell.MarshalBinary(); err != nil {
			return "", err
		}
	}

	value, err := cipher.Encrypt(fsm, template, data)
	if err != nil {
		return "", err
	}
	return strings.Replace(template, "%%"+cipher.Key()+"%%", string(value), -1), nil
}

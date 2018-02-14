package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("model", "spawn", Spawn)
}

func Spawn(fsm marionette.FSM, args ...interface{}) error {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()), zap.String("state", fsm.State()))

	if len(args) < 2 {
		return errors.New("model.spawn: not enough arguments")
	}

	formatName, ok := args[0].(string)
	if !ok {
		return errors.New("model.spawn: invalid format name argument type")
	}

	n, ok := args[1].(int)
	if !ok {
		return errors.New("model.spawn: invalid count argument type")
	}

	// Find & parse format.
	data := mar.Format(formatName, "")
	if len(data) == 0 {
		return fmt.Errorf("model.spawn: format not found: %q", formatName)
	}
	doc, err := mar.NewParser(fsm.Party()).Parse(data)
	if err != nil {
		return err
	}
	doc.Format = formatName

	// Execute a sub-FSM multiple times.
	for i := 0; i < n; i++ {
		logger.Debug("model.spawn: executing", zap.Int("i", i))

		child := fsm.Clone(doc)
		if err := child.Execute(context.TODO()); err != nil {
			child.Reset()
			return err
		}
		child.Reset()
	}

	return nil
}

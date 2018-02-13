package model_test

import (
	"context"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/model"
)

func TestSpawn(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.ResetFn = func() {}

		var executeN int
		fsm.CloneFn = func(doc *mar.Document) marionette.FSM {
			if doc.Format != `ftp_pasv_transfer` {
				t.Fatalf("unexpected format: %s", doc.Format)
			}
			other := mock.FSM{
				ExecuteFn: func(ctx context.Context) error {
					executeN++
					return nil
				},
				ResetFn: func() {},
			}
			return &other
		}

		if ok, err := model.Spawn(&fsm, "ftp_pasv_transfer", 5); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal("expected success")
		} else if executeN != 5 {
			t.Fatalf("unexpected execution count: %d", executeN)
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if _, err := model.Spawn(&fsm); err == nil || err.Error() != `model.spawn: not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		t.Run("format", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if _, err := model.Spawn(&fsm, 123, 456); err == nil || err.Error() != `model.spawn: invalid format name argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})

		t.Run("count", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if _, err := model.Spawn(&fsm, "fmt", "xyz"); err == nil || err.Error() != `model.spawn: invalid count argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})
	})
}

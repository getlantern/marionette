package model_test

import (
	"context"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/model"
	"go.uber.org/zap"
)

func init() {
	if !testing.Verbose() {
		marionette.Logger = zap.NewNop()
	}
}

func TestSpawn(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
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

		if err := model.Spawn(context.Background(), &fsm, "ftp_pasv_transfer", 5); err != nil {
			t.Fatal(err)
		} else if executeN != 5 {
			t.Fatalf("unexpected execution count: %d", executeN)
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := model.Spawn(context.Background(), &fsm); err == nil || err.Error() != `not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		t.Run("format", func(t *testing.T) {
			conn := mock.DefaultConn()
			fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := model.Spawn(context.Background(), &fsm, 123, 456); err == nil || err.Error() != `invalid format name argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})

		t.Run("count", func(t *testing.T) {
			conn := mock.DefaultConn()
			fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := model.Spawn(context.Background(), &fsm, "fmt", "xyz"); err == nil || err.Error() != `invalid count argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})
	})
}

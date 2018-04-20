package io_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/io"
	"go.uber.org/zap"
)

func init() {
	if !testing.Verbose() {
		marionette.Logger = zap.NewNop()
	}
}

func TestGets(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		conn := mock.DefaultConn()
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("foo"))
			return 3, nil
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }

		if err := io.Gets(context.Background(), &fsm, "foo"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := io.Gets(context.Background(), &fsm); err == nil || err.Error() != `not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := io.Gets(context.Background(), &fsm, 123); err == nil || err.Error() != `invalid argument type` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	// Ensure read errors from the underlying connection are passed through.
	t.Run("ErrRead", func(t *testing.T) {
		errMarker := errors.New("marker")
		conn := mock.DefaultConn()
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) { return 0, errMarker }
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }

		if err := io.Gets(context.Background(), &fsm, "foo"); err != errMarker {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure unexpected data is returned as an error.
	t.Run("ErrUnexpectedRead", func(t *testing.T) {
		conn := mock.DefaultConn()
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("bar"))
			return 3, nil
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }

		if err := io.Gets(context.Background(), &fsm, "foo"); err == nil || err.Error() != `unexpected data: "bar"` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

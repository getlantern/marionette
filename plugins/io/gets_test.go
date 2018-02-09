package io_test

import (
	"errors"
	"testing"
	"time"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/io"
)

func TestGets(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("foo"))
			return 3, nil
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if ok, err := io.Gets(&fsm, []interface{}{"foo"}); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal("expected success")
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		var conn mock.Conn
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		if _, err := io.Gets(&fsm, nil); err == nil || err.Error() != `io.gets: not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		var conn mock.Conn
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		if _, err := io.Gets(&fsm, []interface{}{123}); err == nil || err.Error() != `io.gets: invalid argument type` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	// Ensure failure returned but not an error if there are not enough bytes.
	t.Run("ErrBufferFull", func(t *testing.T) {
		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("fo"))
			return 2, nil
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if ok, err := io.Gets(&fsm, []interface{}{"foo"}); err != nil {
			t.Fatal(err)
		} else if ok {
			t.Fatal("expected failure")
		}
	})

	// Ensure read errors from the underlying connection are passed through.
	t.Run("ErrRead", func(t *testing.T) {
		errMarker := errors.New("marker")
		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) { return 0, errMarker }
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if _, err := io.Gets(&fsm, []interface{}{"foo"}); err != errMarker {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure unexpected data is returned as an error.
	t.Run("ErrUnexpectedRead", func(t *testing.T) {
		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("bar"))
			return 3, nil
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if _, err := io.Gets(&fsm, []interface{}{"foo"}); err == nil || err.Error() != `io.gets: unexpected data: "bar"` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

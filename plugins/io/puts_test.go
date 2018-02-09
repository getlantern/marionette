package io_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/io"
)

func TestPuts(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var conn mock.Conn
		conn.WriteFn = func(p []byte) (int, error) {
			if string(p) != "foo" {
				t.Fatalf("unexpected write: %q", p)
			}
			copy(p, []byte("foo"))
			return 3, nil
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if ok, err := io.Puts(&fsm, []interface{}{"foo"}); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal("expected success")
		}
	})

	// Ensure writes are continually attempted if there is a timeout error.
	t.Run("Timeout", func(t *testing.T) {
		var i int
		var conn mock.Conn
		conn.WriteFn = func(p []byte) (int, error) {
			defer func() { i++ }()
			switch i {
			case 0:
				return 1, &TimeoutError{}
			case 1:
				return 0, &TimeoutError{}
			case 2:
				return 2, nil
			default:
				return 0, fmt.Errorf("too many writes: %d", i)
			}
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if ok, err := io.Puts(&fsm, []interface{}{"foo"}); err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal("expected success")
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		var conn mock.Conn
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		if _, err := io.Puts(&fsm, nil); err == nil || err.Error() != `io.puts: not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		var conn mock.Conn
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		if _, err := io.Puts(&fsm, []interface{}{123}); err == nil || err.Error() != `io.puts: invalid argument type` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	// Ensure write errors are passed through.
	t.Run("ErrWrite", func(t *testing.T) {
		errMarker := errors.New("marker")
		var conn mock.Conn
		conn.WriteFn = func(p []byte) (int, error) {
			return 0, errMarker
		}
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())

		if _, err := io.Puts(&fsm, []interface{}{"foo"}); err != errMarker {
			t.Fatalf("unexpected error: %q", err)
		}
	})
}

type TimeoutError struct{}

func (e TimeoutError) Error() string { return "timeout" }
func (e TimeoutError) Timeout() bool { return true }

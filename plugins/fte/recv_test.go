package fte_test

import (
	"errors"
	"testing"
	"time"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/fte"
)

func TestRecv(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		streamSet := marionette.NewStreamSet()

		// Create two streams.
		stream1, stream2 := streamSet.Create(), streamSet.Create()
		defer stream1.Close()
		defer stream2.Close()

		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("barbaz"))
			return 6, nil
		}

		fsm := mock.NewFSM(&conn, streamSet)
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.DecryptFn = func(ciphertext []byte) (plaintext, remainder []byte, err error) {
			if string(ciphertext) != `barbaz` {
				t.Fatalf("unexpected ciphertext: %q", ciphertext)
			}

			cell := &marionette.Cell{
				UUID:       100,
				InstanceID: 200,
				StreamID:   stream1.ID(),
				SequenceID: 0,
				Payload:    []byte(`foo`),
			}
			buf, err := cell.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			return buf, []byte("baz"), nil
		}
		fsm.CipherFn = func(regex string) marionette.Cipher {
			if regex != `([a-z0-9]+)` {
				t.Fatalf("unexpected regex: %s", regex)
			}
			return &cipher
		}

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err != nil {
			t.Fatal(err)
		}

		// Read data from stream.
		buf := make([]byte, 3)
		if n, err := stream1.Read(buf); err != nil {
			t.Fatal(err)
		} else if n != 3 {
			t.Fatalf("unexpected n: %d", n)
		} else if string(buf) != `foo` {
			t.Fatalf("unexpected read: %q", buf)
		}
	})

	// Ensure instance ID can be set and retried.
	t.Run("SetInstanceID", func(t *testing.T) {
		var conn mock.Conn
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("bar"))
			return 3, nil
		}

		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 0 }

		var setInstanceIDInvoked bool
		fsm.SetInstanceIDFn = func(id int) {
			setInstanceIDInvoked = true
			if id != 200 {
				t.Fatalf("unexpected id: %d", id)
			}
		}

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.DecryptFn = func(ciphertext []byte) (plaintext, remainder []byte, err error) {
			cell := &marionette.Cell{UUID: 100, InstanceID: 200, StreamID: 300, SequenceID: 0, Payload: []byte(`foo`)}
			buf, err := cell.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			return buf, nil, nil
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err != marionette.ErrRetryTransition {
			t.Fatal(err)
		} else if !setInstanceIDInvoked {
			t.Fatal("expected FSM.SetInstanceID() invocation")
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := fte.Recv(&fsm); err == nil || err.Error() != `fte.recv: not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		t.Run("regex", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := fte.Recv(&fsm, 123, 456); err == nil || err.Error() != `fte.recv: invalid regex argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})

		t.Run("msg_len", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := fte.Recv(&fsm, "abc", "def"); err == nil || err.Error() != `fte.recv: invalid msg_len argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})
	})

	// Ensure plugin passes through connection errors.
	t.Run("ErrConnPeek", func(t *testing.T) {
		errMarker := errors.New("marker")
		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			return 0, errMarker
		}

		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err != errMarker {
			t.Fatal(err)
		}
	})

	// Ensure plugin passes through decryption errors.
	t.Run("ErrDecrypt", func(t *testing.T) {
		errMarker := errors.New("marker")
		var conn mock.Conn
		conn.SetReadDeadlineFn = func(_ time.Time) error { return nil }
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("foo"))
			return 3, nil
		}

		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.DecryptFn = func(ciphertext []byte) (plaintext, remainder []byte, err error) {
			return nil, nil, errMarker
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err != errMarker {
			t.Fatal(err)
		}
	})

	// Ensure an error is returned if the UUID of the FSM and cell do not match.
	t.Run("ErrUUIDMismatch", func(t *testing.T) {
		var conn mock.Conn
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("bar"))
			return 3, nil
		}

		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.DecryptFn = func(ciphertext []byte) (plaintext, remainder []byte, err error) {
			cell := &marionette.Cell{UUID: 400, InstanceID: 200, StreamID: 300, SequenceID: 0, Payload: []byte(`foo`)}
			buf, err := cell.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			return buf, nil, nil
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err == nil || err.Error() != `uuid mismatch: fsm=100, cell=400` {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Ensure an error is returned if the instance ID of the FSM and cell do not match.
	t.Run("ErrInstanceIDMismatch", func(t *testing.T) {
		var conn mock.Conn
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("bar"))
			return 3, nil
		}

		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.DecryptFn = func(ciphertext []byte) (plaintext, remainder []byte, err error) {
			cell := &marionette.Cell{UUID: 100, InstanceID: 400, StreamID: 300, SequenceID: 0, Payload: []byte(`foo`)}
			buf, err := cell.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			return buf, nil, nil
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err == nil || err.Error() != `instance id mismatch: fsm=200, cell=400` {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Ensure an error is returned if a stream has been closed.
	t.Run("ErrStreamClosed", func(t *testing.T) {
		var conn mock.Conn
		conn.ReadFn = func(p []byte) (int, error) {
			copy(p, []byte("bar"))
			return 3, nil
		}

		streamSet := marionette.NewStreamSet()
		stream := streamSet.Create()
		if err := stream.Close(); err != nil {
			t.Fatal(err)
		}

		fsm := mock.NewFSM(&conn, streamSet)
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.DecryptFn = func(ciphertext []byte) (plaintext, remainder []byte, err error) {
			cell := &marionette.Cell{UUID: 100, InstanceID: 200, StreamID: stream.ID(), SequenceID: 0, Payload: []byte(`foo`)}
			buf, err := cell.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			return buf, nil, nil
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		if err := fte.Recv(&fsm, `([a-z0-9]+)`, 128); err != marionette.ErrStreamClosed {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

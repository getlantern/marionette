package fte_test

import (
	"context"
	"errors"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/fte"
)

func TestSend(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		streamSet := marionette.NewStreamSet()

		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, streamSet)
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() int { return 128 }
		cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
			var cell marionette.Cell
			if err := cell.UnmarshalBinary(plaintext); err != nil {
				t.Fatal(err)
			} else if cell.UUID != 100 {
				t.Fatalf("unexpected uuid: %d", cell.UUID)
			} else if cell.InstanceID != 200 {
				t.Fatalf("unexpected instance id: %d", cell.InstanceID)
			} else if string(cell.Payload) != `foo` {
				t.Fatalf("unexpected payload: %s", plaintext)
			}
			return []byte(`bar`), nil
		}
		fsm.CipherFn = func(regex string) (marionette.Cipher, error) {
			if regex != `([a-z0-9]+)` {
				t.Fatalf("unexpected regex: %s", regex)
			}
			return &cipher, nil
		}

		var writeInvoked bool
		conn.WriteFn = func(p []byte) (int, error) {
			writeInvoked = true
			if string(p) != `bar` {
				t.Fatalf("unexpected write: %q", p)
			}
			return 3, nil
		}

		stream := streamSet.Create()
		if _, err := stream.Write([]byte(`foo`)); err != nil {
			t.Fatal(err)
		}

		if err := fte.Send(context.Background(), &fsm, `([a-z0-9]+)`, 128); err != nil {
			t.Fatal(err)
		} else if !writeInvoked {
			t.Fatal("expected conn.Write()")
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := fte.Send(context.Background(), &fsm); err == nil || err.Error() != `not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		t.Run("regex", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := fte.Send(context.Background(), &fsm, 123, 456); err == nil || err.Error() != `invalid regex argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})

		t.Run("msg_len", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := fte.Send(context.Background(), &fsm, "abc", "def"); err == nil || err.Error() != `invalid msg_len argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})
	})

	t.Run("NoData", func(t *testing.T) {
		t.Run("Sync", func(t *testing.T) {
			streamSet := marionette.NewStreamSet()

			conn := mock.DefaultConn()
			fsm := mock.NewFSM(&conn, streamSet)
			fsm.PartyFn = func() string { return marionette.PartyClient }
			fsm.UUIDFn = func() int { return 100 }
			fsm.InstanceIDFn = func() int { return 200 }

			var cipher mock.Cipher
			cipher.CapacityFn = func() int { return 128 }
			cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
				var cell marionette.Cell
				if err := cell.UnmarshalBinary(plaintext); err != nil {
					t.Fatal(err)
				} else if string(cell.Payload) != `foo` {
					t.Fatalf("unexpected payload: %s", plaintext)
				}
				return []byte(`bar`), nil
			}
			fsm.CipherFn = func(regex string) (marionette.Cipher, error) {
				if regex != `([a-z0-9]+)` {
					t.Fatalf("unexpected regex: %s", regex)
				}
				return &cipher, nil
			}

			var writeInvoked bool
			conn.WriteFn = func(p []byte) (int, error) {
				writeInvoked = true
				if string(p) != `bar` {
					t.Fatalf("unexpected write: %q", p)
				}
				return 3, nil
			}

			stream := streamSet.Create()
			if _, err := stream.Write([]byte(`foo`)); err != nil {
				t.Fatal(err)
			}

			if err := fte.Send(context.Background(), &fsm, `([a-z0-9]+)`, 128); err != nil {
				t.Fatal(err)
			} else if !writeInvoked {
				t.Fatal("expected conn.Write()")
			}
		})

		// Ensure that send does not wait for data to become available.
		t.Run("Async", func(t *testing.T) {
			conn := mock.DefaultConn()
			fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			fsm.UUIDFn = func() int { return 100 }
			fsm.InstanceIDFn = func() int { return 200 }

			var cipher mock.Cipher
			cipher.CapacityFn = func() int { return 128 }
			fsm.CipherFn = func(regex string) (marionette.Cipher, error) {
				if regex != `([a-z0-9]+)` {
					t.Fatalf("unexpected regex: %s", regex)
				}
				return &cipher, nil
			}

			conn.WriteFn = func(p []byte) (int, error) {
				t.Fatal("unexpected write")
				return 0, nil
			}

			if err := fte.SendAsync(context.Background(), &fsm, `([a-z0-9]+)`, 128); err != nil {
				t.Fatal(err)
			}
		})
	})

	// Ensure cipher encryption errors are passed through.
	t.Run("ErrCipherEncrypt", func(t *testing.T) {
		errMarker := errors.New("marker")
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() int { return 128 }
		cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
			return nil, errMarker
		}
		fsm.CipherFn = func(regex string) (marionette.Cipher, error) { return &cipher, nil }

		if err := fte.Send(context.Background(), &fsm, `([a-z0-9]+)`, 128); err != errMarker {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure connection write errors are passed through.
	t.Run("ErrConnWrite", func(t *testing.T) {
		errMarker := errors.New("marker")
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() int { return 128 }
		cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
			return []byte(`bar`), nil
		}
		fsm.CipherFn = func(regex string) (marionette.Cipher, error) { return &cipher, nil }

		conn.WriteFn = func(p []byte) (int, error) {
			return 0, errMarker
		}

		if err := fte.Send(context.Background(), &fsm, `([a-z0-9]+)`, 128); err != errMarker {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

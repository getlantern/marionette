package fte_test

import (
	"errors"
	"testing"
	"time"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/fte"
)

func TestSend(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		streamSet := marionette.NewStreamSet()

		var conn mock.Conn
		fsm := mock.NewFSM(&conn, streamSet)
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
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
		fsm.CipherFn = func(regex string) marionette.Cipher {
			if regex != `([a-z0-9]+)` {
				t.Fatalf("unexpected regex: %s", regex)
			}
			return &cipher
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

		if err := fte.Send(&fsm, `([a-z0-9]+)`, 128); err != nil {
			t.Fatal(err)
		} else if !writeInvoked {
			t.Fatal("expected conn.Write()")
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := fte.Send(&fsm); err == nil || err.Error() != `fte.send: not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		t.Run("regex", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := fte.Send(&fsm, 123, 456); err == nil || err.Error() != `fte.send: invalid regex argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})

		t.Run("msg_len", func(t *testing.T) {
			fsm := mock.NewFSM(&mock.Conn{}, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			if err := fte.Send(&fsm, "abc", "def"); err == nil || err.Error() != `fte.send: invalid msg_len argument type` {
				t.Fatalf("unexpected error: %q", err)
			}
		})
	})

	t.Run("NoData", func(t *testing.T) {
		// Ensure that send waits for data to become available before sending.
		t.Run("Sync", func(t *testing.T) {
			streamSet := marionette.NewStreamSet()

			var conn mock.Conn
			fsm := mock.NewFSM(&conn, streamSet)
			fsm.PartyFn = func() string { return marionette.PartyClient }
			fsm.UUIDFn = func() int { return 100 }
			fsm.InstanceIDFn = func() int { return 200 }

			var cipher mock.Cipher
			cipher.CapacityFn = func() (int, error) { return 128, nil }
			cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
				var cell marionette.Cell
				if err := cell.UnmarshalBinary(plaintext); err != nil {
					t.Fatal(err)
				} else if string(cell.Payload) != `foo` {
					t.Fatalf("unexpected payload: %s", plaintext)
				}
				return []byte(`bar`), nil
			}
			fsm.CipherFn = func(regex string) marionette.Cipher {
				if regex != `([a-z0-9]+)` {
					t.Fatalf("unexpected regex: %s", regex)
				}
				return &cipher
			}

			var writeInvoked bool
			conn.WriteFn = func(p []byte) (int, error) {
				writeInvoked = true
				if string(p) != `bar` {
					t.Fatalf("unexpected write: %q", p)
				}
				return 3, nil
			}

			// Delay write momentarily.
			go func() {
				time.Sleep(100 * time.Millisecond)
				stream := streamSet.Create()
				if _, err := stream.Write([]byte(`foo`)); err != nil {
					t.Fatal(err)
				}
			}()

			if err := fte.Send(&fsm, `([a-z0-9]+)`, 128); err != nil {
				t.Fatal(err)
			} else if !writeInvoked {
				t.Fatal("expected conn.Write()")
			}
		})

		// Ensure that send does not wait for data to become available.
		t.Run("Async", func(t *testing.T) {
			var conn mock.Conn
			fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
			fsm.PartyFn = func() string { return marionette.PartyClient }
			fsm.UUIDFn = func() int { return 100 }
			fsm.InstanceIDFn = func() int { return 200 }

			var cipher mock.Cipher
			cipher.CapacityFn = func() (int, error) { return 128, nil }
			cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
				var cell marionette.Cell
				if err := cell.UnmarshalBinary(plaintext); err != nil {
					t.Fatal(err)
				} else if string(cell.Payload) != `` {
					t.Fatalf("unexpected payload: %s", plaintext)
				}
				return []byte(`bar`), nil
			}
			fsm.CipherFn = func(regex string) marionette.Cipher {
				if regex != `([a-z0-9]+)` {
					t.Fatalf("unexpected regex: %s", regex)
				}
				return &cipher
			}

			var writeInvoked bool
			conn.WriteFn = func(p []byte) (int, error) {
				writeInvoked = true
				if string(p) != `bar` {
					t.Fatalf("unexpected write: %q", p)
				}
				return 3, nil
			}

			if err := fte.SendAsync(&fsm, `([a-z0-9]+)`, 128); err != nil {
				t.Fatal(err)
			} else if !writeInvoked {
				t.Fatal("expected conn.Write()")
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
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
			return nil, errMarker
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		if err := fte.SendAsync(&fsm, `([a-z0-9]+)`, 128); err != errMarker {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure connection write errors are passed through.
	t.Run("ErrConnWrite", func(t *testing.T) {
		errMarker := errors.New("marker")
		var conn mock.Conn
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.UUIDFn = func() int { return 100 }
		fsm.InstanceIDFn = func() int { return 200 }

		var cipher mock.Cipher
		cipher.CapacityFn = func() (int, error) { return 128, nil }
		cipher.EncryptFn = func(plaintext []byte) ([]byte, error) {
			return []byte(`bar`), nil
		}
		fsm.CipherFn = func(regex string) marionette.Cipher { return &cipher }

		conn.WriteFn = func(p []byte) (int, error) {
			return 0, errMarker
		}

		if err := fte.SendAsync(&fsm, `([a-z0-9]+)`, 128); err != errMarker {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

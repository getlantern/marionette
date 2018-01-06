package marionette_test

/*
import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette"
)

func TestStreamBufferSet(t *testing.T) {
	t.Run("SingleStream", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.NORMAL, Payload: []byte("foo"), Length: 28, StreamID: 100, SequenceID: 1, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Padded", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		if cell := s.Pop(2, 3, 10); cell.Length != marionette.CellHeaderSize+10 {
			t.Fatalf("unexpected cell length: %d", cell.Length)
		}
	})

	t.Run("AboveMaxCellLength", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		if cell := s.Pop(2, 3, marionette.MaxCellLength+1); cell.Length != marionette.MaxCellLength {
			t.Fatalf("unexpected cell length: %d", cell.Length)
		}
	})

	t.Run("MultiPush", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		s.Push(100, []byte("bar"))
		s.Push(100, []byte("baz"))
		if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.NORMAL, Payload: []byte("foobarbaz"), Length: 28, StreamID: 100, SequenceID: 1, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("MultiStream", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		s.Push(101, []byte("bar"))
		s.Push(100, []byte("baz"))
		if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.NORMAL, Payload: []byte("foobaz"), Length: 31, StreamID: 100, SequenceID: 1, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		} else if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.NORMAL, Payload: []byte("bar"), Length: 28, StreamID: 101, SequenceID: 1, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("NoDataAvailable", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.NORMAL, Payload: []byte("foo"), Length: 28, StreamID: 100, SequenceID: 1, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		} else if diff := cmp.Diff(s.Pop(2, 3, 0), (*marionette.Cell)(nil)); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("Terminated", func(t *testing.T) {
		s := NewStreamBufferSet()
		s.Push(100, []byte("foo"))
		s.Terminate(100)
		if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.NORMAL, Payload: []byte("foo"), Length: 28, StreamID: 100, SequenceID: 1, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		} else if diff := cmp.Diff(s.Pop(2, 3, 0), &marionette.Cell{Type: marionette.END_OF_STREAM, Length: 25, StreamID: 100, SequenceID: 2, UUID: 2, InstanceID: 3}); diff != "" {
			t.Fatal(diff)
		} else if diff := cmp.Diff(s.Pop(2, 3, 0), (*marionette.Cell)(nil)); diff != "" {
			t.Fatal(diff)
		}
	})
}

// NewStreamBufferSet returns a testable StreamBufferSet.
func NewStreamBufferSet() *marionette.StreamBufferSet {
	s := marionette.NewStreamBufferSet()
	s.Rand = NewRand()
	return s
}
*/

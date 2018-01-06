package marionette_test

/*
import (
	"context"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
)

func TestFSM(t *testing.T) {
	doc := mar.MustParse([]byte(`
connection(tcp, 8082):
  start      handshake  NULL               1.0
  handshake  upstream   upstream_handshake 1.0
  upstream   downstream upstream_async     1.0
  downstream upstream   downstream_async   1.0

action upstream_handshake:
  client fte.send("^.*$", 128)

action upstream_async:
  client fte.send_async("^.*$", 128)

action downstream_async:
  server fte.send_async("^.*$", 128)
`))
	fsm := NewFSM(doc, marionette.PartyClient)

	// Verify the FSM begins in the "start" state.
	if state := fsm.State(); state != `start` {
		t.Fatalf("unexpected initial state: %s", state)
	}

	// There is no action block so the state should move to handshake.
	if err := fsm.Next(context.Background()); err != nil {
		t.Fatal(err)
	} else if state := fsm.State(); state != "handshake" {
		t.Fatalf("unexpected state: %s", state)
	}

	// It should not move past handshake until there is a matching buffer.
	if err := fsm.Next(context.Background()); err != marionette.ErrNoTransition {
		t.Fatal(err)
	} else if state := fsm.State(); state != "handshake" {
		t.Fatalf("unexpected state: %s", state)
	}

	// Add data to the buffer set and retry.
	fsm.StreamBufferSet.Push(100, []byte("foo"))
	if err := fsm.Next(context.Background()); err != nil {
		t.Fatal(err)
	} else if state := fsm.State(); state != "upstream" {
		t.Fatalf("unexpected state: %s", state)
	}
}

type FSM struct {
	*marionette.FSM

	Conn BufferConn

	Streams *marionette.StreamBufferSet

	Buffer      bytes.Buffer
	CellDecoder *marionette.CellDecoder
}

func NewFSM(doc *mar.Document, party string) *FSM {
	var fsm FSM
	fsm.StreamBufferSet = marionette.NewStreamBufferSet()
	fsm.CellDecoder = marionette.NewCellDecoder(&fsm.Buffer)
	fsm.FSM = marionette.NewFSM(doc, party, fsm.StreamBufferSet, fsm.CellDecoder)
	fsm.FSM.SetConn(&fsm.Conn)
	fsm.FSM.SetLogger(NewLogger())
	return &fsm
}
*/

package marionette_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	"github.com/redjack/marionette/mock"
)

func TestFSM_Dummy(t *testing.T) {
	doc := mar.MustParse(mar.Format("dummy", ""))
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

}

type FSM struct {
	*marionette.FSM

	Conn mock.Conn

	StreamBufferSet *marionette.StreamBufferSet

	Buffer      bytes.Buffer
	CellDecoder *marionette.CellDecoder
}

func NewFSM(doc *mar.Document, party string) *FSM {
	var fsm FSM
	fsm.StreamBufferSet = marionette.NewStreamBufferSet()
	fsm.CellDecoder = marionette.NewCellDecoder(&fsm.Buffer)
	fsm.FSM = marionette.NewFSM(doc, party, fsm.StreamBufferSet, fsm.CellDecoder)
	fsm.FSM.SetConn(&fsm.Conn)
	return &fsm
}

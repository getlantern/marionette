package marionette

import (
	"net"

	"github.com/redjack/marionette/mar"
)

// Dialer represents a client-side dialer that communicates over the marionette protocol.
type Dialer struct {
	fsm *FSM
}

// NewDialer returns a new instance of Dialer.
func NewDialer(doc *mar.Document) *Dialer {
	return &Dialer{
		fsm: NewFSM(doc, PartyClient, NewStreamSet()),
	}
}

// Dial returns a new stream from the dialer.
func (d *Dialer) Dial() (net.Conn, error) {
	return d.fsm.streams.Create(), nil
}

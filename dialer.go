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
func NewDialer(doc *mar.Document, addr string) (*Dialer, error) {
	conn, err := net.Dial(doc.Transport, net.JoinHostPort(addr, doc.Port))
	if err != nil {
		return nil, err
	}

	fsm := NewFSM(doc, PartyClient, NewStreamSet())
	fsm.conn = conn
	return &Dialer{fsm: fsm}, nil
}

// Close stops the dialer and its underlying connections.
func (d *Dialer) Close() error {
	return d.fsm.Close()
}

// Dial returns a new stream from the dialer.
func (d *Dialer) Dial() (net.Conn, error) {
	return d.fsm.streams.Create(), nil
}

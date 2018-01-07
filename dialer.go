package marionette

import (
	"context"
	"net"
	"sync"

	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

// Dialer represents a client-side dialer that communicates over the marionette protocol.
type Dialer struct {
	fsm *FSM
	wg  sync.WaitGroup
}

// NewDialer returns a new instance of Dialer.
func NewDialer(doc *mar.Document, addr string) (*Dialer, error) {
	conn, err := net.Dial(doc.Transport, net.JoinHostPort(addr, doc.Port))
	if err != nil {
		return nil, err
	}

	fsm := NewFSM(doc, PartyClient)
	fsm.conn = conn
	fsm.streams.LocalAddr = conn.LocalAddr()
	fsm.streams.RemoteAddr = conn.RemoteAddr()

	// Run execution in a separate goroutine.
	d := &Dialer{fsm: fsm}
	d.wg.Add(1)
	go func() { defer d.wg.Done(); d.execute(context.Background()) }()
	return d, nil
}

// Close stops the dialer and its underlying connections.
func (d *Dialer) Close() error {
	err := d.fsm.Close()
	d.wg.Wait()
	return err
}

// Dial returns a new stream from the dialer.
func (d *Dialer) Dial() (net.Conn, error) {
	return d.fsm.streams.Create(), nil
}

func (d *Dialer) execute(ctx context.Context) {
	Logger.Debug("client fsm executing")
	defer Logger.Debug("client execution complete")

	if err := d.fsm.Execute(ctx); err != nil {
		Logger.Debug("client execution error", zap.Error(err))
	}
}

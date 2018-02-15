package marionette

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

var (
	// ErrDialerClosed is returned when trying to operate on a closed dialer.
	ErrDialerClosed = errors.New("marionette: dialer closed")
)

// Dialer represents a client-side dialer that communicates over the marionette protocol.
type Dialer struct {
	mu        sync.RWMutex
	fsm       FSM
	streamSet *StreamSet

	ctx    context.Context
	cancel func()

	closed bool
	wg     sync.WaitGroup
}

// NewDialer returns a new instance of Dialer.
func NewDialer(doc *mar.Document, addr string, streamSet *StreamSet) (*Dialer, error) {
	conn, err := net.Dial(doc.Transport, net.JoinHostPort(addr, doc.Port))
	if err != nil {
		return nil, err
	}

	// Run execution in a separate goroutine.
	d := &Dialer{
		fsm:       NewFSM(doc, addr, PartyClient, conn, streamSet),
		streamSet: streamSet,
	}
	d.ctx, d.cancel = context.WithCancel(context.Background())

	d.wg.Add(1)
	go func() { defer d.wg.Done(); d.execute() }()
	return d, nil
}

// Close stops the dialer and its underlying connections.
func (d *Dialer) Close() (err error) {
	d.mu.Lock()
	d.closed = true
	err = d.fsm.Conn().Close()
	d.mu.Unlock()

	d.cancel()

	d.wg.Wait()
	return err
}

// Closed returns true if the dialer has been closed.
func (d *Dialer) Closed() bool {
	d.mu.RLock()
	closed := d.closed
	d.mu.RUnlock()
	return closed
}

// Dial returns a new stream from the dialer.
func (d *Dialer) Dial() (net.Conn, error) {
	if d.Closed() {
		return nil, ErrDialerClosed
	}
	return d.streamSet.Create(), nil
}

func (d *Dialer) execute() {
	defer d.Close()

	for !d.Closed() {
		if err := d.fsm.Execute(d.ctx); err != nil {
			Logger.Debug("dialer error", zap.Error(err))
			return
		}
		d.fsm.Reset()
	}
}

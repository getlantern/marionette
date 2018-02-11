package marionette

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

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
		fsm:       NewFSM(doc, "", PartyClient, conn, streamSet),
		streamSet: streamSet,
	}
	d.wg.Add(1)
	go func() { defer d.wg.Done(); d.execute(context.Background()) }()
	return d, nil
}

// Close stops the dialer and its underlying connections.
func (d *Dialer) Close() (err error) {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()

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

func (d *Dialer) execute(ctx context.Context) {
	Logger.Debug("client fsm executing")
	defer Logger.Debug("client execution complete")

	for !d.Closed() {
		if err := d.fsm.Execute(ctx); err != nil {
			if !d.Closed() {
				Logger.Debug("client execution error", zap.Error(err))
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

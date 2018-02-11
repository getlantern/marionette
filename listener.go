package marionette

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

var (
	// ErrListenerClosed is returned when trying to operate on a closed listener.
	ErrListenerClosed = errors.New("marionette: listener closed")
)

// Listener listens on a port and communicates over the marionette protocol.
type Listener struct {
	mu         sync.RWMutex
	iface      string
	ln         net.Listener
	conns      map[net.Conn]struct{}
	doc        *mar.Document
	newStreams chan *Stream
	err        error

	once    sync.Once
	wg      sync.WaitGroup
	closing chan struct{}
	closed  bool
}

// Listen returns a new instance of Listener.
func Listen(doc *mar.Document, iface string) (*Listener, error) {
	// Parse port from MAR specification.
	// TODO: Handle "ftp_pasv_port".
	port, err := strconv.Atoi(doc.Port)
	if err != nil {
		return nil, errors.New("invalid connection port")
	}
	addr := net.JoinHostPort(iface, strconv.Itoa(port))

	Logger.Debug("opening listener", zap.String("transport", doc.Transport), zap.String("bind", addr))

	ln, err := net.Listen(doc.Transport, addr)
	if err != nil {
		return nil, err
	}
	l := &Listener{
		ln:         ln,
		iface:      iface,
		doc:        doc,
		conns:      make(map[net.Conn]struct{}),
		newStreams: make(chan *Stream),
		closing:    make(chan struct{}),
	}

	// Hand off connection handling to separate goroutine.
	l.wg.Add(1)
	go func() { defer l.wg.Done(); l.accept() }()

	return l, nil
}

// Err returns the last error that occurred on the listener.
func (l *Listener) Err() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.err
}

// Addr returns the underlying network address.
func (l *Listener) Addr() net.Addr { return l.ln.Addr() }

// Close stops the listener and waits for the connections to finish.
func (l *Listener) Close() error {
	err := l.ln.Close()

	l.mu.Lock()
	l.closed = true
	for conn := range l.conns {
		if e := conn.Close(); e != nil && err == nil {
			err = e
		}
		delete(l.conns, conn)
	}
	l.mu.Unlock()

	l.once.Do(func() { close(l.closing) })
	l.wg.Wait()

	return err
}

// Closed returns true if the listener has been closed.
func (l *Listener) Closed() bool {
	l.mu.RLock()
	closed := l.closed
	l.mu.RUnlock()
	return closed
}

// Accept waits for a new connection.
func (l *Listener) Accept() (net.Conn, error) {
	select {
	case <-l.closing:
		return nil, ErrListenerClosed
	case stream := <-l.newStreams:
		return stream, l.Err()
	}
}

// accept continually accepts networks connections and multiplexes to streams.
func (l *Listener) accept() {
	defer close(l.newStreams)

	for {
		// Wait for next connection.
		conn, err := l.ln.Accept()
		if err != nil {
			l.mu.Lock()
			if l.closed {
				l.err = ErrListenerClosed
			} else {
				l.err = err
			}
			l.mu.Unlock()
			return
		}

		streamSet := NewStreamSet()
		streamSet.OnNewStream = l.onNewStream

		fsm := NewFSM(l.doc, l.iface, PartyServer, conn, streamSet)

		// Run execution in a separate goroutine.
		l.wg.Add(1)
		go func() { defer l.wg.Done(); l.execute(context.Background(), fsm, conn) }()
	}
}

func (l *Listener) execute(ctx context.Context, fsm FSM, conn net.Conn) {
	Logger.Debug("server fsm executing")
	defer Logger.Debug("server fsm execution complete")

	l.addConn(conn)
	defer l.removeConn(conn)

	for !l.Closed() {
		if err := fsm.Execute(ctx); err != nil {
			if !l.Closed() {
				Logger.Debug("server fsm execution error", zap.Error(err))
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// onNewStream is called everytime the FSM's stream set creates a new stream.
func (l *Listener) onNewStream(stream *Stream) {
	Logger.Debug("new server stream")
	l.newStreams <- stream
}

func (l *Listener) addConn(conn net.Conn) {
	l.mu.Lock()
	l.conns[conn] = struct{}{}
	l.mu.Unlock()
}

func (l *Listener) removeConn(conn net.Conn) {
	l.mu.Lock()
	delete(l.conns, conn)
	l.mu.Unlock()
}

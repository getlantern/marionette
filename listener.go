package marionette

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"

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

	ctx    context.Context
	cancel func()

	once    sync.Once
	wg      sync.WaitGroup
	closing chan struct{}
	closed  bool

	// Specifies directory for dumping stream traces. Passed to StreamSet.TracePath.
	TracePath string
}

// Listen returns a new instance of Listener.
func Listen(doc *mar.Document, iface string) (*Listener, error) {
	// Parse port from MAR specification.
	port, err := strconv.Atoi(doc.Port)
	if err != nil {
		return nil, errors.New("invalid connection port")
	}
	addr := net.JoinHostPort(iface, strconv.Itoa(port))

	Logger.Debug("listen", zap.String("transport", doc.Transport), zap.String("bind", addr))

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
	l.ctx, l.cancel = context.WithCancel(context.Background())

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

	l.once.Do(func() {
		l.cancel()
		close(l.closing)
	})
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
		streamSet.TracePath = l.TracePath

		fsm := NewFSM(l.doc, l.iface, PartyServer, conn, streamSet)

		// Run execution in a separate goroutine.
		l.wg.Add(1)
		go func() { defer l.wg.Done(); l.execute(fsm, conn) }()
	}
}

func (l *Listener) execute(fsm FSM, conn net.Conn) {
	l.addConn(conn)
	defer l.removeConn(conn)

	for !l.Closed() {
		if err := fsm.Execute(l.ctx); err == ErrStreamClosed {
			return
		} else if err == io.EOF {
			Logger.Debug("client disconnected", zap.String("addr", conn.RemoteAddr().String()))
			return
		} else if err != nil {
			Logger.Debug("server fsm execution error", zap.Error(err))
			return
		}
		fsm.Reset()
	}
}

// onNewStream is called everytime the FSM's stream set creates a new stream.
func (l *Listener) onNewStream(stream *Stream) {
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

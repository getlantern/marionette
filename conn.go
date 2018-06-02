package marionette

import (
	"io"
	"net"
	"strings"
	"sync"
)

type BufferedConn struct {
	net.Conn

	mu  sync.RWMutex
	buf []byte
	err error

	closing chan struct{}
	once    sync.Once

	seekNotify  chan struct{} // sent when seeking forward
	writeNotify chan struct{} // sent when data has been written to the buffer.
}

func NewBufferedConn(conn net.Conn, bufferSize int) *BufferedConn {
	c := &BufferedConn{
		Conn:    conn,
		buf:     make([]byte, 0, bufferSize),
		closing: make(chan struct{}, 0),

		seekNotify:  make(chan struct{}, 1),
		writeNotify: make(chan struct{}, 1),
	}
	go c.monitor()
	return c
}

// Close closes the connection.
func (conn *BufferedConn) Close() error {
	conn.once.Do(func() { close(conn.closing) })
	return conn.Conn.Close()
}

// Append adds b to the end of the buffer.
func (conn *BufferedConn) Append(b []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	copy(conn.buf[len(conn.buf):len(conn.buf)+len(b)], b)
	conn.buf = conn.buf[:len(conn.buf)+len(b)]
}

// Read is unavailable for BufferedConn.
func (conn *BufferedConn) Read(p []byte) (int, error) {
	panic("BufferedConn.Read(): unavailable, use Peek/Seek")
}

// Peek returns the first n bytes of the read buffer.
// If n is -1 then returns any available data after attempting a read.
func (conn *BufferedConn) Peek(n int, blocking bool) ([]byte, error) {
	for {
		// Read buffer & error from monitor under read lock.
		conn.mu.RLock()
		buf, err := conn.buf, conn.err
		conn.mu.RUnlock()

		// Return any data that exists in the buffer.
		switch n {
		case -1:
			return buf, err
		default:
			if n <= len(buf) {
				return buf[:n], nil
			} else if isEOFError(err) {
				return buf, io.EOF
			} else if err != nil {
				return buf, err
			}
		}

		// Exit immediately if we are not blocking.
		if !blocking {
			return buf, err
		}

		// Wait for a new write or error from the monitor.
		<-conn.writeNotify
	}
}

// Seek moves the buffer forward a given number of bytes.
// This implementation only supports io.SeekCurrent.
func (conn *BufferedConn) Seek(offset int64, whence int) (int64, error) {
	assert(whence == io.SeekCurrent)
	assert(offset <= int64(len(conn.buf)))

	conn.mu.Lock()
	defer conn.mu.Unlock()

	b := conn.buf[offset:]
	conn.buf = conn.buf[:len(b)]
	copy(conn.buf, b)

	conn.notifySeek()

	return 0, nil
}

// monitor runs in a separate goroutine and continually reads to the buffer.
func (conn *BufferedConn) monitor() {
	buf := make([]byte, cap(conn.buf))
	for {
		// Ensure connection is not closed.
		select {
		case <-conn.closing:
			return
		default:
		}

		// Determine remaining space on buffer.
		// If no capacity remains then wait for seek or connection close.
		capacity := cap(conn.buf) - len(conn.buf)
		if capacity == 0 {
			select {
			case <-conn.closing:
				return
			case <-conn.seekNotify:
				continue
			}
		}

		// Attempt to read next bytes from connection.
		n, err := conn.Conn.Read(buf[:capacity])

		// Append bytes to connection buffer.
		if n > 0 {
			conn.Append(buf[:n])
			conn.notifyWrite()
		}

		// If an error occurred then save on connection and exit.
		if err != nil && !isTemporaryError(err) {
			conn.err = err
			conn.notifyWrite()
			return
		}
	}
}

func (conn *BufferedConn) notifySeek() {
	select {
	case conn.seekNotify <- struct{}{}:
	default:
	}
}

func (conn *BufferedConn) notifyWrite() {
	select {
	case conn.writeNotify <- struct{}{}:
	default:
	}
}

// isTimeoutError returns true if the error is a timeout error.
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	} else if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
		return true
	}
	return false
}

// isTemporaryError returns true if the error is a temporary error.
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	} else if err, ok := err.(interface{ Temporary() bool }); ok && err.Temporary() {
		return true
	}
	return false
}

func isEOFError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "connection reset by peer")
}

package marionette

import (
	"bufio"
	"io"
	"net"
	"time"
)

type BufferedConn struct {
	net.Conn
	buf []byte
}

func NewBufferedConn(conn net.Conn, bufferSize int) *BufferedConn {
	return &BufferedConn{
		Conn: conn,
		buf:  make([]byte, 0, bufferSize),
	}
}

// Read is unavailable for BufferedConn.
func (conn *BufferedConn) Read(p []byte) (int, error) {
	panic("BufferedConn.Read(): unavailable, use Peek/Seek")
}

// Peek returns the current read buffer.
func (conn *BufferedConn) Peek(n int) ([]byte, error) {
	// Read into buffer if it isn't full yet.
	if len(conn.buf) != cap(conn.buf) {
		// Limit deadline to only read what is available.
		if err := conn.SetReadDeadline(time.Now().Add(1 * time.Microsecond)); err != nil {
			return nil, err
		}

		// Read onto the end of the buffer.
		n, err := conn.Conn.Read(conn.buf[len(conn.buf):cap(conn.buf)])
		if err != nil {
			if isTimeoutError(err) {
				return conn.buf, nil
			}
			return conn.buf, err
		}
		conn.buf = conn.buf[:len(conn.buf)+int(n)]
	}

	if n == -1 {
		return conn.buf, nil
	} else if len(conn.buf) < n {
		return conn.buf, bufio.ErrBufferFull
	}
	return conn.buf[:n], nil
}

// Seek moves the buffer forward a given number of bytes.
// This implementation only supports io.SeekCurrent.
func (conn *BufferedConn) Seek(offset int64, whence int) (int64, error) {
	assert(whence == io.SeekCurrent)
	assert(offset <= int64(len(conn.buf)))

	b := conn.buf[offset:]
	conn.buf = conn.buf[:len(b)]
	copy(conn.buf, b)
	return 0, nil
}

// isTimeoutError returns true if the error is a timeout error.
func isTimeoutError(err error) bool {
	if err, ok := err.(interface {
		Timeout() bool
	}); ok && err.Timeout() {
		return true
	}
	return false
}

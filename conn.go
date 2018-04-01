package marionette

import (
	"io"
	"net"
	"strings"
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

// Peek returns the first n bytes of the read buffer.
// If n is -1 then returns any available data.
func (conn *BufferedConn) Peek(n int, blocking bool) ([]byte, error) {
	for {
		if n >= 0 && len(conn.buf) >= n {
			return conn.buf[:n], nil
		} else if n == -1 && len(conn.buf) > 0 {
			return conn.buf, nil
		}

		capacity := cap(conn.buf)
		if n >= 0 {
			capacity = n - len(conn.buf)
		}

		deadline := time.Now()
		if blocking {
			deadline = deadline.Add(24 * time.Hour)
		} else {
			deadline = deadline.Add(100 * time.Microsecond)
		}

		if err := conn.Conn.SetReadDeadline(deadline); err != nil {
			return conn.buf, err
		}

		nn, err := conn.Conn.Read(conn.buf[len(conn.buf) : len(conn.buf)+capacity])
		if isTimeoutError(err) {
			// nop
		} else if isEOFError(err) {
			return conn.buf, io.EOF
		} else if err != nil {
			return conn.buf, err
		}
		conn.buf = conn.buf[:len(conn.buf)+nn]

		if n == -1 {
			return conn.buf, nil
		}
	}
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
	if err == nil {
		return false
	} else if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
		return true
	}
	return false
}

func isEOFError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "connection reset by peer")
}

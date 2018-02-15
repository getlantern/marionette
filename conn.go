package marionette

import (
	"io"
	"net"
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
// If n is -1 then returns available buffer once any bytes are available.
func (conn *BufferedConn) Peek(n int) ([]byte, error) {
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

		nn, err := conn.Conn.Read(conn.buf[len(conn.buf) : len(conn.buf)+capacity])
		if isTimeoutError(err) {
			continue
		} else if err != nil {
			return conn.buf, err
		}
		conn.buf = conn.buf[:len(conn.buf)+nn]
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

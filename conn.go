package marionette

import (
	"io"
	"net"
	"strings"
)

type BufferedConn struct {
	net.Conn
	buf  []byte
	msgs chan bufferedConnMsg
}

func NewBufferedConn(conn net.Conn, bufferSize int) *BufferedConn {
	bufConn := &BufferedConn{
		Conn: conn,
		buf:  make([]byte, 0, bufferSize+bufferedConnMsgSize),
		msgs: make(chan bufferedConnMsg),
	}
	go bufConn.readBuffer()
	return bufConn
}

func (conn *BufferedConn) readBuffer() {
	for {
		buf := make([]byte, bufferedConnMsgSize)
		n, err := conn.Conn.Read(buf)
		conn.msgs <- bufferedConnMsg{buf[:n], err}
		if isEOFError(err) {
			return
		}
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

		// Wait for next message if blocking. Otherwise skip if no message available.
		var msg bufferedConnMsg
		if blocking {
			msg = <-conn.msgs
		} else {
			select {
			case msg = <-conn.msgs:
			default:
			}
		}

		// Append any buffer returned.
		conn.buf = append(conn.buf, msg.buf...)

		// Handle errors.
		if isEOFError(msg.err) {
			return conn.buf, io.EOF
		} else if msg.err != nil {
			return conn.buf, msg.err
		}

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

const bufferedConnMsgSize = 4096

type bufferedConnMsg struct {
	buf []byte
	err error
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

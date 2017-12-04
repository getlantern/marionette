package marionette

import (
	"net"
	"strings"
	"sync"
)

const (
	PartyClient = "client"
	PartyServer = "server"
)

// StripFormatVersion removes any version specified on a format.
func StripFormatVersion(format string) string {
	if i := strings.Index(format, ":"); i != -1 {
		return format[:i]
	}
	return format
}

type bufConn struct {
	net.Conn

	mu  sync.RWMutex
	buf []byte
}

func newBufConn(conn net.Conn) *bufConn {
	c := &bufConn{Conn: conn}
	// TODO: Start goroutine to read into buffer.
	return c
}

// Peek returns the current buffer.
func (conn *bufConn) Peek() []byte {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	return conn.buf
}

// Read reads
func (conn *bufConn) Read(b []byte) (n int, err error) {
	// TODO: Copy buffer into b.
	// TODO: Shift bytes to beginning.
	// TODO: Return byte count.
	panic("TODO")
}

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

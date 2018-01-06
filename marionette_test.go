package marionette_test

import (
	"bytes"
	"math/rand"
	"net"
	"time"

	"github.com/redjack/marionette"
)

func init() {
	// Ensure all PRNGs are consistent for tests.
	marionette.Rand = NewRand
}

// NewRand returns a PRNG with a zero source.
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

var _ net.Conn = &BufferConn{}

// BufferConn represents a read and write buffer and implements net.Conn.
type BufferConn struct {
	Reader bytes.Buffer
	Writer bytes.Buffer
}

func (c *BufferConn) Read(b []byte) (n int, err error)   { return c.Writer.Read(b) }
func (c *BufferConn) Write(b []byte) (n int, err error)  { return c.Reader.Write(b) }
func (c *BufferConn) Close() error                       { return nil }
func (c *BufferConn) LocalAddr() net.Addr                { return nil }
func (c *BufferConn) RemoteAddr() net.Addr               { return nil }
func (c *BufferConn) SetDeadline(t time.Time) error      { return nil }
func (c *BufferConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *BufferConn) SetWriteDeadline(t time.Time) error { return nil }

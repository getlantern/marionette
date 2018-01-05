package mock

import (
	"net"
	"time"
)

// Ensure mock implements interface.
var _ net.Conn = &Conn{}

type Conn struct {
	ReadFunc             func(b []byte) (n int, err error)
	WriteFunc            func(b []byte) (n int, err error)
	CloseFunc            func() error
	LocalAddrFunc        func() net.Addr
	RemoteAddrFunc       func() net.Addr
	SetDeadlineFunc      func(t time.Time) error
	SetReadDeadlineFunc  func(t time.Time) error
	SetWriteDeadlineFunc func(t time.Time) error
}

func (c *Conn) Read(b []byte) (n int, err error)   { return c.ReadFunc(b) }
func (c *Conn) Write(b []byte) (n int, err error)  { return c.WriteFunc(b) }
func (c *Conn) Close() error                       { return c.CloseFunc() }
func (c *Conn) LocalAddr() net.Addr                { return c.LocalAddrFunc() }
func (c *Conn) RemoteAddr() net.Addr               { return c.RemoteAddrFunc() }
func (c *Conn) SetDeadline(t time.Time) error      { return c.SetDeadlineFunc(t) }
func (c *Conn) SetReadDeadline(t time.Time) error  { return c.SetReadDeadlineFunc(t) }
func (c *Conn) SetWriteDeadline(t time.Time) error { return c.SetWriteDeadlineFunc(t) }

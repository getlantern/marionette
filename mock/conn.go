package mock

import (
	"io"
	"net"
	"time"
)

var _ net.Conn = &Conn{}

type Conn struct {
	ReadFn             func(b []byte) (n int, err error)
	WriteFn            func(b []byte) (n int, err error)
	CloseFn            func() error
	LocalAddrFn        func() net.Addr
	RemoteAddrFn       func() net.Addr
	SetDeadlineFn      func(t time.Time) error
	SetReadDeadlineFn  func(t time.Time) error
	SetWriteDeadlineFn func(t time.Time) error
}

func DefaultConn() Conn {
	return Conn{
		ReadFn:             func(b []byte) (n int, err error) { return 0, io.EOF },
		SetDeadlineFn:      func(t time.Time) error { return nil },
		SetReadDeadlineFn:  func(t time.Time) error { return nil },
		SetWriteDeadlineFn: func(t time.Time) error { return nil },
	}
}

func (c *Conn) Read(b []byte) (n int, err error)   { return c.ReadFn(b) }
func (c *Conn) Write(b []byte) (n int, err error)  { return c.WriteFn(b) }
func (c *Conn) Close() error                       { return c.CloseFn() }
func (c *Conn) LocalAddr() net.Addr                { return c.LocalAddrFn() }
func (c *Conn) RemoteAddr() net.Addr               { return c.RemoteAddrFn() }
func (c *Conn) SetDeadline(t time.Time) error      { return c.SetDeadlineFn(t) }
func (c *Conn) SetReadDeadline(t time.Time) error  { return c.SetReadDeadlineFn(t) }
func (c *Conn) SetWriteDeadline(t time.Time) error { return c.SetWriteDeadlineFn(t) }

package marionette

import (
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
)

// Proxy represents a proxy between a marionette listener and another server.
type Proxy struct {
	ln   *Listener
	addr string
	wg   sync.WaitGroup
}

// NewProxy returns a new instance of Proxy.
func NewProxy(ln *Listener, addr string) *Proxy {
	return &Proxy{
		ln:   ln,
		addr: addr,
	}
}

func (p *Proxy) Open() error {
	p.wg.Add(1)
	go func() { defer p.wg.Done(); p.run() }()

	return nil
}

func (p *Proxy) Close() error {
	return nil
}

func (p *Proxy) run() {
	Logger.Debug("proxy: listening")
	defer Logger.Debug("proxy: closed")

	for {
		conn, err := p.ln.Accept()
		if err != nil {
			Logger.Debug("proxy: listener error", zap.Error(err))
			return
		}

		p.wg.Add(1)
		go func() { defer p.wg.Done(); p.handleConn(conn) }()
	}
}

func (p *Proxy) handleConn(conn net.Conn) {
	defer conn.Close()

	Logger.Debug("proxy: connection open")
	defer Logger.Debug("proxy: connection closed")

	// Connect to proxy.
	proxyConn, err := net.Dial("tcp", p.addr)
	if err != nil {
		Logger.Debug("proxy: cannot connect to proxy", zap.String("address", p.addr))
		return
	}
	defer proxyConn.Close()

	// Copy between connection and proxy until an error occurs.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(proxyConn, conn) }()
	go func() { defer wg.Done(); io.Copy(conn, proxyConn) }()
	wg.Wait()
}

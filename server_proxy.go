package marionette

import (
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
)

// ServerProxy represents a proxy between a marionette listener and another server.
type ServerProxy struct {
	ln   *Listener
	addr string
	wg   sync.WaitGroup
}

// NewServerProxy returns a new instance of ServerProxy.
func NewServerProxy(ln *Listener, addr string) *ServerProxy {
	return &ServerProxy{
		ln:   ln,
		addr: addr,
	}
}

func (p *ServerProxy) Open() error {
	p.wg.Add(1)
	go func() { defer p.wg.Done(); p.run() }()

	return nil
}

func (p *ServerProxy) Close() error {
	return nil
}

func (p *ServerProxy) run() {
	Logger.Debug("server proxy: listening")
	defer Logger.Debug("server proxy: closed")

	for {
		conn, err := p.ln.Accept()
		if err != nil {
			Logger.Debug("server proxy: listener error", zap.Error(err))
			return
		}

		p.wg.Add(1)
		go func() { defer p.wg.Done(); p.handleConn(conn) }()
	}
}

func (p *ServerProxy) handleConn(conn net.Conn) {
	defer conn.Close()

	Logger.Debug("server proxy: connection open")
	defer Logger.Debug("server proxy: connection closed")

	// Connect to remote server.
	proxyConn, err := net.Dial("tcp", p.addr)
	if err != nil {
		Logger.Debug("server proxy: cannot connect to remote server", zap.String("address", p.addr))
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

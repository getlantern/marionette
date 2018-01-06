package marionette

import (
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
)

// ClientProxy represents a proxy between incoming connections and a marionette dialer.
type ClientProxy struct {
	ln     net.Listener
	dialer *Dialer
	wg     sync.WaitGroup
}

// NewClientProxy returns a new instance of ClientProxy.
func NewClientProxy(ln net.Listener, dialer *Dialer) *ClientProxy {
	return &ClientProxy{
		ln:     ln,
		dialer: dialer,
	}
}

func (p *ClientProxy) Open() error {
	p.wg.Add(1)
	go func() { defer p.wg.Done(); p.run() }()

	return nil
}

func (p *ClientProxy) Close() error {
	return nil
}

func (p *ClientProxy) run() {
	Logger.Debug("client proxy: listening")
	defer Logger.Debug("client proxy: closed")

	for {
		conn, err := p.ln.Accept()
		if err != nil {
			Logger.Debug("client proxy: listener error", zap.Error(err))
			return
		}

		p.wg.Add(1)
		go func() { defer p.wg.Done(); p.handleConn(conn) }()
	}
}

func (p *ClientProxy) handleConn(incomingConn net.Conn) {
	defer incomingConn.Close()

	Logger.Debug("client proxy: connection open")
	defer Logger.Debug("client proxy: connection closed")

	// Create a new stream.
	stream, err := p.dialer.Dial()
	if err != nil {
		Logger.Debug("client proxy: cannot connect create new stream", zap.Error(err))
		return
	}
	defer stream.Close()

	// Copy between incoming connection and stream until an error occurs.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(incomingConn, stream) }()
	go func() { defer wg.Done(); io.Copy(stream, incomingConn) }()
	wg.Wait()
}

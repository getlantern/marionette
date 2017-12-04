package marionette

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/redjack/marionette/mar"
)

type Server struct {
	ln net.Listener
	wg sync.WaitGroup

	doc *mar.Document

	// enc *CellEncoder
	dec *CellDecoder

	// Network interface name.
	Interface string

	Logger *log.Logger
}

func NewServer(doc *mar.Document) *Server {
	return &Server{
		Interface: "127.0.0.1",
		Logger:    log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (s *Server) Open() (err error) {
	// Parse port from MAR specification.
	// TODO: Handle "ftp_pasv_port".
	port, err := strconv.Atoi(s.doc.Model.Port)
	if err != nil {
		return errors.New("invalid connection port")
	}

	// Open port to listen for new connections.
	if s.ln, err = net.Listen(s.doc.Model.Transport, net.JoinHostPort(s.Interface, strconv.Itoa(port))); err != nil {
		return err
	}

	// Hand off connection handling to separate goroutine.
	s.wg.Add(1)
	go func() { defer s.wg.Done(); s.serve() }()

	return nil
}

func (s *Server) Close() (err error) {
	if s.ln != nil {
		err = s.ln.Close()
	}

	s.wg.Wait()
	return nil
}

func (s *Server) serve() {
	for {
		// Wait for next connection.
		conn, err := s.ln.Accept()
		if err != nil {
			s.Logger.Printf("listen error: %s", err)
			return
		}

		// Build a new execution engine.
		e := NewExecutor(s.doc)
		e.Encoder, e.Decoder = s.enc, s.dec

		// Execute in a separate goroutine.
		s.wg.Add(1)
		go func() { defer s.wg.Done(); e.Execute() }()
	}
}

package marionette

import (
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/redjack/marionette/mar"
)

type Server struct {
	ln net.Listener

	wg      sync.WaitGroup
	closing chan struct{}

	doc *mar.Document

	bufferSet *StreamBufferSet
	dec       *CellDecoder

	connID int

	// Network interface name.
	Interface string

	Logger *log.Logger
}

func NewServer(doc *mar.Document) *Server {
	return &Server{
		doc:     doc,
		closing: make(chan struct{}),

		Interface: "127.0.0.1",
		Logger:    log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (s *Server) Open() (err error) {
	// Parse port from MAR specification.
	// TODO: Handle "ftp_pasv_port".
	port, err := strconv.Atoi(s.doc.Port)
	if err != nil {
		return errors.New("invalid connection port")
	}

	// Run execution in a separate goroutine.
	s.wg.Add(1)
	go func() { defer s.wg.Done(); s.execute() }()

	// Open port to listen for new connections.
	if s.ln, err = net.Listen(s.doc.Transport, net.JoinHostPort(s.Interface, strconv.Itoa(port))); err != nil {
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

	close(s.closing)
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

		// Generate a new id for the connection so we can track it.
		connID := s.connID
		s.connID++

		// Pass off connection handling to a separate goroutine.
		s.wg.Add(1)
		go func() { defer s.wg.Done(); s.handleConn(connID, conn) }()
	}
}

func (s *Server) handleConn(id int, conn net.Conn) {
	log.Printf("conn(%d): connected", id)
	defer log.Printf("conn(%d): disconnected", id)

	// Continually read records from the connection and dump into decoder.
	dec := NewCellDecoder(conn)
	for {
		var cell Cell
		if err := dec.Decode(&cell); err != nil {
			log.Printf("conn(%d): decode error: %s", id, err)
			return
		}

		switch cell.Type {
		case END_OF_STREAM:
			log.Printf("conn(%d): end of stream", id)
			return

		case NORMAL:
			if len(cell.Payload) > 0 {
				s.bufferSet.Push(cell.StreamID, cell.Payload)
			}

		default:
			log.Printf("unexpected cell type: %d", cell.Type)
		}
	}
}

func (s *Server) execute() {
	for {
		// Check if server is closing in between executions.
		select {
		case <-s.closing:
		default:
		}

		// Start execution.
		e := NewExecutor(s.doc, PartyServer, s.bufferSet, s.dec)
		if err := e.Execute(); err != nil {
			log.Printf("execution error: %s", err)
		}
	}
}

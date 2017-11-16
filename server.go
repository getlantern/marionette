package marionette

import (
	"sync"
	"errors"
	"log"
	"net"

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
		Logger: log.New(os.Stderr, "", log.LstdFlags),
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
	go func() { defer s.wg.Done(); s.serve()}()

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

	// ServerDriver
        // self.party_ = party
        // self.running_ = []
        // self.num_executables_completed_ = 0
        // self.executable_ = None
        // self.multiplexer_outgoing_ = None
        // self.multiplexer_incoming_ = None
        // self.state_ = None

    func execute(self, reactor):
        for {
            new_executable = self.executable_.check_for_incoming_connections()
            if new_executable is None:
                break

            self.running_.append(new_executable)
            go new_executable.execute, reactor)
        }

        running_count = len(self.running_)
        self.running_ = [executable for executable
                         in self.running_
                         if executable.isRunning()]

        self.num_executables_completed_ += (running_count - len(self.running_))

    def isRunning(self):
        return len(self.running_)

    def setFormat(self, format_name, format_version=None):
        self.executable_ = marionette_tg.executable.Executable(self.party_, format_name,
                                                         format_version,
                                                         self.multiplexer_outgoing_,
                                                         self.multiplexer_incoming_)

    def set_multiplexer_outgoing(self, multiplexer):
        self.multiplexer_outgoing_ = multiplexer

    def set_multiplexer_incoming(self, multiplexer):
        self.multiplexer_incoming_ = multiplexer

    def set_state(self, state):
        self.state_ = state

        if self.state_:
            for key in self.state_.local_:
                if key not in marionette_tg.executables.pioa.RESERVED_LOCAL_VARS:
                    self.executable_.set_local(key, self.state_.local_[key])

    def stop(self):
        self.executable_.stop()

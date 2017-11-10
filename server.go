package marionette

import (
	"log"
)

type Server struct {
	Logger *log.Logger
}

func NewServer() *Server {
	return &Server{
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (s *Server) Open() (err error) {
	return nil
}

func (s *Server) Close() (err error) {
	return nil
}

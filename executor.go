package marionette

import (
	"github.com/redjack/marionette/mar"
)

type Executor struct {
	format string
	doc    *mar.Document
}

func NewExecutor(format string, doc *mar.Document) *Executor {
	return &Executor{format: format, doc: doc}
}

// ModelUUID returns the generated id based on the document.
func (e *Executor) ModelUUID() int {
	return e.doc.UUID
}

// TransportProtocol returns the protocol from the underlying document.
func (e *Executor) TransportProtocol() string {
	return e.doc.Model.Transport
}

// Port returns the port from the underlying document.
// If port is a named port then it is looked up in the local variables.
func (e *Executor) Port() int {
	if port, err := strconv.Atoi(e.doc.Model.Port); err == nil {
		return port
	}

	port, _ := e.locals[e.doc.Model.Port].(int)
	return port
}

// FirstSender returns the party that initiates the protocol.
func (e *Executor) FirstSender() string {
	if e.format == "ftp_pasv_transfer" {
		return PartyServer
	}
	return PartyClient
}

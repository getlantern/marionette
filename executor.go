package marionette

import (
	"github.com/redjack/marionette/mar"
)

// TODO: Support multiple states (e.g. HTTP 1.0 & HTTP 1.1)

type Executor struct {
	fsm *FSM

	enc *CellEncoder
	dec *CellDecoder
}

func NewExecutor(doc *mar.Document, enc *CellEncoder, dec *CellDecoder) *Executor {
	return &Executor{
		fsm: NewFSM(doc),
		enc: enc,
		dec: dec,
	}
}

func (e *Executor) Execute() error {
	for {
		if err := e.fsm.Next(); err != nil {
			return err
		}
		// TODO: Break when 'dead'.
	}

	// TODO: Close connection when FSM is done.
}

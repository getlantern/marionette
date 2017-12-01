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

func NewExecutor(doc *mar.Document, party string, enc *CellEncoder, dec *CellDecoder) *Executor {
	return &Executor{
		fsm: NewFSM(doc, party, enc, dec),
		enc: enc,
		dec: dec,
	}
}

func (e *Executor) Execute() error {
	for !e.fsm.Dead() {
		if err := e.fsm.Next(); err != nil {
			return err
		}
	}

	// TODO: Close connection when FSM is done.
	return nil
}

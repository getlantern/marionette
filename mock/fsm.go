package mock

import (
	"context"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/fte"
)

var _ marionette.FSM = &FSM{}

type FSM struct {
	UUIDFn          func() int
	InstanceIDFn    func() int
	SetInstanceIDFn func(int)
	PartyFn         func() string
	PortFn          func() int
	StateFn         func() string
	DeadFn          func() bool
	NextFn          func(ctx context.Context) error
	ExecuteFn       func(ctx context.Context) error
	ResetFn         func()
	ConnFn          func() *marionette.BufferedConn
	StreamSetFn     func() *marionette.StreamSet
	CipherFn        func(regex string, msgLen int) (*fte.Cipher, error)
	SetVarFn        func(key string, value interface{})
	VarFn           func(key string) interface{}
}

func (m *FSM) UUID() int {
	return m.UUIDFn()
}

func (m *FSM) InstanceID() int {
	return m.InstanceIDFn()
}

func (m *FSM) SetInstanceID(id int) {
	m.SetInstanceIDFn(id)
}

func (m *FSM) Party() string {
	return m.PartyFn()
}

func (m *FSM) Port() int {
	return m.PortFn()
}

func (m *FSM) State() string {
	return m.StateFn()
}

func (m *FSM) Dead() bool {
	return m.DeadFn()
}

func (m *FSM) Next(ctx context.Context) error {
	return m.NextFn(ctx)
}

func (m *FSM) Execute(ctx context.Context) error {
	return m.ExecuteFn(ctx)
}

func (m *FSM) Reset() {
	m.ResetFn()
}

func (m *FSM) Conn() *marionette.BufferedConn {
	return m.ConnFn()
}

func (m *FSM) StreamSet() *marionette.StreamSet {
	return m.StreamSetFn()
}

func (m *FSM) Cipher(regex string, msgLen int) (*fte.Cipher, error) {
	return m.CipherFn(regex, msgLen)
}

func (m *FSM) SetVar(key string, value interface{}) {
	m.SetVarFn(key, value)
}

func (m *FSM) Var(key string) interface{} {
	return m.VarFn(key)
}

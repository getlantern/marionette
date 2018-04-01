package mock

import (
	"context"
	"net"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

var _ marionette.FSM = &FSM{}

type FSM struct {
	CloseFn         func() error
	UUIDFn          func() int
	InstanceIDFn    func() int
	SetInstanceIDFn func(int)
	HostFn          func() string
	PartyFn         func() string
	PortFn          func() int
	StateFn         func() string
	DeadFn          func() bool
	NextFn          func(ctx context.Context) error
	ExecuteFn       func(ctx context.Context) error
	ResetFn         func()
	ListenFn        func() (int, error)
	ConnFn          func() *marionette.BufferedConn
	StreamSetFn     func() *marionette.StreamSet
	CipherFn        func(regex string) marionette.Cipher
	DFAFn           func(regex string, n int) marionette.DFA
	SetVarFn        func(key string, value interface{})
	VarFn           func(key string) interface{}
	CloneFn         func(doc *mar.Document) marionette.FSM
	LoggerFn        func() *zap.Logger

	BufferedConn *marionette.BufferedConn
}

// NewFSM returns an instance of FSM with conn and streamSet attached.
func NewFSM(conn net.Conn, streamSet *marionette.StreamSet) FSM {
	fsm := FSM{
		BufferedConn: marionette.NewBufferedConn(conn, marionette.MaxCellLength),
	}
	fsm.StateFn = func() string { return "default" }
	fsm.ConnFn = func() *marionette.BufferedConn { return fsm.BufferedConn }
	fsm.StreamSetFn = func() *marionette.StreamSet { return streamSet }
	fsm.LoggerFn = func() *zap.Logger { return marionette.Logger }
	return fsm
}

func (m *FSM) Close() error         { return m.CloseFn() }
func (m *FSM) UUID() int            { return m.UUIDFn() }
func (m *FSM) InstanceID() int      { return m.InstanceIDFn() }
func (m *FSM) SetInstanceID(id int) { m.SetInstanceIDFn(id) }
func (m *FSM) Host() string         { return m.HostFn() }
func (m *FSM) Party() string        { return m.PartyFn() }
func (m *FSM) Port() int            { return m.PortFn() }

func (m *FSM) State() string { return m.StateFn() }
func (m *FSM) Dead() bool    { return m.DeadFn() }

func (m *FSM) Next(ctx context.Context) error    { return m.NextFn(ctx) }
func (m *FSM) Execute(ctx context.Context) error { return m.ExecuteFn(ctx) }
func (m *FSM) Reset()                            { m.ResetFn() }

func (m *FSM) Listen() (int, error)             { return m.ListenFn() }
func (m *FSM) Conn() *marionette.BufferedConn   { return m.ConnFn() }
func (m *FSM) StreamSet() *marionette.StreamSet { return m.StreamSetFn() }

func (m *FSM) SetVar(key string, value interface{}) { m.SetVarFn(key, value) }
func (m *FSM) Var(key string) interface{}           { return m.VarFn(key) }

func (m *FSM) Cipher(regex string) marionette.Cipher {
	return m.CipherFn(regex)
}

func (m *FSM) DFA(regex string, msgLen int) marionette.DFA {
	return m.DFAFn(regex, msgLen)
}

func (m *FSM) Clone(doc *mar.Document) marionette.FSM { return m.CloneFn(doc) }

func (m *FSM) Logger() *zap.Logger { return m.LoggerFn() }

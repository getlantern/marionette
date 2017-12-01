package marionette

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"

	"github.com/redjack/marionette/fte"
	"github.com/redjack/marionette/mar"
)

type FSM struct {
	doc   *mar.Document
	party string
	enc   *CellEncoder
	dec   *CellDecoder

	state string
	stepN int
	rand  *rand.Rand
	conn  *bufConn // channel

	// Lookup of transitions by src state.
	transitions map[string][]*mar.Transition

	// Special variables.
	ModelInstanceID int

	vars        map[string]interface{}
	fteEncoders map[fteEncoderKey]*fte.Encoder
}

func NewFSM(doc *mar.Document, party string, enc *CellEncoder, dec *CellDecoder) *FSM {
	fsm := &FSM{
		doc:         doc,
		party:       party,
		state:       "start",
		enc:         enc,
		dec:         dec,
		vars:        make(map[string]interface{}),
		transitions: make(map[string][]*mar.Transition),
	}

	// Build transition map.
	for _, t := range doc.Transitions {
		fsm.transitions[t.Source] = append(fsm.transitions[t.Source], t)
	}

	// The initial sender generates the model instance ID.
	if party == doc.FirstSender() {
		fsm.ModelInstanceID = int(rand.Int31())
		fsm.rand = rand.New(rand.NewSource(int64(fsm.ModelInstanceID)))
	}

	return fsm
}

func (fsm *FSM) ModelUUID() int {
	return fsm.doc.UUID
}

// Port returns the port from the underlying document.
// If port is a named port then it is looked up in the local variables.
func (fsm *FSM) Port() int {
	if port, err := strconv.Atoi(fsm.doc.Port); err == nil {
		return port
	}

	// port, _ := fsm.locals[fsm.doc.Port].(int)
	// return port
	panic("TODO")
}

// Dead returns true when the FSM is complete.
func (fsm *FSM) Dead() bool { return fsm.state == "dead" }

func (fsm *FSM) Next() (err error) {
	// Create a new connection from the client if none is available.
	if fsm.party == PartyClient && fsm.conn == nil {
		const serverIP = "127.0.0.1" // TODO: Pass in.
		conn, err := net.Dial(fsm.doc.Transport, net.JoinHostPort(serverIP, fsm.doc.Port))
		if err != nil {
			return err
		}
		fsm.conn = newBufConn(conn)
	}

	// Exit if no connection available.
	if fsm.conn == nil {
		return errors.New("fsm.Next(): no connection available")
	}

	// Generate a new PRNG once we have an instance ID.
	fsm.init()

	// If we have a successful transition, update our state info.
	// Exit if no transitions were successful.
	if nextState := fsm.next(); nextState == "" {
		return errors.New("fsm.Next(): no matching transition action")
	} else {
		fsm.stepN += 1
		fsm.state = nextState
	}

	return nil
}

func (fsm *FSM) next() (nextState string) {
	// Find all possible transitions from the current state.
	// Then filter by PRNG (if available) or return all (if unavailable).
	transitions := mar.FilterTransitionsBySource(fsm.doc.Transitions, fsm.state)
	transitions = mar.ChooseTransitions(transitions, fsm.rand)

	// Extract just the destination names.
	destinations := mar.TransitionsDestinations(transitions)
	assert(len(destinations) > 0)

	// Append error state to try last if other destinations don't succeed.
	if errorState := mar.TransitionsErrorState(transitions); errorState != "" {
		destinations = append(destinations, errorState)
	}

	// Attempt each possible transition (and error transition at the end).
	for _, destination := range destinations {
		// Find all actions for this destination and current party.
		blk := fsm.doc.ActionBlock(destination)
		if blk == nil {
			continue
		}
		actions := mar.FilterActionsByParty(blk.Actions, fsm.party)

		// Attempt to execute each action.
		if matched, err := fsm.evalActions(actions); err != nil {
			log.Printf("EXCEPTION: %s", err)
		} else if matched {
			return destination
		}
	}
	return ""
}

// init initializes the PRNG if we now have a model instance id.
func (fsm *FSM) init() {
	if fsm.rand != nil || fsm.ModelInstanceID == 0 {
		return
	}

	// Create new PRNG.
	fsm.rand = rand.New(rand.NewSource(int64(fsm.ModelInstanceID)))

	// Restart FSM from the beginning and iterate until the current step.
	fsm.state = "start"
	for i := 0; i < fsm.stepN; i++ {
		fsm.state = fsm.next()
		assert(fsm.state != "")
	}
}

func (fsm *FSM) next_transition(src_state, dst_state string) *mar.Transition {
	for _, transition := range fsm.transitions[src_state] {
		if transition.Destination == dst_state {
			return transition
		}
	}
	return nil
}

func (fsm *FSM) evalActions(actions []*mar.Action) (bool, error) {
	if len(actions) == 0 {
		return true, nil
	}

	for _, action := range actions {
		// If there is no matching regex then simply evaluate action.
		if action.Regex != "" {
			// Compile regex.
			// TODO(benbjohnson): Compile at parse time and store.
			re, err := regexp.Compile(action.Regex)
			if err != nil {
				return false, err
			}

			// Only evaluate action if buffer matches.
			incoming_buffer := fsm.conn.Peek()
			if !re.Match(incoming_buffer) {
				continue
			}
		}

		if success, err := fsm.evalAction(action); err != nil {
			return false, err
		} else if success {
			return true, nil
		}
		continue
	}

	return false, nil
}

func (fsm *FSM) evalAction(action *mar.Action) (bool, error) {
	fn := FindPlugin(action.Name, action.Method)
	if fn == nil {
		return false, fmt.Errorf("fsm.evalAction(): action not found: %s.%s", action.Name, action.Method)
	}
	return fn(fsm, action.ArgValues())
}

/*
func (fsm *FSM) do_precomputations() {
	for _, action := range fsm.actions_ {
		if action.module_ == "fte" && strings.HasPrefix(action.method_, "send") {
			fsm.get_fte_obj(action.Arg(0), action.Arg(1))
		}
	}
}


func transitionKeys(transitions map[string]PATransition) []string {
	a := make([]string, 0, len(transitions))
	for k := range transitions {
		a = append(a, k)
	}
	sort.Strings(a)
	return a
}

func (fsm *FSM) isRunning() bool {
	return fsm.state != "dead"
}

func (fsm *FSM) add_state(name string) {
	if !stringSliceContains(fsm.states_.keys(), name) {
		fsm.states_[name] = PAState(name)
	}
}

func (fsm *FSM) set_multiplexer_outgoing(multiplexer *OutgoingBuffer) {
	fsm.global["multiplexer_outgoing"] = multiplexer
}

func (fsm *FSM) set_multiplexer_incoming(multiplexer *IncomingBuffer) {
	fsm.global["multiplexer_incoming"] = multiplexer
}

func (fsm *FSM) stop() {
	fsm.state = "dead"
}

func (fsm *FSM) set_port(port int) { // TODO: Maybe string?
	fsm.port_ = port
}

func (fsm *FSM) get_port() int {
	if fsm.port_ != 0 {
		return fsm.port_
	}
	return fsm.local[fsm.port_]
}

*/

func (fsm *FSM) Var(key string) interface{} {
	switch key {
	case "model_instance_id":
		return fsm.ModelInstanceID
	case "model_uuid":
		return fsm.doc.UUID
	case "party":
		return fsm.party
	case "multiplexer_incoming":
		return fsm.dec
	case "multiplexer_outgoing":
		return fsm.enc
	default:
		return fsm.vars[key]
	}
}

func (fsm *FSM) SetVar(key string, value interface{}) {
	fsm.vars[key] = value
}

func (fsm *FSM) fteEncoder(regex string, msgLen int) *fte.Encoder {
	key := fteEncoderKey{regex, msgLen}
	if fsm.fteEncoders[key] == nil {
		enc := fte.NewEncoder(regex, msgLen)
		fsm.fteEncoders[key] = enc
	}
	return fsm.fteEncoders[key]
}

type fteEncoderKey struct {
	regex  string
	msgLen int
}

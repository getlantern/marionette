package marionette

import (
	"math/rand"
	"net"
	"strconv"

	"github.com/redjack/marionette/mar"
)

type FSM struct {
	doc   *mar.Document
	party string

	state string
	stepN int
	rand  *rand.Rand

	// Lookup of transitions by src state.
	transitions map[string][]*mar.Transition

	vars map[string]interface{}
}

func NewFSM(doc *mar.Document, party string) *FSM {
	fsm := &FSM{
		doc:         doc,
		party:       party,
		state:       "start",
		vars:        make(map[string]interface{}),
		transitions: make(map[string][]*mar.Transition),
	}

	// Build transition map.
	for _, t := range doc.Model.Transitions {
		fsm.transitions[t.Source] = append(fsm.transitions[t.Source], t)
	}

	// The initial sender generates the model instance ID.
	if party == doc.FirstSender() {
		fsm.modelInstanceID = uint32(rand.Int31())
		fsm.rand = rand.New(rand.NewSource(int64(fsm.modelInstanceID)))
	}

	return fsm
}

// Port returns the port from the underlying document.
// If port is a named port then it is looked up in the local variables.
func (fsm *FSM) Port() int {
	if port, err := strconv.Atoi(fsm.doc.Model.Port); err == nil {
		return port
	}

	// port, _ := fsm.locals[fsm.doc.Model.Port].(int)
	// return port
	panic("TODO")
}

func (fsm *FSM) Next() error {
	// Create a new connection from the client if none is available.
	if fsm.party == PartyClient && fsm.channel == nil && !fsm.channelRequested {
		const serverIP = "127.0.0.1" // TODO: Pass in.
		fsm.channel = net.Dial(fsm.doc.Model.Transport, net.JoinHostPort(serverIP, fsm.doc.Model.Port))
		fsm.channelRequested = true
	}

	// Exit if no channel available.
	if fsm.channel == nil {
		return false
	}

	// Generate a new PRNG once we have an instance ID.
	fsm.init()

	transitions := fsm.chooseNextTransitions()

	assert(len(transitions) > 0)

	// attempt to do a normal transition
	var fatal int
	var success bool
	var dst_state string
	for _, dst_state = range transitions {
		action_block := fsm.determine_action_block(fsm.state, dst_state)

		if success, err := fsm.eval_action_block(action_block); err != nil {
			log.Printf("EXCEPTION: %s", err)
			fatal += 1
		} else if success {
			break
		}
	}

	// if all potential transitions are fatal, attempt the error transition
	if !success && fatal == len(transitions) {
		src_state := fsm.state
		dst_state = fsm.states_[fsm.state].error_state_

		if dst_state != "" {
			action_block := fsm.determine_action_block(src_state, dst_state)
			success, _ = fsm.eval_action_block(action_block)
		}
	}

	if !success {
		return false
	}

	// if we have a successful transition, update our state info.
	fsm.stepN += 1
	fsm.state = dst_state
	fsm.next_state_ = ""

	if fsm.state == "dead" {
		fsm.success_ = true
	}
	return true
}

// init initializes the PRNG if we now have a model instance id.
func (fsm *FSM) init() {
	if fsm.rand != nil || fsm.modelInstanceID == 0 {
		return
	}

	// Create new PRNG.
	fsm.rand = rand.New(rand.NewSource(int64(fsm.modelInstanceID)))

	// Restart FSM from the beginning and iterate until the current step.
	fsm.state = "start"
	for i := 0; i < fsm.stepN; i++ {
		fsm.state = fsm.states[fsm.state].transition(fsm.rand)
	}
	fsm.nextState = ""
}

func (fsm *FSM) chooseNextTransitions() []*mar.Transition {
	// If PRNG is not set yet then return all transitions with a non-zero probability.
	if fsm.rand == nil {
		var a []*mar.Transition
		for _, t := range fsm.transition[fsm.state] {
			if t.Probability > 0 {
				a = append(t)
			}
		}
		return a
	}

	// If there is only one transition then return it.
	transitions := fsm.transition[fsm.state]
	if len(transitions) == 1 {
		return transitions
	}

	// Otherwise randomly choose a transition based on probability.
	sum, coin := float64(0), fsm.rand.Float64()
	for _, t := range transitions {
		if t.Probability <= 0 {
			continue
		}
		sum += t.Probability
		if sum >= coin {
			return []*mar.Transition{t}
		}
	}
	return []*mar.Transition{transitions[len(transitions)-1]}
}

/*
func (fsm *FSM) do_precomputations() {
	for _, action := range fsm.actions_ {
		if action.module_ == "fte" && strings.HasPrefix(action.method_, "send") {
			fsm.get_fte_obj(action.Arg(0), action.Arg(1))
		}
	}
}

func (fsm *FSM) determine_action_block(src_state, dst_state string) []*Action {
	var retval []*Action
	for _, action := range fsm.actions_ {
		action_name := fsm.states_[src_state].transitions_[dst_state].name
		if action.party_ == fsm.party_ && action.name_ == action_name {
			retval = append(retval, action)
		}
	}
	return retval
}

func transitionKeys(transitions map[string]PATransition) []string {
	a := make([]string, 0, len(transitions))
	for k := range transitions {
		a = append(a, k)
	}
	sort.Strings(a)
	return a
}

func (fsm *FSM) eval_action_block(action_block []*Action) (bool, error) {
	var retval bool

	if len(action_block) == 0 {
		return true, nil
	}

	if len(action_block) >= 1 {
		for _, action_obj := range action_block {
			if action_obj.regex_match_incoming_ != nil {
				incoming_buffer := fsm.channel_.peek()
				if action_obj.regex_match_incoming_.Match(incoming_buffer) {
					retval = fsm.eval_action(action_obj)
				}
			} else {
				retval = fsm.eval_action(action_obj)
			}
			if retval {
				break
			}
		}
	}

	return retval, nil
}

func (fsm *FSM) isRunning() bool {
	return fsm.state != "dead"
}

func (fsm *FSM) eval_action(action_obj *Action) {
	module := action_obj.get_module()
	method := action_obj.get_method()
	args := action_obj.get_args()

	i := importlib.import_module("marionette_tg.plugins._" + module)
	method_obj = getattr(i, method)

	return method_obj(fsm.channel_, fsm.marionette_state_, args)
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

func (fsm *FSM) get_fte_obj(regex, msg_len string) interface{} {
	fte_key := fmt.Sprintf("fte_obj-%s%d", regex, msg_len)
	if _, ok := fsm.global[fte_key]; !ok {
		dfa := regex2dfa.Regex2dfa(regex)
		fte_obj := fte.Encode(dfa, msg_len)
		poia.global[fte_key] = fte_obj
	}

	return poia.global[fte_key]
}
*/

func (fsm *FSM) GetVar(key string) interface{} {
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

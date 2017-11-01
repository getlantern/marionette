package marionette

import (
	"encoding/binary"
	"log"
	"math/rand"
	"sort"
	"strings"
)

type PIOA struct {
	actions_            []*Action
	channel_            *Channel
	channel_requested_  bool
	current_state_      string //= 'start'
	first_sender_       string
	next_state_         string
	party_              string
	port_               int
	transport_protocol_ string
	rng_                *rand.Rand
	history_len_        int
	states_             map[string]*PAState
	success_            bool

	global map[string]interface{}
	local  map[string]interface{}
}

func NewPIOA(party, first_sender string) *PIOA {
	pioa := &PIOA{
		party_:         party,
		first_sender_:  first_sender,
		current_state_: "start",
		global:         make(map[string]interface{}),
		local:          map[string]interface{}{"party": party},
		states_:        make(map[string]*PAState),
	}

	if party == first_sender {
		source := uint32(rand.Int31())
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, source)
		pioa.local["model_instance_id"] = buf
		pioa.rng_ = rand.New(rand.NewSource(int64(source)))
	}

	return pioa
}

func (pioa *PIOA) do_precomputations() {
	for _, action := range pioa.actions_ {
		if action.module_ == "fte" && strings.HasPrefix(action.method_, "send") {
			pioa.get_fte_obj(action.Arg(0), action.Arg(1))
		}
	}
}

func (pioa *PIOA) execute() {
	if pioa.isRunning() {
		pioa.transition()
		// reactor.callLater(EVENT_LOOP_FREQUENCY_S, pioa.execute, reactor)
	} else {
		pioa.channel_.close()
	}
}

func (pioa *PIOA) check_channel_state() bool {
	if pioa.party_ == "client" {
		if pioa.channel_ != nil {
			if !pioa.channel_requested_ {
				panic("TODO: Create new TCP or UDP connection. No need for twisted callbacks.")
				// open_new_channel(pioa.transport_protocol_, pioa.get_port(), pioa.set_channel)
				pioa.channel_requested_ = true
			}
		}
	}
	return pioa.channel_ != nil
}

func (pioa *PIOA) set_channel(channel *Channel) {
	pioa.channel_ = channel
}

func (pioa *PIOA) check_rng_state() {
	if pioa.local["model_instance_id"] == nil {
		return
	}

	if pioa.rng_ != nil {
		model_instance_id := pioa.local["model_instance_id"].(int)
		pioa.rng_ = rand.New(rand.NewSource(int64(model_instance_id)))

		pioa.current_state_ = "start"
		for i := 0; i < pioa.history_len_; i++ {
			pioa.current_state_ = pioa.states_[pioa.current_state_].transition(pioa.rng_)
		}
		pioa.next_state_ = ""
	}

	// Reset history length once RNGs are sync'd
	pioa.history_len_ = 0
}

func (pioa *PIOA) determine_action_block(src_state, dst_state string) []*Action {
	var retval []*Action
	for _, action := range pioa.actions_ {
		action_name := pioa.states_[src_state].transitions_[dst_state].name
		if action.party_ == pioa.party_ && action.name_ == action_name {
			retval = append(retval, action)
		}
	}
	return retval
}

func (pioa *PIOA) get_potential_transitions() []string {
	var retval []string

	if pioa.rng_ != nil {
		if pioa.next_state_ == "" {
			pioa.next_state_ = pioa.states_[pioa.current_state_].transition(pioa.rng_)
		}
		retval = append(retval, pioa.next_state_)
	} else {
		for _, transition := range transitionKeys(pioa.states_[pioa.current_state_].transitions_) {
			if pioa.states_[pioa.current_state_].transitions_[transition].probability > 0 {
				retval = append(retval, transition)
			}
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

func (pioa *PIOA) advance_to_next_state() bool {
	// get the list of possible transitions we could make
	potential_transitions := pioa.get_potential_transitions()
	assert(len(potential_transitions) > 0)

	// attempt to do a normal transition
	var fatal int
	var success bool
	var dst_state string
	for _, dst_state = range potential_transitions {
		action_block := pioa.determine_action_block(pioa.current_state_, dst_state)

		if success, err := pioa.eval_action_block(action_block); err != nil {
			log.Printf("EXCEPTION: %s", err)
			fatal += 1
		} else if success {
			break
		}
	}

	// if all potential transitions are fatal, attempt the error transition
	if !success && fatal == len(potential_transitions) {
		src_state := pioa.current_state_
		dst_state = pioa.states_[pioa.current_state_].get_error_transition()

		if dst_state != "" {
			action_block := pioa.determine_action_block(src_state, dst_state)
			success, _ = pioa.eval_action_block(action_block)
		}
	}

	if !success {
		return false
	}

	// if we have a successful transition, update our state info.
	pioa.history_len_ += 1
	pioa.current_state_ = dst_state
	pioa.next_state_ = ""

	if pioa.current_state_ == "dead" {
		pioa.success_ = true
	}
	return true
}

func (pioa *PIOA) eval_action_block(action_block []*Action) (bool, error) {
	var retval bool

	if len(action_block) == 0 {
		return true, nil
	}

	if len(action_block) >= 1 {
		for _, action_obj := range action_block {
			if action_obj.regex_match_incoming_ != nil {
				incoming_buffer := pioa.channel_.peek()
				if action_obj.regex_match_incoming_.Match(incoming_buffer) {
					retval = pioa.eval_action(action_obj)
				}
			} else {
				retval = pioa.eval_action(action_obj)
			}
			if retval {
				break
			}
		}
	}

	return retval, nil
}

func (pioa *PIOA) transition() {
	var success bool
	if pioa.check_channel_state() {
		pioa.check_rng_state()
		success = pioa.advance_to_next_state()
	}
	return success
}

func (pioa *PIOA) replicate() *PIOA {
	other := NewPIOA(pioa.party_, pioa.first_sender_)
	other.actions_ = pioa.actions_
	other.states_ = pioa.states_
	other.global_ = pioa.global_
	other.local["model_uuid"] = pioa.local["model_uuid"]
	other.port_ = pioa.port_
	other.transport_protocol_ = pioa.transport_protocol_
	return other
}

func (pioa *PIOA) isRunning() bool {
	return pioa.current_state_ != "dead"
}

func (pioa *PIOA) eval_action(action_obj *Action) {
	module := action_obj.get_module()
	method := action_obj.get_method()
	args := action_obj.get_args()

	i := importlib.import_module("marionette_tg.plugins._" + module)
	method_obj = getattr(i, method)

	return method_obj(pioa.channel_, pioa.marionette_state_, args)
}

func (pioa *PIOA) add_state(name string) {
	if !stringSliceContains(pioa.states_.keys(), name) {
		pioa.states_[name] = PAState(name)
	}
}

func (pioa *PIOA) set_multiplexer_outgoing(multiplexer *OutgoingBuffer) {
	pioa.global["multiplexer_outgoing"] = multiplexer
}

func (pioa *PIOA) set_multiplexer_incoming(multiplexer *IncomingBuffer) {
	pioa.global["multiplexer_incoming"] = multiplexer
}

func (pioa *PIOA) stop() {
	pioa.current_state_ = "dead"
}

func (pioa *PIOA) set_port(port int) { // TODO: Maybe string?
	pioa.port_ = port
}

func (pioa *PIOA) get_port() int {
	if pioa.port_ != 0 {
		return pioa.port_
	}
	return pioa.local[pioa.port_]
}

func (pioa *PIOA) get_fte_obj(regex, msg_len string) interface{} {
	fte_key := fmt.Sprintf("fte_obj-%s%d", regex, msg_len)
	if _, ok := pioa.global[fte_key]; !ok {
		dfa := regex2dfa.Regex2dfa(regex)
		fte_obj := fte.Encode(dfa, msg_len)
		poia.global[fte_key] = fte_obj
	}

	return poia.global[fte_key]
}

type PAState struct {
	name_        string
	transitions_ map[string]PATransition
	// format_type_  interface{} // = None
	// format_value_ interface{} // = None
	error_state_ string // = None
}

func NewPAState(name string) *PAState {
	return &PAState{
		name_:        name,
		transitions_: make(map[string]PATransition),
	}
}

func (s *PAState) add_transition(dst, action_name string, probability float64) {
	s.transitions_[dst] = PATransition{name: action_name, probability: probability}
}

func (s *PAState) set_error_transition(error_state string) {
	s.error_state_ = error_state
}

func (s *PAState) get_error_transition() string {
	return s.error_state_
}

func (s *PAState) transition(rng *rand.Rand) string {
	assert(rng != nil || len(s.transitions_) == 1)
	if rng && len(s.transitions_) > 1 {
		coin = rng.random()
		sum = 0
		for _, state := range s.transitions_ {
			if s.transitions_[state].probability == 0 {
				continue
			}
			sum += s.transitions_[state].probability
			if sum >= coin {
				break
			}
		}
	} else {
		state = s.firstTransitionKey()
	}
	return state
}

func (s *PAState) firstTransitionKey() string {
	if len(s.transitions_) == 0 {
		return ""
	}

	a := make([]string, 0, len(s.transitions_))
	for k := range s.transitions_ {
		a = append(a, k)
	}
	sort.Strings(a)
	return a[0]
}

// NOTE: This replaces an untyped tuple.
type PATransition struct {
	name        string
	probability float64
}

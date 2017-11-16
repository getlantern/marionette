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

type PAState struct {
	name        string
	transitions map[string]PATransition
	errorState  string
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

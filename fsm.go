package marionette

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/redjack/marionette/fte"
	"github.com/redjack/marionette/mar"
	"go.uber.org/zap"
)

var ErrNoTransition = errors.New("no matching transition")

type FSM struct {
	doc     *mar.Document
	party   string
	streams *StreamSet

	state string
	stepN int
	rand  *rand.Rand

	conn net.Conn // channel
	buf  []byte   // connection buffer

	// Lookup of transitions by src state.
	transitions map[string][]*mar.Transition

	vars    map[string]interface{}
	ciphers map[cipherKey]*fte.Cipher

	// Set by the first sender and used to seed PRNG.
	InstanceID int
}

// NewFSM returns a new FSM. If party is the first sender then the instance id is set.
func NewFSM(doc *mar.Document, party string) *FSM {
	fsm := &FSM{
		doc:         doc,
		party:       party,
		streams:     NewStreamSet(),
		transitions: make(map[string][]*mar.Transition),
		buf:         make([]byte, 0, MaxCellLength),
	}
	fsm.Reset()

	// Build transition map.
	for _, t := range doc.Transitions {
		fsm.transitions[t.Source] = append(fsm.transitions[t.Source], t)
	}

	// The initial sender generates the instance ID.
	if party == doc.FirstSender() {
		fsm.InstanceID = int(rand.Int31())
		fsm.rand = rand.New(rand.NewSource(int64(fsm.InstanceID)))
	}

	return fsm
}

func (fsm *FSM) Reset() {
	fsm.state = "start"
	fsm.vars = make(map[string]interface{})

	for _, c := range fsm.ciphers {
		if err := c.Close(); err != nil {
			fsm.logger().Error("cannot close cipher", zap.Error(err))
		}
	}
	fsm.ciphers = make(map[cipherKey]*fte.Cipher)
}

func (fsm *FSM) UUID() int {
	return fsm.doc.UUID
}

// State returns the current state of the FSM.
func (fsm *FSM) State() string {
	return fsm.state
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

// Execute runs the the FSM to completion.
func (fsm *FSM) Execute(ctx context.Context) error {
	fsm.Reset()

	for !fsm.Dead() {
		if err := fsm.Next(ctx); err == ErrNoTransition {
			time.Sleep(100 * time.Millisecond)
			continue
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (fsm *FSM) Next(ctx context.Context) (err error) {
	assert(fsm.conn != nil)

	logger := fsm.logger()
	logger.Debug("fsm: Next()", zap.String("state", fsm.state))

	// Generate a new PRNG once we have an instance ID.
	if err := fsm.init(); err != nil {
		logger.Debug("fsm: cannot initialize fsm", zap.Error(err))
		return err
	}

	// If we have a successful transition, update our state info.
	// Exit if no transitions were successful.
	if nextState, err := fsm.next(); err != nil {
		logger.Debug("fsm: cannot move to next state")
		return err
	} else if nextState == "" {
		logger.Debug("fsm: no transition available")
		return ErrNoTransition
	} else {
		fsm.stepN += 1
		fsm.state = nextState
		logger.Debug("fsm: transition successful", zap.String("state", fsm.state), zap.Int("step", fsm.stepN))
	}

	return nil
}

func (fsm *FSM) next() (nextState string, err error) {
	logger := fsm.logger()

	// Find all possible transitions from the current state.
	transitions := mar.FilterTransitionsBySource(fsm.doc.Transitions, fsm.state)
	errorTransitions := mar.FilterErrorTransitions(transitions)

	// Then filter by PRNG (if available) or return all (if unavailable).
	transitions = mar.FilterNonErrorTransitions(transitions)
	transitions = mar.ChooseTransitions(transitions, fsm.rand)
	assert(len(transitions) > 0)

	logger.Debug("fsm: evaluating transitions", zap.Int("n", len(transitions)))

	// Add error transitions back in after selection.
	transitions = append(transitions, errorTransitions...)

	// Attempt each possible transition.
	for _, transition := range transitions {
		logger.Debug("fsm: evaluating transition", zap.String("src", transition.Source), zap.String("dest", transition.Destination))

		// If there's no action block then move to the next state.
		if transition.ActionBlock == "NULL" {
			logger.Debug("fsm: no action block, matched")
			return transition.Destination, nil
		}

		// Find all actions for this destination and current party.
		blk := fsm.doc.ActionBlock(transition.ActionBlock)
		if blk == nil {
			return "", fmt.Errorf("fsm.Next(): action block not found: %q", transition.ActionBlock)
		}
		actions := mar.FilterActionsByParty(blk.Actions, fsm.party)

		// Attempt to execute each action.
		logger.Debug("fsm: evaluating action block", zap.String("name", transition.ActionBlock), zap.Int("actions", len(actions)))
		if matched, err := fsm.evalActions(actions); err != nil {
			return "", err
		} else if matched {
			return transition.Destination, nil
		}
	}
	return "", nil
}

// init initializes the PRNG if we now have a instance id.
func (fsm *FSM) init() (err error) {
	if fsm.rand != nil || fsm.InstanceID == 0 {
		return nil
	}

	logger := fsm.logger()
	logger.Debug("fsm: initializing fsm")

	// Create new PRNG.
	fsm.rand = rand.New(rand.NewSource(int64(fsm.InstanceID)))

	// Restart FSM from the beginning and iterate until the current step.
	fsm.state = "start"
	for i := 0; i < fsm.stepN; i++ {
		fsm.state, err = fsm.next()
		if err != nil {
			return err
		}
		assert(fsm.state != "")
	}
	return nil
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
	logger := fsm.logger()

	if len(actions) == 0 {
		logger.Debug("fsm: no actions, matched")
		return true, nil
	}

	for _, action := range actions {
		logger.Debug("fsm: evaluating action", zap.String("name", action.Module+"."+action.Method), zap.String("regex", action.Regex))

		// If there is no matching regex then simply evaluate action.
		if action.Regex != "" {
			// Compile regex.
			re, err := regexp.Compile(action.Regex)
			if err != nil {
				return false, err
			}

			// Only evaluate action if buffer matches.
			buf, err := fsm.ReadBuffer()
			if err != nil {
				return false, err
			} else if !re.Match(buf) {
				logger.Debug("fsm: no regex match, skipping")
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
	fn := FindPlugin(action.Module, action.Method)
	if fn == nil {
		return false, fmt.Errorf("fsm.evalAction(): action not found: %s", action.Name())
	}
	fsm.logger().Debug("fsm: execute plugin", zap.String("name", action.Name()))
	return fn(fsm, action.ArgValues())
}

// ReadBuffer reads available data into the read buffer and returns the buffer.
func (fsm *FSM) ReadBuffer() ([]byte, error) {
	// Buffer full, exit.
	if len(fsm.buf) == cap(fsm.buf) {
		fsm.logger().Debug("fsm: buffer full")
		return fsm.buf, nil
	}

	// Read from connection with any available buffer space.
	if err := fsm.conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond)); err != nil {
		return nil, err
	}
	n, err := fsm.conn.Read(fsm.buf[len(fsm.buf):cap(fsm.buf)])
	if err != nil {
		if isTimeoutError(err) {
			return fsm.buf, nil
		}
		return fsm.buf, err
	}
	fsm.buf = fsm.buf[:len(fsm.buf)+int(n)]

	return fsm.buf, nil
}

// SetReadBuffer copies p to the buffer.
func (fsm *FSM) SetReadBuffer(p []byte) {
	assert(len(p) < cap(fsm.buf))
	fsm.buf = fsm.buf[:len(p)]
	copy(fsm.buf, p)
}

func (fsm *FSM) Var(key string) interface{} {
	switch key {
	case "model_instance_id":
		return fsm.InstanceID
	case "model_uuid":
		return fsm.doc.UUID
	case "party":
		return fsm.party
	default:
		return fsm.vars[key]
	}
}

func (fsm *FSM) SetVar(key string, value interface{}) {
	fsm.vars[key] = value
}

// Cipher returns a cipher with the given settings.
// If no cipher exists then a new one is created and returned.
func (fsm *FSM) Cipher(regex string, msgLen int) (_ *fte.Cipher, err error) {
	key := cipherKey{regex, msgLen}
	cipher := fsm.ciphers[key]
	if cipher != nil {
		return cipher, nil
	}

	cipher = fte.NewCipher(regex)
	if err := cipher.Open(); err != nil {
		return nil, err
	}

	fsm.ciphers[key] = cipher
	return cipher, nil
}

type cipherKey struct {
	regex  string
	msgLen int
}

func (fsm *FSM) logger() *zap.Logger {
	return Logger.With(zap.String("party", fsm.party))
}

// isTimeoutError returns true if the error is a timeout error.
func isTimeoutError(err error) bool {
	if err, ok := err.(interface {
		Timeout() bool
	}); ok && err.Timeout() {
		return true
	}
	return false
}

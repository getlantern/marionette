package fte

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os/exec"
	"strconv"
	"sync"
)

type DFA struct {
	mu   sync.Mutex
	once sync.Once

	regex    string
	msgLen   int
	filename string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	bufout *bufio.Reader

	capacity int
}

// NewDFA returns a new instance of DFA.
func NewDFA(regex string, msgLen int) *DFA {
	return &DFA{regex: regex, msgLen: msgLen}
}

// Open initializes the external DFA process.
func (dfa *DFA) Open() error {
	// Generate filename for temporary program file.
	f, err := ioutil.TempFile("", "fte-dfa-")
	if err != nil {
		return err
	} else if err := f.Close(); err != nil {
		return err
	}
	dfa.filename = f.Name() + ".py"

	// Copy program to temporary path.
	if err := ioutil.WriteFile(dfa.filename, []byte(dfaProgram), 0700); err != nil {
		return err
	}

	// Start process.
	dfa.cmd = exec.Command("python2", dfa.filename, dfa.regex, strconv.Itoa(dfa.msgLen))
	dfa.cmd.Stderr = stderr()
	if dfa.stdin, err = dfa.cmd.StdinPipe(); err != nil {
		return err
	} else if dfa.stdout, err = dfa.cmd.StdoutPipe(); err != nil {
		return err
	}
	dfa.bufout = bufio.NewReader(dfa.stdout)

	if err := dfa.cmd.Start(); err != nil {
		return err
	}

	// Read initial capacity.
	if line, err := dfa.bufout.ReadBytes('\n'); err != nil {
		return fmt.Errorf("fte.DFA.Open(): cannot read from stdout: %s", err)
	} else if dfa.capacity, err = strconv.Atoi(string(bytes.TrimSpace(line))); err != nil {
		return fmt.Errorf("fte.DFA.Open(): cannot convert capacity to int: %s", err)
	}
	return nil
}

// Close stops the cipher process.
func (dfa *DFA) Close() (err error) {
	dfa.once.Do(func() {
		if dfa.cmd != nil {
			if e := dfa.stdin.Close(); e != nil && err == nil {
				err = e
			}
			if e := dfa.cmd.Wait(); e != nil && err == nil {
				err = e
			}
			dfa.cmd = nil
		}
	})
	return err
}

func (dfa *DFA) init() error {
	if dfa.cmd != nil {
		return nil
	}
	return dfa.Open()
}

// Capacity returns the capacity left in the encoder.
func (dfa *DFA) Capacity() (int, error) {
	if err := dfa.init(); err != nil {
		return 0, err
	}
	return (dfa.capacity / 8), nil
}

// Rank maps s into an integer ranking.
func (dfa *DFA) Rank(s string) (rank *big.Int, err error) {
	dfa.mu.Lock()
	defer dfa.mu.Unlock()

	if err := dfa.init(); err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(dfa.stdin, "R%09d%x", len(s), s); err != nil {
		return nil, fmt.Errorf("fte.DFA.Rank(): cannot write to stdin: %s", err)
	}

	line, err := dfa.bufout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("fte.DFA.Rank(): cannot read from stdout: %s", err)
	}

	// Line is split into <rank,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 2)

	rank = &big.Int{}
	if _, ok := rank.SetString(string(bytes.TrimSpace(segments[0])), 10); !ok {
		return nil, fmt.Errorf("fte.DFA.Rank(): cannot decode big.Int: %q", bytes.TrimSpace(segments[0]))
	}
	if dfa.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[1]))); err != nil {
		return nil, fmt.Errorf("fte.DFA.Rank(): cannot convert capacity to int: %s", err)
	}
	return rank, nil
}

// Unrank reverses the map from an integer to a string.
func (dfa *DFA) Unrank(rank *big.Int) (ret string, err error) {
	dfa.mu.Lock()
	defer dfa.mu.Unlock()

	if err := dfa.init(); err != nil {
		return "", err
	}

	req := rank.String()
	if _, err := fmt.Fprintf(dfa.stdin, "U%09d%x", len(req), req); err != nil {
		return "", fmt.Errorf("fte.DFA.Unrank(): cannot write to stdin: %s", err)
	}

	line, err := dfa.bufout.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("fte.DFA.Unrank(): cannot read from stdout: %s", err)
	}

	// Line is split into <plaintext,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 2)

	var retBytes []byte
	if retBytes, err = hex.DecodeString(string(bytes.TrimSpace(segments[0]))); err != nil {
		return "", fmt.Errorf("fte.DFA.Unrank(): cannot decode hex: %s", err)
	}
	if dfa.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[1]))); err != nil {
		return "", fmt.Errorf("fte.DFA.Unrank(): cannot convert capacity to int: %s", err)
	}
	return string(retBytes), nil
}

// NumWordsInSlice executes DFA.getNumWordsInSlice.
func (dfa *DFA) NumWordsInSlice(n int) (numWords *big.Int, err error) {
	dfa.mu.Lock()
	defer dfa.mu.Unlock()

	req := strconv.Itoa(n)
	if _, err := fmt.Fprintf(dfa.stdin, "N%09d%x", len(req), req); err != nil {
		return nil, fmt.Errorf("fte.DFA.NumWordsInSlice(): cannot write to stdin: %s", err)
	}

	line, err := dfa.bufout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("fte.DFA.NumWordsInSlice(): cannot read from stdout: %s", err)
	}

	// Line is split into <numWords,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 2)

	numWords = &big.Int{}
	if _, ok := numWords.SetString(string(bytes.TrimSpace(segments[0])), 10); !ok {
		return nil, fmt.Errorf("fte.DFA.NumWordsInSlice(): cannot decode numWords to big int: %s", string(bytes.TrimSpace(segments[0])))
	}
	if dfa.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[1]))); err != nil {
		return nil, fmt.Errorf("fte.DFA.NumWordsInSlice(): cannot convert capacity to int: %s", err)
	}
	return numWords, nil
}

const dfaProgram = `
import sys
import regex2dfa
import fte.encoder
import binascii

regex = sys.argv[1]
msg_len = int(sys.argv[2])

dfa = regex2dfa.regex2dfa(regex)
cDFA = fte.cDFA.DFA(dfa, msg_len)
encoder = fte.dfa.DFA(cDFA, msg_len)

sys.stdout.write(str(encoder.getCapacity()))
sys.stdout.write("\n")
sys.stdout.flush()

def rank(payload):
	ret = encoder.rank(payload.encode('utf8'))
	sys.stdout.write(str(ret))
	sys.stdout.write(":")
	sys.stdout.write(str(encoder.getCapacity()))
	sys.stdout.write("\n")
	sys.stdout.flush()

def unrank(payload):
	v = encoder.unrank(int(payload))
	sys.stdout.write(binascii.hexlify(v))
	sys.stdout.write(":")
	sys.stdout.write(str(encoder.getCapacity()))
	sys.stdout.write("\n")
	sys.stdout.flush()

def num_words_in_slice(payload):
	num_words = encoder.getNumWordsInSlice(int(payload))
	sys.stdout.write(str(num_words))
	sys.stdout.write(":")
	sys.stdout.write(str(encoder.getCapacity()))
	sys.stdout.write("\n")
	sys.stdout.flush()

while True:
	cmd = sys.stdin.read(1)
	if cmd == "":
		break
	assert cmd == 'R' or cmd == 'U' or cmd == 'N'

	sz = int(sys.stdin.read(9))
	payload = binascii.unhexlify(sys.stdin.read(sz*2))

	if cmd == 'R':
		rank(payload)
	elif cmd == 'U':
		unrank(payload)
	else:
		num_words_in_slice(payload)
`

package fte

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	stderr io.ReadCloser

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
	dfa.cmd.Stderr = os.Stderr
	if dfa.stdin, err = dfa.cmd.StdinPipe(); err != nil {
		return err
	} else if dfa.stdout, err = dfa.cmd.StdoutPipe(); err != nil {
		return err
	}
	dfa.bufout = bufio.NewReader(dfa.stdout)

	if err := dfa.cmd.Start(); err != nil {
		return err
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

// Capacity returns the capacity left in the encoder.
func (dfa *DFA) Capacity() int {
	return (dfa.capacity / 8)
}

// Encrypt encrypts plaintext into ciphertext.
func (dfa *DFA) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	dfa.mu.Lock()
	defer dfa.mu.Unlock()

	if _, err := fmt.Fprintf(dfa.stdin, "E%09d%x", len(plaintext), plaintext); err != nil {
		return nil, fmt.Errorf("fte.DFA.Encrypt(): cannot write to stdin: %s", err)
	}

	line, err := dfa.bufout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("fte.DFA.Encrypt(): cannot read from stdout: %s", err)
	}

	// Line is split into <plaintext,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 2)

	if ciphertext, err = hex.DecodeString(string(bytes.TrimSpace(segments[0]))); err != nil {
		return nil, fmt.Errorf("fte.DFA.Encrypt(): cannot decode ciphertext hex: %s", err)
	}
	if dfa.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[1]))); err != nil {
		return nil, fmt.Errorf("fte.DFA.Encrypt(): cannot convert capacity to int: %s", err)
	}
	return ciphertext, nil
}

// Decrypt decrypts ciphertext into plaintext.
func (dfa *DFA) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	dfa.mu.Lock()
	defer dfa.mu.Unlock()

	if _, err := fmt.Fprintf(dfa.stdin, "D%09d%x", len(ciphertext), ciphertext); err != nil {
		return nil, fmt.Errorf("fte.DFA.Decrypt(): cannot write to stdin: %s", err)
	}

	line, err := dfa.bufout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("fte.DFA.Decrypt(): cannot read from stdout: %s", err)
	}

	// Line is split into <plaintext,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 2)

	if plaintext, err = hex.DecodeString(string(bytes.TrimSpace(segments[0]))); err != nil {
		return nil, fmt.Errorf("fte.DFA.Decrypt(): cannot decode plaintext hex: %s", err)
	}
	if dfa.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[2]))); err != nil {
		return nil, fmt.Errorf("fte.DFA.Decrypt(): cannot convert capacity to int: %s", err)
	}
	return plaintext, nil
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

def encode(payload):
	raise Exception(encoder.rank(payload))
	ciphertext = bytearray(encoder.rank(payload))
	sys.stdout.write(binascii.hexlify(ciphertext))
	sys.stdout.write(":")
	sys.stdout.write(str(encoder.getCapacity()))
	sys.stdout.write("\n")
	sys.stdout.flush()

def decode(payload):
	[plaintext, remainder] = encoder.unrank(payload)
	sys.stdout.write(binascii.hexlify(plaintext))
	sys.stdout.write(":")
	sys.stdout.write(str(encoder.getCapacity()))
	sys.stdout.write("\n")
	sys.stdout.flush()

while True:
	cmd = sys.stdin.read(1)
	if cmd == "":
		break
	assert cmd == 'E' or cmd == 'D'

	sz = int(sys.stdin.read(9))
	payload = binascii.unhexlify(sys.stdin.read(sz*2))

	if cmd == 'E':
		encode(payload)
	else:
		decode(payload)
`

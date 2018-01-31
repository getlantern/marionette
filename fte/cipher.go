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

type Cipher struct {
	mu   sync.Mutex
	once sync.Once

	regex    string
	filename string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	bufout *bufio.Reader
	stderr io.ReadCloser

	capacity int
}

// NewCipher returns a new instance of Cipher.
func NewCipher(regex string) *Cipher {
	return &Cipher{regex: regex}
}

// Open initializes the external cipher process.
func (c *Cipher) Open() error {
	// Generate filename for temporary program file.
	f, err := ioutil.TempFile("", "fte-cipher-")
	if err != nil {
		return err
	} else if err := f.Close(); err != nil {
		return err
	}
	c.filename = f.Name() + ".py"

	// Copy program to temporary path.
	if err := ioutil.WriteFile(c.filename, []byte(cipherProgram), 0700); err != nil {
		return err
	}

	// Start process.
	c.cmd = exec.Command("python2", c.filename, c.regex)
	c.cmd.Stderr = os.Stderr
	if c.stdin, err = c.cmd.StdinPipe(); err != nil {
		return err
	} else if c.stdout, err = c.cmd.StdoutPipe(); err != nil {
		return err
	}
	c.bufout = bufio.NewReader(c.stdout)

	if err := c.cmd.Start(); err != nil {
		return err
	}

	return nil
}

// Close stops the cipher process.
func (c *Cipher) Close() (err error) {
	c.once.Do(func() {
		if c.cmd != nil {
			if e := c.stdin.Close(); e != nil && err == nil {
				err = e
			}
			if e := c.cmd.Wait(); e != nil && err == nil {
				err = e
			}
			c.cmd = nil
		}
	})
	return err
}

// Capacity returns the capacity left in the encoder.
func (c *Cipher) Capacity() int {
	return (c.capacity / 8) - COVERTEXT_HEADER_LEN_CIPHERTTEXT - CTXT_EXPANSION
}

// Encrypt encrypts plaintext into ciphertext.
func (c *Cipher) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := fmt.Fprintf(c.stdin, "E%09d%x", len(plaintext), plaintext); err != nil {
		return nil, fmt.Errorf("fte.Cipher.Encrypt(): cannot write to stdin: %s", err)
	}

	line, err := c.bufout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("fte.Cipher.Encrypt(): cannot read from stdout: %s", err)
	}

	// Line is split into <plaintext,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 2)

	if ciphertext, err = hex.DecodeString(string(bytes.TrimSpace(segments[0]))); err != nil {
		return nil, fmt.Errorf("fte.Cipher.Encrypt(): cannot decode ciphertext hex: %s", err)
	}
	if c.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[1]))); err != nil {
		return nil, fmt.Errorf("fte.Cipher.Encrypt(): cannot convert capacity to int: %s", err)
	}
	return ciphertext, nil
}

// Decrypt decrypts ciphertext into plaintext.
func (c *Cipher) Decrypt(ciphertext []byte) (plaintext, remainder []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := fmt.Fprintf(c.stdin, "D%09d%x", len(ciphertext), ciphertext); err != nil {
		return nil, nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot write to stdin: %s", err)
	}

	line, err := c.bufout.ReadBytes('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot read from stdout: %s", err)
	}

	// Line is split into <plaintext,remainder,capacity>.
	segments := bytes.SplitN(line, []byte(":"), 3)

	if plaintext, err = hex.DecodeString(string(bytes.TrimSpace(segments[0]))); err != nil {
		return nil, nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot decode plaintext hex: %s", err)
	}
	if remainder, err = hex.DecodeString(string(bytes.TrimSpace(segments[1]))); err != nil {
		return nil, nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot decode remainder hex: %s", err)
	}
	if c.capacity, err = strconv.Atoi(string(bytes.TrimSpace(segments[2]))); err != nil {
		return nil, nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot convert capacity to int: %s", err)
	}
	return plaintext, remainder, nil
}

const cipherProgram = `
import sys
import regex2dfa
import fte.encoder
import binascii

regex = sys.argv[1]
encoder = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(regex), 512)

def encode(payload):
	ciphertext = bytearray(encoder.encode(payload))
	sys.stdout.write(binascii.hexlify(ciphertext))
	sys.stdout.write(":")
	sys.stdout.write(str(encoder.getCapacity()))
	sys.stdout.write("\n")
	sys.stdout.flush()

def decode(payload):
	[plaintext, remainder] = encoder.decode(payload)
	sys.stdout.write(binascii.hexlify(plaintext))
	sys.stdout.write(":")
	sys.stdout.write(binascii.hexlify(remainder))
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

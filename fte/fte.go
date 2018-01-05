package fte

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
)

const (
	COVERTEXT_HEADER_LEN_CIPHERTTEXT = 16
)

const (
	IV_LENGTH          = 7
	MSG_COUNTER_LENGTH = 8
	CTXT_EXPANSION     = 1 + IV_LENGTH + MSG_COUNTER_LENGTH + aes.BlockSize
)

type Cipher struct {
	mu sync.Mutex

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
	f, err := ioutil.TempFile("", "fte-")
	if err != nil {
		return err
	} else if err := f.Close(); err != nil {
		return err
	}
	c.filename = f.Name() + ".py"

	// Copy program to temporary path.
	if err := ioutil.WriteFile(c.filename, []byte(program), 0700); err != nil {
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
func (c *Cipher) Close() error {
	if c.cmd != nil {
		if err := c.stdin.Close(); err != nil {
			return err
		} else if err := c.cmd.Wait(); err != nil {
			return err
		}
		c.cmd = nil
	}
	return nil
}

func (c *Cipher) Capacity() int {
	// int(math.Floor(float64(fteEncoder.Capacity())/8.0)) - fte.COVERTEXT_HEADER_LEN_CIPHERTTEXT - fte.CTXT_EXPANSION
	return c.capacity
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

	if ciphertext, err = hex.DecodeString(string(bytes.TrimSpace(line))); err != nil {
		return nil, fmt.Errorf("fte.Cipher.Encrypt(): cannot decode result hex: %s", err)
	}
	return ciphertext, nil
}

// Decrypt decrypts ciphertext into plaintext.
func (c *Cipher) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := fmt.Fprintf(c.stdin, "D%09d%x", len(ciphertext), ciphertext); err != nil {
		return nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot write to stdin: %s", err)
	}

	line, err := c.bufout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot read from stdout: %s", err)
	}

	// Line is split into <plaintext,remainder>.
	segments := bytes.SplitN(line, []byte(" "), 2)

	if plaintext, err = hex.DecodeString(string(bytes.TrimSpace(segments[0]))); err != nil {
		return nil, fmt.Errorf("fte.Cipher.Decrypt(): cannot decode result hex: %s", err)
	}
	return plaintext, nil
}

const program = `
import sys
import regex2dfa
import fte.encoder
import binascii

regex = sys.argv[1]
encoder = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(regex), 512)

def encode(payload):
	ciphertext = bytearray(encoder.encode(payload))
	sys.stdout.write(binascii.hexlify(ciphertext))
	sys.stdout.write("\n")
	sys.stdout.flush()

def decode(payload):
	[plaintext, remainder] = encoder.decode(payload)
	sys.stdout.write(binascii.hexlify(plaintext))
	sys.stdout.write(" ")
	sys.stdout.write(binascii.hexlify(remainder))
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

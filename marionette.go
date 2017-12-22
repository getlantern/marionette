package marionette

import (
	"context"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/redjack/marionette/fte"
)

const (
	PartyClient = "client"
	PartyServer = "server"
)

// Rand returns a new PRNG seeded from the current time.
// This function can be overridden by the tests to provide a repeatable PRNG.
var Rand = func() *rand.Rand { return rand.New(rand.NewSource(time.Now().UnixNano())) }

// StripFormatVersion removes any version specified on a format.
func StripFormatVersion(format string) string {
	if i := strings.Index(format, ":"); i != -1 {
		return format[:i]
	}
	return format
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type bufConn struct {
	net.Conn

	mu  sync.RWMutex
	buf []byte
}

func newBufConn(conn net.Conn) *bufConn {
	c := &bufConn{Conn: conn}
	// TODO: Start goroutine to read into buffer.
	return c
}

// Peek returns the current buffer.
func (conn *bufConn) Peek() []byte {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	return conn.buf
}

// Unshift pushes data to the beginning of the buffer.
func (conn *bufConn) Unshift(data []byte) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.buf = append(data, conn.buf...)
}

// Read reads
func (conn *bufConn) Read(b []byte) (n int, err error) {
	// TODO: Copy buffer into b.
	// TODO: Shift bytes to beginning.
	// TODO: Return byte count.
	panic("TODO")
}

// NewCipherFunc returns a new instance of Cipher.
type NewCipherFunc func(regex string, n int) (Cipher, error)

// Cipher represents an interface for encrypting & decrypting messages.
type Cipher interface {
	Encrypt(plaintext []byte) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte) (plaintext []byte, err error)
	Capacity() int
}

// NewFTECipher returns a new instance of fte.Cipher.
func NewFTECipher(regex string, n int) (Cipher, error) {
	return fte.NewCipher(regex, n)
}

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

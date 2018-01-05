package marionette

import (
	"context"
	"math/rand"
	"net"
	"strings"
	"time"
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

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

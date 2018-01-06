package marionette

import (
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	PartyClient = "client"
	PartyServer = "server"
)

// Rand returns a new PRNG seeded from the current time.
// This function can be overridden by the tests to provide a repeatable PRNG.
var Rand = func() *rand.Rand { return rand.New(rand.NewSource(time.Now().UnixNano())) }

// Logger is the global marionette logger.
var Logger = zap.NewNop()

// StripFormatVersion removes any version specified on a format.
func StripFormatVersion(format string) string {
	if i := strings.Index(format, ":"); i != -1 {
		return format[:i]
	}
	return format
}

func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

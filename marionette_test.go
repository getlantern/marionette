package marionette_test

import (
	"math/rand"

	"github.com/redjack/marionette"
)

func init() {
	// Ensure all PRNGs are consistent for tests.
	marionette.Rand = NewRand
}

// NewRand returns a PRNG with a zero source.
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(0))
}

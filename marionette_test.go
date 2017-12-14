package marionette_test

import (
	"math/rand"
)

func NewRandZero() rand.Source {
	return zeroSource(0)
}

type zeroSource int

func (zeroSource) Int63() int64    { return 0 }
func (zeroSource) Seed(seed int64) {}

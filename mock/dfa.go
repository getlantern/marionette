package mock

import (
	"math/big"

	"github.com/redjack/marionette"
)

var _ marionette.DFA = (*DFA)(nil)

type DFA struct {
	CapacityFn        func() (int, error)
	RankFn            func(s string) (rank *big.Int, err error)
	UnrankFn          func(rank *big.Int) (ret string, err error)
	NumWordsInSliceFn func(n int) (numWords *big.Int, err error)
}

func (m *DFA) Capacity() (int, error) {
	return m.CapacityFn()
}

func (m *DFA) Rank(s string) (rank *big.Int, err error) {
	return m.RankFn(s)
}

func (m *DFA) Unrank(rank *big.Int) (ret string, err error) {
	return m.UnrankFn(rank)
}

func (m *DFA) NumWordsInSlice(n int) (numWords *big.Int, err error) {
	return m.NumWordsInSliceFn(n)
}

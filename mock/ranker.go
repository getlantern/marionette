package mock

import "math/big"

type Ranker struct {
	CapacityFn        func() int
	RankFn            func(s string) (rank *big.Int, err error)
	UnrankFn          func(rank *big.Int) (ret string, err error)
	NumWordsInSliceFn func(n int) (numWords int, err error)
}

func (m *Ranker) Capacity() int {
	return m.Capacity()
}

func (m *Ranker) Rank(s string) (rank *big.Int, err error) {
	return m.Rank(s)
}

func (m *Ranker) Unrank(rank *big.Int) (ret string, err error) {
	return m.Unrank(rank)
}

func (m *Ranker) NumWordsInSlice(n int) (numWords int, err error) {
	return m.NumWordsInSlice(n)
}

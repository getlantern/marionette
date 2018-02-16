package fte

import (
	"crypto/aes"
)

const (
	COVERTEXT_HEADER_LEN_CIPHERTTEXT = 16
)

const (
	IV_LENGTH          = 7
	MSG_COUNTER_LENGTH = 8
	CTXT_EXPANSION     = 1 + IV_LENGTH + MSG_COUNTER_LENGTH + aes.BlockSize
)

// Cache represents a cache of Ciphers & DFAs.
type Cache struct {
	ciphers map[string]*Cipher
	dfas    map[dfaKey]*DFA
}

// NewCache returns a new instance of Cache.
func NewCache() *Cache {
	return &Cache{
		ciphers: make(map[string]*Cipher),
		dfas:    make(map[dfaKey]*DFA),
	}
}

// Close close and removes all ciphers & dfas.
func (c *Cache) Close() (err error) {
	for _, cipher := range c.ciphers {
		if e := cipher.Close(); e != nil && err == nil {
			err = e
		}
	}
	c.ciphers = nil

	for _, dfa := range c.dfas {
		if e := dfa.Close(); e != nil && err == nil {
			err = e
		}
	}
	c.dfas = nil

	return err
}

// Cipher returns a instance of Cipher associated with regex & n.
// Creates a new cipher if one doesn't already exist.
func (c *Cache) Cipher(regex string) *Cipher {
	cipher := c.ciphers[regex]
	if cipher == nil {
		cipher = NewCipher(regex)
		c.ciphers[regex] = cipher
	}
	return cipher
}

// DFA returns a instance of DFA associated with regex & n.
// Creates a new DFA if one doesn't already exist.
func (c *Cache) DFA(regex string, n int) *DFA {
	dfa := c.dfas[dfaKey{regex, n}]
	if dfa == nil {
		dfa = NewDFA(regex, n)
		c.dfas[dfaKey{regex, n}] = dfa
	}
	return dfa
}

type dfaKey struct {
	regex string
	n     int
}

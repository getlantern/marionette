package tg

import (
	"strings"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/fte"
)

type FTECipher struct {
	key         string
	regex       string
	useCapacity bool
}

func NewFTECipher(key, regex string, msg_len int, useCapacity bool) *FTECipher {
	return &FTECipher{
		key:         key,
		regex:       regex,
		useCapacity: useCapacity,
	}
}

func (c *FTECipher) Key() string {
	return c.key
}

func (c *FTECipher) Capacity(fsm marionette.FSM) (int, error) {
	if !c.useCapacity && strings.HasSuffix(c.regex, ".+") {
		return (1 << 18), nil
	}
	capacity, err := fsm.Cipher(c.regex).Capacity()
	if err != nil {
		return 0, err
	}
	return capacity - fte.COVERTEXT_HEADER_LEN_CIPHERTTEXT - fte.CTXT_EXPANSION, nil
}

func (c *FTECipher) Encrypt(fsm marionette.FSM, template string, data []byte) (ciphertext []byte, err error) {
	return fsm.Cipher(c.regex).Encrypt(data)
}

func (c *FTECipher) Decrypt(fsm marionette.FSM, ciphertext []byte) (plaintext []byte, err error) {
	plaintext, _, err = fsm.Cipher(c.regex).Decrypt(ciphertext)
	return plaintext, err
}

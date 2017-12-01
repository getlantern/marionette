package fte

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"os/exec"
)

const (
	COVERTEXT_HEADER_LEN_CIPHERTTEXT = 16
)

const (
	IV_LENGTH          = 7
	MSG_COUNTER_LENGTH = 8
	CTXT_EXPANSION     = 1 + IV_LENGTH + MSG_COUNTER_LENGTH + aes.BlockSize
)

const FixedSlice = 512

type Encoder struct {
	capacity int
}

func NewEncoder(regex string, msgLen int) *Encoder {
	return &Encoder{}
}

func (enc *Encoder) Capacity() int { return enc.capacity }

func (enc *Encoder) Encode(plaintext []byte) ([]byte, error) {
	panic("TODO")
}

// Encode encodes plaintext using the specified regex and returns the ciphertext.
func Encode(regex, plaintext string) (string, error) {
	cmd := exec.Command("python",
		"-c", fmt.Sprintf(`import regex2dfa; import fte.encoder; print fte.encoder.DfaEncoder(regex2dfa.regex2dfa(%q), %d).encode(%q)`, regex, FixedSlice, plaintext),
	)
	out, err := cmd.Output()
	return string(bytes.TrimSpace(out)), err
}

type Decoder struct{}

// Decode decodes ciphertext using the specified regex and returns the plaintext.
func Decode(regex, ciphertext string) (string, error) {
	cmd := exec.Command("python",
		"-c", fmt.Sprintf(`import regex2dfa; import fte.encoder; [plaintext, remainder] = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(%q), %d).decode(%q); print plaintext`, regex, FixedSlice, ciphertext),
	)
	out, err := cmd.Output()
	return string(bytes.TrimSpace(out)), err
}

package fte

import (
	"bytes"
	"fmt"
	"os/exec"
)

const FixedSlice = 512

// Encode encodes plaintext using the specified regex and returns the ciphertext.
func Encode(regex, plaintext string) (string, error) {
	cmd := exec.Command("python",
		"-c", fmt.Sprintf(`import regex2dfa; import fte.encoder; print fte.encoder.DfaEncoder(regex2dfa.regex2dfa(%q), %d).encode(%q)`, regex, FixedSlice, plaintext),
	)
	out, err := cmd.Output()
	return string(bytes.TrimSpace(out)), err
}

// Decode decodes ciphertext using the specified regex and returns the plaintext.
func Decode(regex, ciphertext string) (string, error) {
	cmd := exec.Command("python",
		"-c", fmt.Sprintf(`import regex2dfa; import fte.encoder; [plaintext, remainder] = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(%q), %d).decode(%q); print plaintext`, regex, FixedSlice, ciphertext),
	)
	out, err := cmd.Output()
	return string(bytes.TrimSpace(out)), err
}

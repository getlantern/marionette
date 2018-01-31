package fte_test

import (
	"strings"
	"testing"

	"github.com/redjack/marionette/fte"
)

func TestDFA(t *testing.T) {
	dfa := fte.NewDFA(`[a-zA-Z0-9\?\-\.\&]+`, 2048)
	if err := dfa.Open(); err != nil {
		t.Fatal(err)
	}
	defer dfa.Close()

	msg0 := strings.Repeat("A", 2048)
	msg1 := strings.Repeat("B", 2048)

	// Encode/decode first message.
	if ciphertext, err := dfa.Encrypt([]byte(msg0)); err != nil {
		t.Fatal(err)
	} else if plaintext, err := dfa.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != msg0 {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	}

	// Encode/decode second message.
	if ciphertext, err := dfa.Encrypt([]byte(msg1)); err != nil {
		t.Fatal(err)
	} else if plaintext, err := dfa.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `foo bar` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	}

	if err := dfa.Close(); err != nil {
		t.Fatal(err)
	}
}

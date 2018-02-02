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
	if rank, err := dfa.Rank(msg0); err != nil {
		t.Fatal(err)
	} else if other, err := dfa.Unrank(rank); err != nil {
		t.Fatal(err)
	} else if other != msg0 {
		t.Fatalf("unexpected unrank: %q", other)
	}

	// Encode/decode second message.
	if rank, err := dfa.Rank(msg1); err != nil {
		t.Fatal(err)
	} else if other, err := dfa.Unrank(rank); err != nil {
		t.Fatal(err)
	} else if other != msg1 {
		t.Fatalf("unexpected unrank: %q", other)
	}

	if err := dfa.Close(); err != nil {
		t.Fatal(err)
	}
}

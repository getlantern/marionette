package regex2dfa_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/regex2dfa"
)

func TestRegex2DFA(t *testing.T) {
	for i := 1; i <= 8; i++ {
		name := fmt.Sprintf("test%d", i)

		regex, err := ioutil.ReadFile(`testdata/` + name + `.regex`)
		if err != nil {
			t.Fatal(name, err)
		}

		exp, err := ioutil.ReadFile(`testdata/` + name + `.dfa`)
		if err != nil {
			t.Fatal(name, err)
		}

		dfa, err := regex2dfa.Regex2DFA(string(regex))
		if err != nil {
			t.Fatal(err)
		} else if diff := cmp.Diff(strings.TrimSpace(string(exp)), strings.TrimSpace(dfa)); diff != "" {
			t.Fatal(name, diff)
		}
	}
}

func TestRegex2DFA_HTTP(t *testing.T) {
	if _, err := regex2dfa.Regex2DFA("^HTTP/1\\.1\\ 200 OK\r\nContent-Type:\\ ([a-zA-Z0-9]+)\r\n\r\n\\C{64}$"); err != nil {
		t.Fatal(err)
	}
}

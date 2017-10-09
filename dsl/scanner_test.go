package dsl_test

import (
	"testing"

	"github.com/redjack/marionette/dsl"
)

func TestScanner_Scan(t *testing.T) {
	t.Run("EOF", func(t *testing.T) {
		if tok, lit, pos := Scan(""); tok != dsl.EOF {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("WS", func(t *testing.T) {
		if tok, lit, pos := Scan(" \n \t "); tok != dsl.WS {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != " \n \t " {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("IDENT", func(t *testing.T) {
		if tok, lit, pos := Scan("iDent_123"); tok != dsl.IDENT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `iDent_123` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("NULL", func(t *testing.T) {
		if tok, lit, pos := Scan("NULL"); tok != dsl.NULL {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `NULL` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("STRING", func(t *testing.T) {
		t.Run("SingleQuote", func(t *testing.T) {
			if tok, lit, pos := Scan(`'test string'`); tok != dsl.STRING {
				t.Fatalf("unexpected token: %s", tok.String())
			} else if lit != `test string` {
				t.Fatalf("unexpected literal: %s", lit)
			} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
				t.Fatalf("unexpected pos: %#v", pos)
			}
		})

		t.Run("DoubleQuote", func(t *testing.T) {
			if tok, lit, pos := Scan(`"test string"`); tok != dsl.STRING {
				t.Fatalf("unexpected token: %s", tok.String())
			} else if lit != `test string` {
				t.Fatalf("unexpected literal: %s", lit)
			} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
				t.Fatalf("unexpected pos: %#v", pos)
			}
		})
	})

	t.Run("INTEGER", func(t *testing.T) {
		if tok, lit, pos := Scan("12376"); tok != dsl.INTEGER {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `12376` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("FLOAT", func(t *testing.T) {
		if tok, lit, pos := Scan("12376.21387"); tok != dsl.FLOAT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `12376.21387` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("LPAREN", func(t *testing.T) {
		if tok, lit, pos := Scan("("); tok != dsl.LPAREN {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `(` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("RPAREN", func(t *testing.T) {
		if tok, lit, pos := Scan(")"); tok != dsl.RPAREN {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `)` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("DOT", func(t *testing.T) {
		if tok, lit, pos := Scan("."); tok != dsl.DOT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `.` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("COMMA", func(t *testing.T) {
		if tok, lit, pos := Scan(","); tok != dsl.COMMA {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `,` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("COLON", func(t *testing.T) {
		if tok, lit, pos := Scan(":"); tok != dsl.COLON {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `:` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("ACTION", func(t *testing.T) {
		if tok, lit, pos := Scan("action"); tok != dsl.ACTION {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `action` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("CLIENT", func(t *testing.T) {
		if tok, lit, pos := Scan("client"); tok != dsl.CLIENT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `client` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("IF", func(t *testing.T) {
		if tok, lit, pos := Scan("if"); tok != dsl.IF {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `if` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("END", func(t *testing.T) {
		if tok, lit, pos := Scan("end"); tok != dsl.END {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `end` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("REGEX_MATCH_INCOMING", func(t *testing.T) {
		if tok, lit, pos := Scan("regex_match_incoming"); tok != dsl.REGEX_MATCH_INCOMING {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `regex_match_incoming` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("SERVER", func(t *testing.T) {
		if tok, lit, pos := Scan("server"); tok != dsl.SERVER {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `server` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("START", func(t *testing.T) {
		if tok, lit, pos := Scan("start"); tok != dsl.START {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `start` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (dsl.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})
}

func Scan(data string) (tok dsl.Token, lit string, pos dsl.Pos) {
	return dsl.NewScanner([]byte(data)).Scan()
}

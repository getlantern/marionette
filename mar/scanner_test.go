package mar_test

import (
	"testing"

	"github.com/redjack/marionette/mar"
)

func TestScanner_Scan(t *testing.T) {
	t.Run("EOF", func(t *testing.T) {
		if tok, lit, pos := Scan(""); tok != mar.EOF {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("WS", func(t *testing.T) {
		if tok, lit, pos := Scan(" \n \t "); tok != mar.WS {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != " \n \t " {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("IDENT", func(t *testing.T) {
		if tok, lit, pos := Scan("iDent_123"); tok != mar.IDENT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `iDent_123` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("NULL", func(t *testing.T) {
		if tok, lit, pos := Scan("NULL"); tok != mar.NULL {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `NULL` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("STRING", func(t *testing.T) {
		t.Run("SingleQuote", func(t *testing.T) {
			if tok, lit, pos := Scan(`'test string'`); tok != mar.STRING {
				t.Fatalf("unexpected token: %s", tok.String())
			} else if lit != `test string` {
				t.Fatalf("unexpected literal: %s", lit)
			} else if pos != (mar.Pos{Line: 0, Char: 0}) {
				t.Fatalf("unexpected pos: %#v", pos)
			}
		})

		t.Run("DoubleQuote", func(t *testing.T) {
			if tok, lit, pos := Scan(`"test string"`); tok != mar.STRING {
				t.Fatalf("unexpected token: %s", tok.String())
			} else if lit != `test string` {
				t.Fatalf("unexpected literal: %s", lit)
			} else if pos != (mar.Pos{Line: 0, Char: 0}) {
				t.Fatalf("unexpected pos: %#v", pos)
			}
		})
	})

	t.Run("INTEGER", func(t *testing.T) {
		if tok, lit, pos := Scan("12376"); tok != mar.INTEGER {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `12376` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("FLOAT", func(t *testing.T) {
		if tok, lit, pos := Scan("12376.21387"); tok != mar.FLOAT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `12376.21387` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("LPAREN", func(t *testing.T) {
		if tok, lit, pos := Scan("("); tok != mar.LPAREN {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `(` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("RPAREN", func(t *testing.T) {
		if tok, lit, pos := Scan(")"); tok != mar.RPAREN {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `)` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("DOT", func(t *testing.T) {
		if tok, lit, pos := Scan("."); tok != mar.DOT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `.` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("COMMA", func(t *testing.T) {
		if tok, lit, pos := Scan(","); tok != mar.COMMA {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `,` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("COLON", func(t *testing.T) {
		if tok, lit, pos := Scan(":"); tok != mar.COLON {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `:` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("ACTION", func(t *testing.T) {
		if tok, lit, pos := Scan("action"); tok != mar.ACTION {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `action` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("CLIENT", func(t *testing.T) {
		if tok, lit, pos := Scan("client"); tok != mar.CLIENT {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `client` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("IF", func(t *testing.T) {
		if tok, lit, pos := Scan("if"); tok != mar.IF {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `if` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("END", func(t *testing.T) {
		if tok, lit, pos := Scan("end"); tok != mar.END {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `end` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("REGEX_MATCH_INCOMING", func(t *testing.T) {
		if tok, lit, pos := Scan("regex_match_incoming"); tok != mar.REGEX_MATCH_INCOMING {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `regex_match_incoming` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("SERVER", func(t *testing.T) {
		if tok, lit, pos := Scan("server"); tok != mar.SERVER {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `server` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})

	t.Run("START", func(t *testing.T) {
		if tok, lit, pos := Scan("start"); tok != mar.START {
			t.Fatalf("unexpected token: %s", tok.String())
		} else if lit != `start` {
			t.Fatalf("unexpected literal: %s", lit)
		} else if pos != (mar.Pos{Line: 0, Char: 0}) {
			t.Fatalf("unexpected pos: %#v", pos)
		}
	})
}

func Scan(data string) (tok mar.Token, lit string, pos mar.Pos) {
	return mar.NewScanner([]byte(data)).Scan()
}

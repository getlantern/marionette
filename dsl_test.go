package marionette_test

import (
	"testing"

	"github.com/redjack/marionette-go"
)

func TestScanner_Scan(t *testing.T) {
	t.Run("EOF", func(t *testing.T) {
		if tok, err := Scan(``); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenTypeEOF}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})

	t.Run("Whitespace", func(t *testing.T) {
		t.Run("Space", func(t *testing.T) {
			if tok, err := Scan(`   `); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenTypeWhitespace, Value: `   `}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})

		t.Run("LineFeed", func(t *testing.T) {
			if tok, err := Scan("\n"); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenTypeWhitespace, Value: " \n"}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})

		t.Run("FormFeed", func(t *testing.T) {
			if tok, err := Scan("\f"); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenTypeWhitespace, Value: " \n"}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})

		t.Run("CarriageReturn", func(t *testing.T) {
			if tok, err := Scan("\r"); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenTypeWhitespace, Value: " \n"}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			if tok, err := Scan(`""`); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenString, Value: ``}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})
		t.Run("Simple", func(t *testing.T) {
			if tok, err := Scan(`"hello world"`); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenString, Value: `hello world`}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})
		t.Run("Escape", func(t *testing.T) {
			if tok, err := Scan("'foo\\\nbar'"); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenString, Value: "foo\nbar"}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})
		t.Run("Unicode", func(t *testing.T) {
			if tok, err := Scan(`'frosty the \2603'`); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenString, Value: `frosty the ☃`}) {
				t.Fatalf("unexpected token: %#v", tok)
			}
		})
	})

	t.Run("Integer", func(t *testing.T) {
		if tok, err := Scan(`100`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenInteger, Value: `0`, Number: 0.0}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})
	t.Run("Float", func(t *testing.T) {
		if tok, err := Scan(`1.123`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.TokenFloat, Value: `1.123`, Number: 1.123}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})

	t.Run("Ident", func(t *testing.T) {
		if tok, err := Scan(`myIdent`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.IdentToken, Value: `myIdent`}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})

	t.Run("Comma", func(t *testing.T) {
		if tok, err := Scan(`,`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.CommaToken}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})
	t.Run("Colon", func(t *testing.T) {
		if tok, err := Scan(`:`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.ColonToken}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})
	t.Run("Semicolon", func(t *testing.T) {
		if tok, err := Scan(`,`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.SemicolonToken}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})
	t.Run("LParen", func(t *testing.T) {
		if tok, err := Scan(`(`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.LParenToken}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})
	t.Run("RParen", func(t *testing.T) {
		if tok, err := Scan(`)`); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(tok, marionette.Token{Type: marionette.RParenToken}) {
			t.Fatalf("unexpected token: %#v", tok)
		}
	})
}

func TestParser_Parse(t *testing.T) {
	t.Run("Test1", func(t *testing.T) {
		if doc, err := Parse(`
			connection(tcp, 80):
			  start      downstream NULL     1.0
			  downstream upstream   http_get 1.0
			  upstream   end        http_ok  1.0

			action http_get:
			  client fte.send("^regex\r\n\r\n$", 128)

			action http_ok:
			  server fte.send("^regex\r\n\r\n\C*$", 128)`,
		); !reflect.DeepEqual(doc, &Doc{}) {
			t.Fatalf("unexpected value: %#v", doc)
		}
	})
}

/*
    def test1(self):
        mar_format = """connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_ok:
          server fte.send("^regex\r\n\r\n\C*$", 128)"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 80)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_dst(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(), "http_get")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[2].get_src(), "upstream")
        self.assertEquals(parsed_format.get_transitions()[2].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_action_block(), "http_ok")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_args(), [
                "^regex\r\n\r\n$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_name(), "http_ok")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_args(), [
                "^regex\r\n\r\n\C*$", 128])

    def test2(self):
        mar_format = """connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_ok:
          server fte.send("^regex\r\n\r\n\C*$", 128)

        action http_put:
          client fte.send("^regex\r\n\r\n$", 128)"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 80)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_dst(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(), "http_get")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[2].get_src(), "upstream")
        self.assertEquals(parsed_format.get_transitions()[2].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_action_block(), "http_ok")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_args(), [
                "^regex\r\n\r\n$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_name(), "http_ok")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_args(), [
                "^regex\r\n\r\n\C*$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_name(), "http_put")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_args(), [
                "^regex\r\n\r\n$", 128])

    def test3(self):
        mar_format = """connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_ok:
          server fte.send("^regex\r\n\r\n\C*$", 128)

        action http_put:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_notok:
          server fte.send("^regex\r\n\r\n\C*$", 128)"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 80)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_dst(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(), "http_get")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[2].get_src(), "upstream")
        self.assertEquals(parsed_format.get_transitions()[2].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_action_block(), "http_ok")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_args(), [
                "^regex\r\n\r\n$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_name(), "http_ok")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_args(), [
                "^regex\r\n\r\n\C*$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_name(), "http_put")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_args(), [
                "^regex\r\n\r\n$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[3].get_name(), "http_notok")
        self.assertEquals(
            parsed_format.get_action_blocks()[3].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[3].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[3].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[3].get_args(), [
                "^regex\r\n\r\n\C*$", 128])

    def test4(self):
        mar_format = """connection(tcp, 8082):
  start      handshake  NULL               1.0
  handshake  upstream   upstream_handshake 1.0
  upstream   downstream upstream_async     1.0
  downstream upstream   downstream_async   1.0

action upstream_handshake:
  client fte.send("^.*$", 128)

action upstream_async:
  client fte.send_async("^.*$", 128)

action downstream_async:
  server fte.send_async("^.*$", 128)
"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 8082)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "handshake")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "handshake")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_dst(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(),
            "upstream_handshake")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[2].get_src(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_dst(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_action_block(),
            "upstream_async")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[3].get_src(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[3].get_dst(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[3].get_action_block(),
            "downstream_async")
        self.assertEquals(
            parsed_format.get_transitions()[3].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_name(),
            "upstream_handshake")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_args(), ["^.*$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_name(), "upstream_async")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_method(), "send_async")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_args(), ["^.*$", 128])

        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_name(),
            "downstream_async")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_method(), "send_async")
        self.assertEquals(
            parsed_format.get_action_blocks()[2].get_args(), ["^.*$", 128])

    def test5(self):
        mar_format = """connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$")

        action http_ok:
          server fte.send("^regex\r\n\r\n\C*$")"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 80)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "downstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_dst(), "upstream")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(), "http_get")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[2].get_src(), "upstream")
        self.assertEquals(parsed_format.get_transitions()[2].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_action_block(), "http_ok")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_probability(), float(1.0))

    def test6(self):
        mar_format = """connection(tcp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0

        action http_get:
          client fte.send("^regex1\r\n\r\n$")
          server fte.recv("^regex2\r\n\r\n$")"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 80)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "do_nothing")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "do_nothing")
        self.assertEquals(parsed_format.get_transitions()[1].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))

        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_args(),
            ["^regex1\r\n\r\n$"])

        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_method(), "recv")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_args(),
            ["^regex2\r\n\r\n$"])


    def test7(self):
        mar_format = """connection(tcp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0
          start do_err NULL error
          do_err end NULL error

        action http_get:
          client fte.send("^regex1\r\n\r\n$")"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "tcp")
        self.assertEquals(parsed_format.get_port(), 80)

        self.assertEquals(
            parsed_format.get_transitions()[0].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_dst(), "do_nothing")
        self.assertEquals(
            parsed_format.get_transitions()[0].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[0].get_probability(), float(1.0))
        self.assertEquals(
            parsed_format.get_transitions()[0].is_error_transition(), False)

        self.assertEquals(
            parsed_format.get_transitions()[1].get_src(), "do_nothing")
        self.assertEquals(parsed_format.get_transitions()[1].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[1].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[1].get_probability(), float(1.0))
        self.assertEquals(
            parsed_format.get_transitions()[1].is_error_transition(), False)

        self.assertEquals(
            parsed_format.get_transitions()[2].get_src(), "start")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_dst(), "do_err")
        self.assertEquals(
            parsed_format.get_transitions()[2].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[2].get_probability(), float(0))
        self.assertEquals(
            parsed_format.get_transitions()[2].is_error_transition(), True)

        self.assertEquals(
            parsed_format.get_transitions()[3].get_src(), "do_err")
        self.assertEquals(
            parsed_format.get_transitions()[3].get_dst(), "end")
        self.assertEquals(
            parsed_format.get_transitions()[3].get_action_block(), None)
        self.assertEquals(
            parsed_format.get_transitions()[3].get_probability(), float(0))
        self.assertEquals(
            parsed_format.get_transitions()[3].is_error_transition(), True)


    def test8(self):
        mar_format = """connection(tcp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0

        action http_get:
          client fte.send("^regex1\r\n\r\n$")
          server fte.recv("^regex2\r\n\r\n$") if regex_match_incoming("^regex2.*")"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_method(), "send")
        self.assertEquals(
            parsed_format.get_action_blocks()[0].get_args(),
            ["^regex1\r\n\r\n$"])

        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_name(), "http_get")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_party(), "server")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_module(), "fte")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_method(), "recv")
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_args(),
            ["^regex2\r\n\r\n$"])
        self.assertEquals(
            parsed_format.get_action_blocks()[1].get_regex_match_incoming(),"^regex2.*")

    def test9(self):
        mar_format = """connection(udp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0

        action http_get:
          client fte.send("^regex1\r\n\r\n$")"""

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_transport(), "udp")

    def test_hex_input_strings(self):
        mar_files = marionette_tg.dsl.find_mar_files('client',
                                                     'test_hex_input_strings',
                                                     '20150701')
        with open(mar_files[0]) as f:
            mar_format = f.read()

        parsed_format = marionette_tg.dsl.parse(mar_format)

        self.assertEquals(parsed_format.get_action_blocks()[0].get_name(), "null_puts")
        self.assertEquals(parsed_format.get_action_blocks()[0].get_party(), "client")
        self.assertEquals(parsed_format.get_action_blocks()[0].get_module(), "io")
        self.assertEquals(parsed_format.get_action_blocks()[0].get_method(), "puts")
        self.assertEquals(parsed_format.get_action_blocks()[0].get_args()[0], "\x41\x42\\backslash")


if __name__ == "__main__":
    unittest.main()
*/

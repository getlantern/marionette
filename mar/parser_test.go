package mar_test

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/redjack/marionette/dsl"
)

func TestParser_Parse(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_ok",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_ok:
          server fte.send("^regex\r\n\r\n\\C*$", 128)
		`)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test2", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_ok",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_put",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_ok:
          server fte.send("^regex\r\n\r\n\\C*$", 128)

        action http_put:
          client fte.send("^regex\r\n\r\n$", 128)
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test3", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_ok",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_put",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_notok",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_ok:
          server fte.send("^regex\r\n\r\n\\C*$", 128)

        action http_put:
          client fte.send("^regex\r\n\r\n$", 128)

        action http_notok:
          server fte.send("^regex\r\n\r\n\\C*$", 128)
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test4", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "8082",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "handshake",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "handshake",
						Destination: "upstream",
						ActionBlock: "upstream_handshake",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "upstream",
						Destination: "downstream",
						ActionBlock: "upstream_async",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "downstream_async",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "upstream_handshake",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^.*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "upstream_async",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send_async",
							Args: []*dsl.Arg{
								{Value: "^.*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "downstream_async",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send_async",
							Args: []*dsl.Arg{
								{Value: "^.*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 8082):
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
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test5", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n$"},
							},
							Regex: "",
						},
					},
				},
				&dsl.ActionBlock{
					Name: "http_ok",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start      downstream NULL     1.0
          downstream upstream   http_get 1.0
          upstream   end        http_ok  1.0

        action http_get:
          client fte.send("^regex\r\n\r\n$")

        action http_ok:
          server fte.send("^regex\r\n\r\n\\C*$")
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test6", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex1\r\n\r\n$"},
							},
							Regex: "",
						},
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "recv",
							Args: []*dsl.Arg{
								{Value: "^regex2\r\n\r\n$"},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0

        action http_get:
          client fte.send("^regex1\r\n\r\n$")
          server fte.recv("^regex2\r\n\r\n$")
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test7", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:            "start",
						Destination:       "do_err",
						ActionBlock:       "NULL",
						IsErrorTransition: true,
					},
					&dsl.Transition{
						Source:            "do_err",
						Destination:       "end",
						ActionBlock:       "NULL",
						IsErrorTransition: true,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex1\r\n\r\n$"},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0
          start do_err NULL error
          do_err end NULL error

        action http_get:
          client fte.send("^regex1\r\n\r\n$")
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test8", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex1\r\n\r\n$"},
							},
							Regex: "",
						},
						&dsl.Action{
							Party:  "server",
							Name:   "fte",
							Method: "recv",
							Args: []*dsl.Arg{
								{Value: "^regex2\r\n\r\n$"},
							},
							Regex: "^regex2.*",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0

        action http_get:
          client fte.send("^regex1\r\n\r\n$")
          server fte.recv("^regex2\r\n\r\n$") if regex_match_incoming("^regex2.*")
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("test9", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "udp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "http_get",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*dsl.Arg{
								{Value: "^regex1\r\n\r\n$"},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(udp, 80):
          start do_nothing NULL 1.0
          do_nothing end NULL 1.0

        action http_get:
          client fte.send("^regex1\r\n\r\n$")
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})

	t.Run("hex_input_strings", func(t *testing.T) {
		exp := &dsl.Document{
			Model: &dsl.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*dsl.Transition{
					&dsl.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&dsl.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*dsl.ActionBlock{
				&dsl.ActionBlock{
					Name: "null_puts",
					Actions: []*dsl.Action{
						&dsl.Action{
							Party:  "client",
							Name:   "io",
							Method: "puts",
							Args: []*dsl.Arg{
								{Value: "\x41\x42\\backslash"},
							},
							Regex: "",
						},
					},
				},
			},
		}

		doc, err := Parse(`connection(tcp, 80):
        start do_nothing NULL 1.0
        do_nothing end NULL 1.0
        action null_puts:
        client io.puts('\x41\x42\\backslash')
        `)
		if err != nil {
			t.Fatal(err)
		} else if StripPos(doc); !reflect.DeepEqual(doc, exp) {
			t.Fatalf("document mismatch:\n\ngot:%s\n\nexp:%s", spew.Sprintf("%#v", doc), spew.Sprintf("%#v", exp))
		}
	})
}

func Parse(data string) (*dsl.Document, error) {
	return dsl.NewParser().Parse([]byte(data))
}

// StripPos removes all position data from a node and its descendents.
func StripPos(node dsl.Node) {
	dsl.Walk(dsl.VisitorFunc(func(node dsl.Node) {
		switch node := node.(type) {
		case *dsl.Model:
			node.Connection = dsl.Pos{}
			node.Lparen = dsl.Pos{}
			node.TransportPos = dsl.Pos{}
			node.Comma = dsl.Pos{}
			node.PortPos = dsl.Pos{}
			node.Rparen = dsl.Pos{}
			node.Colon = dsl.Pos{}

		case *dsl.Transition:
			node.SourcePos = dsl.Pos{}
			node.DestinationPos = dsl.Pos{}
			node.ActionBlockPos = dsl.Pos{}
			node.ProbabilityPos = dsl.Pos{}

		case *dsl.ActionBlock:
			node.Action = dsl.Pos{}
			node.NamePos = dsl.Pos{}
			node.Colon = dsl.Pos{}

		case *dsl.Action:
			node.PartyPos = dsl.Pos{}
			node.NamePos = dsl.Pos{}
			node.Dot = dsl.Pos{}
			node.MethodPos = dsl.Pos{}
			node.Lparen = dsl.Pos{}
			node.Rparen = dsl.Pos{}
			node.If = dsl.Pos{}
			node.RegexMatchIncoming = dsl.Pos{}
			node.RegexMatchIncomingLparen = dsl.Pos{}
			node.RegexPos = dsl.Pos{}
			node.RegexMatchIncomingRparen = dsl.Pos{}

		case *dsl.Arg:
			node.Pos = dsl.Pos{}
			node.EndPos = dsl.Pos{}
		}
	}), node)
}

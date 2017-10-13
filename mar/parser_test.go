package mar_test

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/redjack/marionette/mar"
)

func TestParser_Parse(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_ok",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_ok",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_put",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_ok",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n\\C*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_put",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_notok",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "8082",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "handshake",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "handshake",
						Destination: "upstream",
						ActionBlock: "upstream_handshake",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "upstream",
						Destination: "downstream",
						ActionBlock: "upstream_async",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "downstream_async",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "upstream_handshake",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^.*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "upstream_async",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send_async",
							Args: []*mar.Arg{
								{Value: "^.*$"},
								{Value: 128},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "downstream_async",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send_async",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "downstream",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "downstream",
						Destination: "upstream",
						ActionBlock: "http_get",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "upstream",
						Destination: "end",
						ActionBlock: "http_ok",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex\r\n\r\n$"},
							},
							Regex: "",
						},
					},
				},
				&mar.ActionBlock{
					Name: "http_ok",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex1\r\n\r\n$"},
							},
							Regex: "",
						},
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "recv",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:            "start",
						Destination:       "do_err",
						ActionBlock:       "NULL",
						IsErrorTransition: true,
					},
					&mar.Transition{
						Source:            "do_err",
						Destination:       "end",
						ActionBlock:       "NULL",
						IsErrorTransition: true,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
								{Value: "^regex1\r\n\r\n$"},
							},
							Regex: "",
						},
						&mar.Action{
							Party:  "server",
							Name:   "fte",
							Method: "recv",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "udp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "http_get",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "fte",
							Method: "send",
							Args: []*mar.Arg{
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
		exp := &mar.Document{
			Model: &mar.Model{
				Transport: "tcp",
				Port:      "80",
				Transitions: []*mar.Transition{
					&mar.Transition{
						Source:      "start",
						Destination: "do_nothing",
						ActionBlock: "NULL",
						Probability: 1,
					},
					&mar.Transition{
						Source:      "do_nothing",
						Destination: "end",
						ActionBlock: "NULL",
						Probability: 1,
					},
				},
			},
			ActionBlocks: []*mar.ActionBlock{
				&mar.ActionBlock{
					Name: "null_puts",
					Actions: []*mar.Action{
						&mar.Action{
							Party:  "client",
							Name:   "io",
							Method: "puts",
							Args: []*mar.Arg{
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

func Parse(data string) (*mar.Document, error) {
	return mar.NewParser().Parse([]byte(data))
}

// StripPos removes all position data from a node and its descendents.
func StripPos(node mar.Node) {
	mar.Walk(mar.VisitorFunc(func(node mar.Node) {
		switch node := node.(type) {
		case *mar.Model:
			node.Connection = mar.Pos{}
			node.Lparen = mar.Pos{}
			node.TransportPos = mar.Pos{}
			node.Comma = mar.Pos{}
			node.PortPos = mar.Pos{}
			node.Rparen = mar.Pos{}
			node.Colon = mar.Pos{}

		case *mar.Transition:
			node.SourcePos = mar.Pos{}
			node.DestinationPos = mar.Pos{}
			node.ActionBlockPos = mar.Pos{}
			node.ProbabilityPos = mar.Pos{}

		case *mar.ActionBlock:
			node.Action = mar.Pos{}
			node.NamePos = mar.Pos{}
			node.Colon = mar.Pos{}

		case *mar.Action:
			node.PartyPos = mar.Pos{}
			node.NamePos = mar.Pos{}
			node.Dot = mar.Pos{}
			node.MethodPos = mar.Pos{}
			node.Lparen = mar.Pos{}
			node.Rparen = mar.Pos{}
			node.If = mar.Pos{}
			node.RegexMatchIncoming = mar.Pos{}
			node.RegexMatchIncomingLparen = mar.Pos{}
			node.RegexPos = mar.Pos{}
			node.RegexMatchIncomingRparen = mar.Pos{}

		case *mar.Arg:
			node.Pos = mar.Pos{}
			node.EndPos = mar.Pos{}
		}
	}), node)
}

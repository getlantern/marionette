package marionette_test

import (
	"io"
	"testing"
	// "time"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
)

func TestIntegration(t *testing.T) {
	t.Run("fte/async", func(t *testing.T) {
		RunIntegration(t, []byte(`
			connection(tcp, 8082):
			  start      upstream  NULL          1.0
			  upstream   downstream upstream     1.0
			  downstream upstream   downstream   1.0

			action upstream:
			  client fte.send_async("^.*$", 128)

			action downstream:
			  server fte.send_async("^.*$", 128)
			`),
		)
	})

	t.Run("fte/sync", func(t *testing.T) {
		RunIntegration(t, []byte(`
			connection(tcp, 8082):
			  start      upstream  NULL          1.0
			  upstream   downstream upstream     1.0
			  downstream upstream   downstream   1.0

			action upstream:
			  client fte.send("^.*$", 128)

			action downstream:
			  server fte.send("^.*$", 128)
			`),
		)
	})
}

func RunIntegration(t *testing.T, program []byte) {
	// Open listener.
	ln, err := marionette.Listen(mar.MustParse(marionette.PartyServer, program), "")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Connect using the dialer.
	dialer, err := marionette.NewDialer(mar.MustParse(marionette.PartyClient, program), "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	defer dialer.Close()

	// Open a stream and write.
	input, err := dialer.Dial()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()

	if n, err := input.Write([]byte("foo")); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatalf("unexpected n: %d", n)
	}

	// Accept stream from server-side.
	output, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	// Read data from server-side.
	buf := make([]byte, 3)
	if _, err := io.ReadFull(output, buf); err != nil {
		t.Fatal(err)
	} else if string(buf) != `foo` {
		t.Fatalf("unexpected read: %s", buf)
	}
}

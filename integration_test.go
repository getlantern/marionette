package marionette_test

import (
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
)

func TestIntegration(t *testing.T) {
	input := []byte(`
connection(tcp, 8082):
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

	println("dbg/intg.1.listen")
	// Open listener.
	ln, err := marionette.Listen(mar.MustParse(marionette.PartyServer, input), "")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	println("dbg/intg.2.dial")
	// Connect using the dialer.
	dialer, err := marionette.NewDialer(mar.MustParse(marionette.PartyClient, input), "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	defer dialer.Close()

	println("dbg/intg.3.stream")
	// Open a stream and write.
	stream, err := dialer.Dial()
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	println("dbg/intg.4.write")
	if n, err := stream.Write([]byte("foo")); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatalf("unexpected n: %d", n)
	}

	println("dbg/intg.5.wait")

	// TEMP(benbjohnson): Wait
	<-(chan struct{})(nil)
}

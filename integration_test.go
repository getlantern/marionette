package marionette_test

import (
	"io"
	"testing"
	"time"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
)

func TestIntegration(t *testing.T) {
	t.Run("fte/async", func(t *testing.T) {
		prog := []byte(`
			connection(tcp, 8082):
			  start      upstream  NULL          1.0
			  upstream   downstream upstream     1.0
			  downstream upstream   downstream   1.0

			action upstream:
			  client fte.send_async("^.*$", 128)

			action downstream:
			  server fte.send_async("^.*$", 128)
			`)

		println("dbg/intg.1.listen")
		// Open listener.
		ln, err := marionette.Listen(mar.MustParse(marionette.PartyServer, prog), "")
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		println("dbg/intg.2.dialer")
		// Connect using the dialer.
		dialer, err := marionette.NewDialer(mar.MustParse(marionette.PartyClient, prog), "127.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		defer dialer.Close()

		println("dbg/intg.3.dial")
		// Open a stream and write.
		input, err := dialer.Dial()
		if err != nil {
			t.Fatal(err)
		}
		defer input.Close()

		println("dbg/intg.4.write")
		if n, err := input.Write([]byte("foo")); err != nil {
			t.Fatal(err)
		} else if n != 3 {
			t.Fatalf("unexpected n: %d", n)
		}

		// Accept stream from server-side.
		println("dbg/intg.5.accept")
		time.Sleep(1 * time.Second)
		output, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer output.Close()

		println("dbg/intg.6.read")
		// Read data from server-side.
		buf := make([]byte, 3)
		if _, err := io.ReadFull(output, buf); err != nil {
			t.Fatal(err)
		} else if string(buf) != `foo` {
			t.Fatalf("unexpected read: %s", buf)
		}
		println("dbg/intg.7.success")
	})
}

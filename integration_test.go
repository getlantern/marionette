package marionette_test

import (
	"bytes"
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

	t.Run("active_probing", func(t *testing.T) {
		t.Run("ftp_pureftpd_10", func(t *testing.T) {
			RunIntegration(t, mar.Format("active_probing/ftp_pureftpd_10", ""))
		})

		t.Run("http_apache_247", func(t *testing.T) {
			RunIntegration(t, mar.Format("active_probing/http_apache_247", ""))
		})

		t.Run("ssh_openssh_661", func(t *testing.T) {
			RunIntegration(t, mar.Format("active_probing/ssh_openssh_661", ""))
		})
	})

	t.Run("dns_request", func(t *testing.T) {
		t.Skip("TODO: udp")
		RunIntegration(t, mar.Format("dns_request", ""))
	})

	t.Run("dummy", func(t *testing.T) {
		RunIntegration(t, mar.Format("dummy", ""))
	})

	t.Run("ftp_pasv_transfer", func(t *testing.T) {
		t.Skip("TODO: invalid_connection_port")
		RunIntegration(t, mar.Format("ftp_pasv_transfer", ""))
	})

	t.Run("ftp_simple_blocking", func(t *testing.T) {
		t.Skip("TODO: tg")
		RunIntegration(t, mar.Format("ftp_simple_blocking", ""))
	})

	t.Run("http_active_probing", func(t *testing.T) {
		RunIntegration(t, mar.Format("http_active_probing", ""))
	})

	t.Run("http_active_probing2", func(t *testing.T) {
		RunIntegration(t, mar.Format("http_active_probing2", ""))
	})

	t.Run("http_probabilistic_blocking", func(t *testing.T) {
		RunIntegration(t, mar.Format("http_probabilistic_blocking", ""))
	})

	t.Run("http_simple_blocking", func(t *testing.T) {
		t.Skip("TODO: tg")
		RunIntegration(t, mar.Format("http_simple_blocking", ""))
	})

	t.Run("http_simple_blocking_with_msg_lens", func(t *testing.T) {
		t.Skip("TODO: tg")
		RunIntegration(t, mar.Format("http_simple_blocking_with_msg_lens", ""))
	})

	t.Run("http_simple_nonblocking", func(t *testing.T) {
		RunIntegration(t, mar.Format("http_simple_nonblocking", ""))
	})

	t.Run("http_squid_blocking", func(t *testing.T) {
		t.Skip("TODO: tg")
		RunIntegration(t, mar.Format("http_squid_blocking", ""))
	})

	t.Run("http_timings", func(t *testing.T) {
		RunIntegration(t, mar.Format("http_timings", ""))
	})

	t.Run("https_simple_blocking", func(t *testing.T) {
		RunIntegration(t, mar.Format("https_simple_blocking", ""))
	})

	t.Run("nmap", func(t *testing.T) {
		t.Run("kpdyer.com", func(t *testing.T) {
			RunIntegration(t, mar.Format("nmap/kpdyer.com", ""))
		})
	})

	t.Run("smb_simple_nonblocking", func(t *testing.T) {
		t.Skip("TODO: python2: invalid argument")
		RunIntegration(t, mar.Format("smb_simple_nonblocking", ""))
	})

	t.Run("ssh_simple_nonblocking", func(t *testing.T) {
		RunIntegration(t, mar.Format("ssh_simple_nonblocking", ""))
	})

	t.Run("ta", func(t *testing.T) {
		t.Skip("TODO: tg")

		t.Run("amzn_conn", func(t *testing.T) {
			RunIntegration(t, mar.Format("ta/amzn_conn", ""))
		})

		t.Run("amzn_sess", func(t *testing.T) {
			RunIntegration(t, mar.Format("ta/amzn_sess", ""))
		})
	})

	t.Run("test_hex_input_strings", func(t *testing.T) {
		t.Skip("TODO: Create 'end' transition if it doesn't exist")
		program := mar.Format("test_hex_input_strings", "")
		program = bytes.Replace(program, []byte(`connection(tcp, 80)`), []byte(`connection(tcp, 8080)`), -1)
		RunIntegration(t, program)
	})

	t.Run("udp_test_format", func(t *testing.T) {
		t.Skip("TODO: udp")
		RunIntegration(t, mar.Format("udp_test_format", ""))
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

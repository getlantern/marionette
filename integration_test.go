// +build integration

package marionette_test

import (
	"bytes"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mar"
)

func TestIntegration(t *testing.T) {
	t.Run("fte/async", func(t *testing.T) {
		RunSimpleIntegration(t, []byte(`
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
		RunSimpleIntegration(t, []byte(`
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
			RunSimpleIntegration(t, mar.Format("active_probing/ftp_pureftpd_10", ""))
		})

		t.Run("http_apache_247", func(t *testing.T) {
			RunSimpleIntegration(t, mar.Format("active_probing/http_apache_247", ""))
		})

		t.Run("ssh_openssh_661", func(t *testing.T) {
			RunSimpleIntegration(t, mar.Format("active_probing/ssh_openssh_661", ""))
		})
	})

	t.Run("dns_request", func(t *testing.T) {
		t.Skip("TODO: udp")
		RunSimpleIntegration(t, mar.Format("dns_request", ""))
	})

	t.Run("dummy", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("dummy", ""))
	})

	t.Run("ftp_simple_blocking", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("ftp_simple_blocking", ""))
	})

	t.Run("http_active_probing", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("http_active_probing", ""))
	})

	t.Run("http_active_probing2", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("http_active_probing2", ""))
	})

	t.Run("http_probabilistic_blocking", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("http_probabilistic_blocking", ""))
	})

	t.Run("http_simple_blocking", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("http_simple_blocking", ""))
	})

	t.Run("http_simple_blocking_with_msg_lens", func(t *testing.T) {
		t.Skip("TODO: tg")
		RunSimpleIntegration(t, mar.Format("http_simple_blocking_with_msg_lens", ""))
	})

	t.Run("http_simple_nonblocking", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("http_simple_nonblocking", ""))
	})

	t.Run("http_squid_blocking", func(t *testing.T) {
		t.Skip("TODO: tg")
		RunSimpleIntegration(t, mar.Format("http_squid_blocking", ""))
	})

	t.Run("http_timings", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("http_timings", ""))
	})

	t.Run("https_simple_blocking", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("https_simple_blocking", ""))
	})

	t.Run("nmap", func(t *testing.T) {
		t.Run("kpdyer.com", func(t *testing.T) {
			RunSimpleIntegration(t, mar.Format("nmap/kpdyer.com", ""))
		})
	})

	t.Run("smb_simple_nonblocking", func(t *testing.T) {
		t.Skip("TODO: python2: invalid argument")
		RunSimpleIntegration(t, mar.Format("smb_simple_nonblocking", ""))
	})

	t.Run("ssh_simple_nonblocking", func(t *testing.T) {
		RunSimpleIntegration(t, mar.Format("ssh_simple_nonblocking", ""))
	})

	t.Run("ta", func(t *testing.T) {
		t.Skip("TODO: tg")

		t.Run("amzn_conn", func(t *testing.T) {
			RunSimpleIntegration(t, mar.Format("ta/amzn_conn", ""))
		})

		t.Run("amzn_sess", func(t *testing.T) {
			RunSimpleIntegration(t, mar.Format("ta/amzn_sess", ""))
		})
	})

	t.Run("test_hex_input_strings", func(t *testing.T) {
		t.Skip("TODO: Create 'end' transition if it doesn't exist")
		program := mar.Format("test_hex_input_strings", "")
		program = bytes.Replace(program, []byte(`connection(tcp, 80)`), []byte(`connection(tcp, 8080)`), -1)
		RunSimpleIntegration(t, program)
	})

	t.Run("udp_test_format", func(t *testing.T) {
		t.Skip("TODO: udp")
		RunSimpleIntegration(t, mar.Format("udp_test_format", ""))
	})
}

func RunSimpleIntegration(t *testing.T, program []byte) {
	tt := MustOpenIntegrationTest(program)
	defer MustCloseIntegrationTest(tt)

	tt.Conn(
		func(conn net.Conn) {
			if _, err := conn.Write([]byte(`foo`)); err != nil {
				t.Fatal(err)
			} else if _, err := conn.Write([]byte(`bar`)); err != nil {
				t.Fatal(err)
			}

			resp := make([]byte, 11)
			if _, err := io.ReadFull(conn, resp); err != nil {
				t.Fatal(err)
			} else if string(resp) != `lorem ipsum` {
				t.Fatalf("unexpected response: %s", resp)
			}

			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		},
		func(conn net.Conn) {
			req := make([]byte, 6)
			if _, err := io.ReadFull(conn, req); err != nil {
				t.Fatal(err)
			} else if string(req) != `foobar` {
				t.Fatalf("unexpected request: %s", req)
			}

			if _, err := conn.Write([]byte(`lorem`)); err != nil {
				t.Fatal(err)
			} else if _, err := conn.Write([]byte(` `)); err != nil {
				t.Fatal(err)
			} else if _, err := conn.Write([]byte(`ipsum`)); err != nil {
				t.Fatal(err)
			}

			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		},
	)

	tt.Wait()
}

type IntegrationTest struct {
	mu sync.Mutex
	wg sync.WaitGroup

	ClientStreamSet *marionette.StreamSet
	ServerStreamSet *marionette.StreamSet

	Listener *marionette.Listener
	Dialer   *marionette.Dialer
}

func MustOpenIntegrationTest(program []byte) *IntegrationTest {
	tt := &IntegrationTest{
		ClientStreamSet: marionette.NewStreamSet(),
		ServerStreamSet: marionette.NewStreamSet(),
	}

	ln, err := marionette.Listen(mar.MustParse(marionette.PartyServer, program), "")
	if err != nil {
		panic(err)
	}
	tt.Listener = ln

	d, err := marionette.NewDialer(mar.MustParse(marionette.PartyClient, program), "127.0.0.1", tt.ClientStreamSet)
	if err != nil {
		tt.Listener.Close()
		panic(err)
	}
	tt.Dialer = d

	return tt
}

func MustCloseIntegrationTest(tt *IntegrationTest) {
	if err := tt.Close(); err != nil {
		panic(err)
	}
}

func (tt *IntegrationTest) Close() (err error) {
	if e := tt.Dialer.Close(); e != nil && err == nil {
		err = e
	}
	if e := tt.Listener.Close(); e != nil && err == nil {
		err = e
	}
	return err
}

func (tt *IntegrationTest) Wait() {
	tt.wg.Wait()
}

// Conn creates a client/server connection that runs in separate goroutines.
//
// The connection must be created before the function returns so that the
// listener accepts the matching client connection.
func (tt *IntegrationTest) Conn(clientFn, serverFn func(conn net.Conn)) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	tt.wg.Add(2)
	connected := make(chan struct{})

	// Execute client connection.
	var streamID uint32
	go func() {
		defer tt.wg.Done()

		conn, err := tt.Dialer.Dial()
		if err != nil {
			panic(err)
		}
		atomic.StoreUint32(&streamID, uint32(conn.(*marionette.Stream).ID()))

		clientFn(conn)
	}()

	// Execute server connection.
	go func() {
		defer tt.wg.Done()

		conn, err := tt.Listener.Accept()
		if err != nil {
			panic(err)
		} else if id := atomic.LoadUint32(&streamID); id != uint32(conn.(*marionette.Stream).ID()) {
			panic("stream mismatch")
		}

		connected <- struct{}{}
		serverFn(conn)
	}()

	// Wait until client/server connections are created.
	<-connected
}

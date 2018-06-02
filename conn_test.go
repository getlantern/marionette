package marionette_test

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/redjack/marionette"
)

func TestBufferedConn(t *testing.T) {
	// Generate 10MB of data.
	data := make([]byte, 10*1024*1024)
	for i := range data {
		data[i] = byte(i % 100)
	}

	// Open random port.
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Accept a connection in a separate goroutine and stream data to client.
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		if _, err := io.Copy(conn, bytes.NewReader(data)); err != nil {
			t.Fatal(err)
		}
	}()

	// Connect to listener.
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Wrap in bufferred connection.
	bufConn := marionette.NewBufferedConn(conn, marionette.MaxCellLength)

	// Read all data.
	var buf bytes.Buffer
	for {
		// Read 100b at a time.
		b, err := bufConn.Peek(100, false)
		if buf.Write(b); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}

		if _, err := bufConn.Seek(int64(len(b)), io.SeekCurrent); err != nil {
			t.Fatal(err)
		}
	}

	// Verify correctness.
	if b := buf.Bytes(); !bytes.Equal(b, data) {
		t.Fatalf("incorrect bytes read: got=%d, exp=%d", len(b), len(data))
	}
}

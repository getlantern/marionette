package mar_test

import (
	"bytes"
	"testing"

	"github.com/redjack/marionette/mar"
)

func TestFormat(t *testing.T) {
	t.Run("WithVersion", func(t *testing.T) {
		if buf := mar.Format("active_probing/ftp_pureftpd_10", "20150701"); !bytes.Contains(buf, []byte("Welcome to Pure-FTPd")) {
			t.Fatal("incorrect file")
		}
	})

	t.Run("NoVersion", func(t *testing.T) {
		if buf := mar.Format("http_simple_blocking", ""); !bytes.Contains(buf, []byte(`HTTP/1\.0`)) {
			t.Fatal("incorrect file")
		}
	})
}

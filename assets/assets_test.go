package assets_test

import (
	"bytes"
	"testing"

	"github.com/redjack/marionette/assets"
)

func TestFormat(t *testing.T) {
	t.Run("WithVersion", func(t *testing.T) {
		if buf := assets.Format("active_probing/ftp_pureftpd_10", "20150701"); !bytes.Contains(buf, []byte("Welcome to Pure-FTPd")) {
			t.Fatal("incorrect file")
		}
	})

	t.Run("NoVersion", func(t *testing.T) {
		if buf := assets.Format("http_simple_blocking", ""); !bytes.Contains(buf, []byte(`HTTP/1\.0`)) {
			t.Fatal("incorrect file")
		}
	})
}

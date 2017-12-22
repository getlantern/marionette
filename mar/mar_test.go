package mar_test

import (
	"bytes"
	"reflect"
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

func TestFormats(t *testing.T) {
	formats := mar.Formats()
	if !reflect.DeepEqual(formats, []string{
		"active_probing/ftp_pureftpd_10:20150701",
		"active_probing/http_apache_247:20150701",
		"active_probing/ssh_openssh_661:20150701",
		"dns_request:20150701",
		"dummy:20150701",
		"ftp_pasv_transfer:20150701",
		"ftp_simple_blocking:20150701",
		"http_active_probing:20150701",
		"http_active_probing2:20150701",
		"http_probabilistic_blocking:20150701",
		"http_simple_blocking:20150701",
		"http_simple_blocking_with_msg_lens:20150701",
		"http_simple_nonblocking:20150701",
		"http_squid_blocking:20150701",
		"http_timings:20150701",
		"https_simple_blocking:20150701",
		"nmap/kpdyer.com:20150701",
		"smb_simple_nonblocking:20150701",
		"ssh_simple_nonblocking:20150701",
		"ta/amzn_conn:20150701",
		"ta/amzn_sess:20150701",
		"test_hex_input_strings:20150701",
		"udp_test_format:20150701",
		"http_simple_blocking:20150702",
	}) {
		t.Fatalf("unexpected formats: %+v", formats)
	}
}

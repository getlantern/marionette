package tg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/plugins/tg"
)

func TestParse_DNSRequest(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := tg.Parse("dns_request", "AB\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00\x03foo\x03com\x00\x00\x01\x00\x01")
		if diff := cmp.Diff(m, map[string]string{
			"DNS_TRANSACTION_ID": "AB",
			"DNS_DOMAIN":         "\x03foo\x03com\x00",
		}); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestParse_DNSResponse(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := tg.Parse("dns_response", "AB\x81\x80\x00\x01\x00\x01\x00\x00\x00\x00\x03foo\x03com\x00\x00\x01\x00\x01\xc0\x0c\x00\x01\x00\x01\x00\x00\x00\x02\x00\x04\x0A\x0B\x0C\x0D")
		if diff := cmp.Diff(m, map[string]string{
			"DNS_TRANSACTION_ID": "AB",
			"DNS_DOMAIN":         "\x03foo\x03com\x00",
			"DNS_IP":             "\x0A\x0B\x0C\x0D",
		}); diff != "" {
			t.Fatal(diff)
		}
	})
}

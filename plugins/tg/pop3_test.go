package tg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/plugins/tg"
)

func TestParse_POP3(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := tg.Parse("pop3_message_response", "+OK 3 octets\nReturn-Path: sender@example.com\nReceived: from client.example.com ([192.0.2.1])\nFrom: sender@example.com\nSubject: Test message\nTo: recipient@example.com\n\nfoo\n.\n")
		if diff := cmp.Diff(m, map[string]string{
			"POP3-RESPONSE-BODY": "foo",
			"CONTENT-LENGTH":     "3",
		}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("ErrMissingBody", func(t *testing.T) {
		if m := tg.Parse("pop3_message_response", "+OK 0 octets\nReturn-Path: sender@example.com\nReceived: from client.example.com ([192.0.2.1])\nFrom: sender@example.com\nSubject: Test message\nTo: recipient@example.com\n.\n"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})

	t.Run("ErrMissingTrailer", func(t *testing.T) {
		if m := tg.Parse("pop3_message_response", "+OK 0 octets\nReturn-Path: sender@example.com\nReceived: from client.example.com ([192.0.2.1])\nFrom: sender@example.com\nSubject: Test message\nTo: recipient@example.com\n\nfoo"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})
}

func TestParse_POP3Password(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := tg.Parse("pop3_password", "PASS foo\n")
		if diff := cmp.Diff(m, map[string]string{
			"PASSWORD": "foo",
		}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("ErrMissingPrefix", func(t *testing.T) {
		if m := tg.Parse("pop3_password", "foo\n"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})

	t.Run("ErrMissingSuffix", func(t *testing.T) {
		if m := tg.Parse("pop3_password", "PASS foo"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})
}

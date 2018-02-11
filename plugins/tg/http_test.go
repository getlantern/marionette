package tg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/plugins/tg"
)

func TestParse_HTTPRequest(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		t.Run("WithScheme", func(t *testing.T) {
			m := tg.Parse("http_request", "GET http://127.0.0.1:8080/foo HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: keep-alive\r\n\r\n")
			if diff := cmp.Diff(m, map[string]string{
				"URL": "foo",
			}); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("WithoutScheme", func(t *testing.T) {
			m := tg.Parse("http_request", "GET /foo HTTP/1.1\r\nUser-Agent: marionette 0.1\r\nConnection: keep-alive\r\n\r\n")
			if diff := cmp.Diff(m, map[string]string{
				"URL": "foo",
			}); diff != "" {
				t.Fatal(diff)
			}
		})
	})

	t.Run("ErrInvalidMethod", func(t *testing.T) {
		if m := tg.Parse("http_request", "POST http://127.0.0.1:8080/foo"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})

	t.Run("ErrMissingBody", func(t *testing.T) {
		if m := tg.Parse("http_request", "GET http://127.0.0.1:8080/foo"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})
}

func TestParse_HTTPResponse(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		t.Run("WithBody", func(t *testing.T) {
			m := tg.Parse("http_response", "HTTP/1.1 200 OK\r\nContent-Length: 3\r\nConnection: keep-alive\r\n\r\nfoo")
			if diff := cmp.Diff(m, map[string]string{
				"COOKIE":             "",
				"CONTENT-LENGTH":     "3",
				"HTTP-RESPONSE-BODY": "foo",
			}); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("WithoutBody", func(t *testing.T) {
			m := tg.Parse("http_response", "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: keep-alive\r\n\r\n")
			if diff := cmp.Diff(m, map[string]string{
				"COOKIE":             "",
				"CONTENT-LENGTH":     "0",
				"HTTP-RESPONSE-BODY": "",
			}); diff != "" {
				t.Fatal(diff)
			}
		})
	})

	t.Run("ErrMissingVersion", func(t *testing.T) {
		if m := tg.Parse("http_response", "XYZ"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})

	t.Run("ErrContentLengthMismatch", func(t *testing.T) {
		if m := tg.Parse("http_response", "HTTP/1.1 200 OK\r\nContent-Length: 10\r\nConnection: keep-alive\r\n\r\nfoo"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})
}

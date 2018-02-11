package tg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/plugins/tg"
)

func TestParse_FTP(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := tg.Parse("ftp_entering_passive", "227 Entering Passive Mode (127,0,0,1,100,200).\n")
		if diff := cmp.Diff(m, map[string]string{
			"FTP_PASV_PORT_X": "100",
			"FTP_PASV_PORT_Y": "200",
		}); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("ErrMissingPrefix", func(t *testing.T) {
		if m := tg.Parse("ftp_entering_passive", "FOO Entering Passive Mode (127,0,0,1,100,200).\n"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})

	t.Run("ErrMissingSuffix", func(t *testing.T) {
		if m := tg.Parse("ftp_entering_passive", "227 Entering Passive Mode (127,0,0,1,100,200.\n"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})

	t.Run("ErrMissingArguments", func(t *testing.T) {
		if m := tg.Parse("ftp_entering_passive", "227 Entering Passive Mode (127,0,0,100,200).\n"); m != nil {
			t.Fatalf("unexpected values: %#v", m)
		}
	})
}

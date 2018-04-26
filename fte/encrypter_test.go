package fte_test

import (
	"encoding/hex"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/redjack/marionette/fte"
)

func TestEncrypter(t *testing.T) {
	enc, dec := MustNewEncrypter(), MustNewDecrypter()
	for _, plaintext := range [][]byte{
		[]byte("0fb37292bc72a5ce563448c9f9cc0154e3b1d2eb7dd0dc61bc2cb769756345dd5dbebca1b2"),
	} {
		t.Run(string(plaintext), func(t *testing.T) {
			if ciphertext, err := enc.Encrypt(plaintext); err != nil {
				t.Fatalf("encrypt(%x): %s", plaintext, err)
			} else if other, err := dec.Decrypt(ciphertext); err != nil {
				t.Fatalf("decrypt(%x): %s", plaintext, err)
			} else if diff := cmp.Diff(plaintext, other); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestEncrypter_Quick(t *testing.T) {
	enc := MustNewEncrypter()
	dec := MustNewDecrypter()

	if err := quick.Check(func(plaintext []byte) bool {
		if ciphertext, err := enc.Encrypt(plaintext); err != nil {
			t.Fatalf("encrypt(%x): %s", plaintext, err)
		} else if other, err := dec.Decrypt(ciphertext); err != nil {
			t.Fatalf("decrypt(%x): %s", plaintext, err)
		} else if diff := cmp.Diff(plaintext, other); diff != "" {
			t.Fatal(diff)
		}
		return true
	}, nil); err != nil {
		t.Fatal(err)
	}
}

func MustNewEncrypter() *fte.Encrypter {
	enc, err := fte.NewEncrypter()
	if err != nil {
		panic(err)
	}
	return enc
}

func MustNewDecrypter() *fte.Decrypter {
	dec, err := fte.NewDecrypter()
	if err != nil {
		panic(err)
	}
	return dec
}

func MustHexDecodeString(s string) []byte {
	buf, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return buf
}

// +build python

package fte_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/redjack/marionette/fte"
)

func TestEncrypter_Python(t *testing.T) {
	plaintext, err := hex.DecodeString(`e41748f79bb761b395b493cb23553226fa8fb42f48a63edb5eec42bf2da6970c47e870fc86db59dd3e811779fa`)
	if err != nil {
		t.Fatal(err)
	}

	// Create Go encrypter.
	enc, err := fte.NewEncrypter()
	if err != nil {
		t.Fatal(err)
	}
	enc.IV = []byte("\x00\x00\x00\x00\x00\x00\x00")
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// Encrypt using Python.
	exp, err := RunPythonEncrypterEncrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}

	// Compare.
	if !bytes.Equal(exp, ciphertext) {
		t.Fatalf("ciphertext mismatch:\nexp:%x\ngot:%x\n", exp, ciphertext)
	}
}

// RunPythonEncrypterEncrypt runs the Python Encrypter.Encrypt().
func RunPythonEncrypterEncrypt(plaintext []byte) (ciphertext []byte, err error) {
	stdout, err := RunPython(EncrypterEncryptProgram, hex.EncodeToString(plaintext))
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(stdout))
}

const EncrypterEncryptProgram = `
import sys
import fte.conf
import fte.encrypter
import binascii

plaintext = binascii.unhexlify(sys.argv[1])
encrypter = fte.encrypter.Encrypter()
ciphertext = bytearray(encrypter.encrypt(plaintext, iv_bytes='\x00\x00\x00\x00\x00\x00\x00'))
sys.stdout.write(binascii.hexlify(ciphertext))
sys.stdout.flush()
`

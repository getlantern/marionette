// +build python

package fte_test

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"testing/quick"

	"github.com/redjack/marionette/fte"
)

func TestCipher_Python(t *testing.T) {
	for _, tt := range []struct {
		name      string
		regex     string
		plaintext []byte
	}{
		{"http_simple_blocking/client", `^GET\ \/([a-zA-Z0-9\.\/]*) HTTP/1\.1\r\n\r\n$`, MustHexDecodeString(`000001720000004ec87f9d674d65822178629a100000000001474554202f20485454502f312e310d0a486f73743a206c6f63616c686f73743a383037390d0a557365722d4167656e743a206375726c2f372e35342e300d0a4163636570743a202a2f2a0d0a0d0a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000`)},
		{"http_simple_blocking/server", `^HTTP/1\.1\ 200 OK\r\nContent-Type:\ ([a-zA-Z0-9]+)\r\n\r\n\C*$`, []byte("foobar")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Create Go cipher.
			cipher, err := fte.NewCipher(tt.regex)
			if err != nil {
				t.Fatal(err)
			}
			defer cipher.Close()

			// Read capacity using Python cipher.
			if pyCapacity, err := RunPythonCipherCapacity(tt.regex); err != nil {
				t.Fatal(err)
			} else if cipher.Capacity() != pyCapacity/8 {
				t.Fatalf("capacity mismatch: got:%d, exp:%d", cipher.Capacity(), pyCapacity/8)
			}

			// Encrypt using Python cipher.
			pyCiphertext, err := RunPythonCipherEncrypt(tt.regex, tt.plaintext)
			if err != nil {
				t.Fatal(err)
			}

			// Encrypt using Go cipher.
			goCiphertext, err := cipher.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatal(err)
			}

			if len(pyCiphertext) != len(goCiphertext) {
				t.Fatalf("ciphertext len mismatch: py=%d go=%d", len(pyCiphertext), len(goCiphertext))
			}

			// Decrypt Go ciphertext using Python cipher.
			plaintext0, _, err := RunPythonCipherDecrypt(tt.regex, goCiphertext)
			if err != nil {
				t.Fatal(err)
			} else if !bytes.Equal(tt.plaintext, plaintext0) {
				t.Fatalf("decrypt mismatch:\nexp=%x\ngot=%x", tt.plaintext, plaintext0)
			}
		})
	}
}

func TestCipher_Quick_Python(t *testing.T) {
	for _, tt := range []struct {
		name  string
		regex string
	}{
		{"http_simple_blocking/client", `^GET\ \/([a-zA-Z0-9\.\/]*) HTTP/1\.1\r\n\r\n$`},
		{"http_simple_blocking/server", `^HTTP/1\.1\ 200 OK\r\nContent-Type:\ ([a-zA-Z0-9]+)\r\n\r\n\C*$`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			quick.Check(func(plaintext []byte) bool {
				// Skip blank.
				if len(plaintext) == 0 {
					return true
				}

				// Create Go cipher.
				cipher, err := fte.NewCipher(tt.regex)
				if err != nil {
					t.Fatal(err)
				}
				defer cipher.Close()

				// Read capacity using Python cipher.
				exp, err := RunPythonCipherCapacity(tt.regex)
				if err != nil {
					t.Fatal(err)
				}

				// Convert bits to bytes.
				if cipher.Capacity() != exp/8 {
					t.Fatalf("capacity mismatch: got:%d, exp:%d", cipher.Capacity(), exp/8)
				}

				// Encrypt using Go cipher.
				ciphertext0, err := cipher.Encrypt(plaintext)
				if err != nil {
					t.Fatal(err)
				}

				// Encrypt using Python cipher.
				ciphertext1, err := RunPythonCipherEncrypt(tt.regex, plaintext)
				if err != nil {
					t.Fatal(err)
				} else if len(ciphertext1) != 512 {
					t.Fatalf("unexpected Python ciphertext len: %d", len(ciphertext1))
				}

				// Ensure ciphertext length matches.
				if len(ciphertext0) != len(ciphertext1) {
					t.Fatalf("unexpected ciphertext len mismatch: go=%d, py=%d", len(ciphertext0), len(ciphertext1))
				}

				// Decrypt Go ciphertext using Python cipher.
				plaintext0, remainder0, err := RunPythonCipherDecrypt(tt.regex, ciphertext0)
				if err != nil {
					t.Fatal(err)
				} else if !bytes.Equal(plaintext, plaintext0) {
					t.Fatalf("go decode mismatch:\nexp=%x\ngot=%x", plaintext, plaintext0)
				}

				// Decrypt Python ciphertext using Go cipher.
				plaintext1, remainder1, err := cipher.Decrypt(ciphertext1)
				if err != nil {
					t.Fatal(err)
				} else if !bytes.Equal(plaintext, plaintext1) {
					t.Fatalf("py decode mismatch:\nexp=%x\ngot=%x", plaintext, plaintext1)
				}

				// Ensure remainder is the same for both.
				if !bytes.Equal(remainder0, remainder1) {
					t.Fatalf("remainder mismatch:\npy=%x\ngo=%x", remainder0, remainder1)
				}

				return true
			}, &quick.Config{})
		})
	}
}

// RunPythonCipherEncrypt runs the Python Cipher.Encrypt().
func RunPythonCipherEncrypt(regex string, plaintext []byte) (ciphertext []byte, err error) {
	const prog = `
import sys
import regex2dfa
import fte.encoder
import binascii

regex = sys.argv[1]
plaintext = binascii.unhexlify(sys.argv[2])

encoder = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(regex), 512)

ciphertext = bytearray(encoder.encode(plaintext))
sys.stdout.write(binascii.hexlify(ciphertext))
sys.stdout.flush()
`
	stdout, err := RunPython(prog, regex, hex.EncodeToString(plaintext))
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(stdout))
}

// RunPythonCipherDecrypt runs the Python Cipher.Decrypt().
func RunPythonCipherDecrypt(regex string, ciphertext []byte) (plaintext, remainder []byte, err error) {
	const prog = `
import sys
import regex2dfa
import fte.encoder
import binascii

regex = sys.argv[1]
ciphertext = binascii.unhexlify(sys.argv[2])

encoder = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(regex), 512)

plaintext, remainder = encoder.decode(ciphertext)
sys.stdout.write(binascii.hexlify(bytearray(plaintext)))
sys.stdout.write(":")
sys.stdout.write(binascii.hexlify(bytearray(remainder)))
sys.stdout.flush()
`
	stdout, err := RunPython(prog, regex, hex.EncodeToString(ciphertext))
	if err != nil {
		return nil, nil, err
	}

	a := strings.Split(string(stdout), ":")
	if plaintext, err = hex.DecodeString(a[0]); err != nil {
		return nil, nil, err
	} else if remainder, err = hex.DecodeString(a[1]); err != nil {
		return nil, nil, err
	}
	return plaintext, remainder, nil
}

// RunPythonCipherCapacity runs the Python Cipher.Capacity().
func RunPythonCipherCapacity(regex string) (int, error) {
	const prog = `
import sys
import regex2dfa
import fte.encoder
import binascii

regex = sys.argv[1]
encoder = fte.encoder.DfaEncoder(regex2dfa.regex2dfa(regex), 512)
sys.stdout.write(str(encoder.getCapacity()))
sys.stdout.flush()
`

	stdout, err := RunPython(prog, regex)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(stdout))
}

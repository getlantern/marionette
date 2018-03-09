package fte_test

import (
	"testing"

	"github.com/redjack/marionette/fte"
	//"fmt"
)

func TestCipher(t *testing.T) {
	cipher := fte.NewCipher(`^(a|b)+$`)
	if err := cipher.Open(); err != nil {
		t.Fatal(err)
	}
	defer cipher.Close()

	// Verify initial capacity.
	if capacity, err := cipher.Capacity(); err != nil {
		t.Fatal(err)
	} else if capacity != 15 {
		t.Fatalf("unexpected initial capacity: %d", capacity)
	}

	// Encode/decode first message.
	if ciphertext, err := cipher.Encrypt([]byte(`test`)); err != nil {
		t.Fatal(err)
	} else if plaintext, remainder, err := cipher.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `test` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	} else if string(remainder) != `` {
		t.Fatalf("unexpected remainder: %q", remainder)
	}

	// Encode/decode second message.
	if ciphertext, err := cipher.Encrypt([]byte(`foo bar`)); err != nil {
		t.Fatal(err)
	} else if plaintext, remainder, err := cipher.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `foo bar` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	} else if string(remainder) != `` {
		t.Fatalf("unexpected remainder: %q", remainder)
	}

	if err := cipher.Close(); err != nil {
		t.Fatal(err)
	}
}

/*func TestCipher2(t *testing.T) {
	cipher := fte.NewCipher(`^(a|b|c)+$`)
	if err := cipher.Open(); err != nil {
		t.Fatal(err)
	}
	defer cipher.Close()

	// Encode/decode first message.
	if ciphertext, err := cipher.Encrypt([]byte(`test`)); err != nil {
		t.Fatal(err)
	} else if plaintext, err := cipher.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `test` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	} else {
		fmt.Printf("%s\n", ciphertext);
		fmt.Printf("%s\n", plaintext);
	}


	// Encode/decode second message.
	if ciphertext, err := cipher.Encrypt([]byte(`foo bar`)); err != nil {
		t.Fatal(err)
	} else if plaintext, err := cipher.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `foo bar` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	}

	if err := cipher.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCipher3(t *testing.T) {
	cipher := fte.NewCipher(`^panda(a|b)+$`)
	if err := cipher.Open(); err != nil {
		t.Fatal(err)
	}
	defer cipher.Close()

	// Encode/decode first message.
	if ciphertext, err := cipher.Encrypt([]byte(`test`)); err != nil {
		t.Fatal(err)
	} else if plaintext, err := cipher.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `test` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	} else {
		fmt.Printf("%s\n", ciphertext);
		fmt.Printf("%s\n", plaintext);
	}


	// Encode/decode second message.
	if ciphertext, err := cipher.Encrypt([]byte(`foo bar`)); err != nil {
		t.Fatal(err)
	} else if plaintext, err := cipher.Decrypt(ciphertext); err != nil {
		t.Fatal(err)
	} else if string(plaintext) != `foo bar` {
		t.Fatalf("unexpected plaintext: %q", plaintext)
	}

	if err := cipher.Close(); err != nil {
		t.Fatal(err)
	}
} */
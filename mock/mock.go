package mock

import (
	"context"
	"net"
)

type Cipher struct {
	EncryptFunc  func(plaintext []byte) (ciphertext []byte, err error)
	DecryptFunc  func(ciphertext []byte) (plaintext []byte, err error)
	CapacityFunc func() int
}

func (c *Cipher) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	return c.EncryptFunc(plaintext)
}

func (c *Cipher) Capacity() int {
	return c.CapacityFunc()
}

func (c *Cipher) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	return c.DecryptFunc(ciphertext)
}

type Dialer struct {
	DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.DialContextFunc(ctx, network, address)
}

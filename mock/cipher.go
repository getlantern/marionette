package mock

type Cipher struct {
	CapacityFn func() int
	EncryptFn  func(plaintext []byte) (ciphertext []byte, err error)
	DecryptFn  func(ciphertext []byte) (plaintext, remainder []byte, err error)
}

func (m *Cipher) Capacity() int {
	return m.CapacityFn()
}

func (m *Cipher) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	return m.EncryptFn(plaintext)
}

func (m *Cipher) Decrypt(ciphertext []byte) (plaintext, remainder []byte, err error) {
	return m.DecryptFn(ciphertext)
}

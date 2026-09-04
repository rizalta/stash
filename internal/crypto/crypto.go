package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/argon2"
)

type KDFParams struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

func DeriveKey(password, salt []byte, params KDFParams) []byte {
	return argon2.IDKey(password, salt, params.Time, params.Memory, params.Threads, params.KeyLen)
}

func DefaultKDFParams() KDFParams {
	return KDFParams{Time: 1, Memory: 64 * 1024, Threads: 4, KeyLen: 32}
}

func GenerateSalt() ([]byte, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

func Encrypt(key, p []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	c := gcm.Seal(nonce, nonce, p, nil)

	return c, nil
}

func Decrypt(key, c []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(c) < nonceSize {
		return nil, errors.New("crypto: ciphertext too short")
	}
	nonce, ciphertext := c[:nonceSize], c[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

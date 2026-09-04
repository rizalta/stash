package crypto

import (
	"bytes"
	"testing"
)

var (
	password  = []byte("some_password")
	salt, _   = GenerateSalt()
	kdfParams = DefaultKDFParams()
)

func TestEncryptDecrypt(t *testing.T) {
	key := DeriveKey(password, salt, kdfParams)

	plaintext := []byte("testing encryption")
	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decryptedText, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decryptedText) {
		t.Errorf("expected %s, got %s", plaintext, decryptedText)
	}
}

func TestWrongKey(t *testing.T) {
	key := DeriveKey(password, salt, DefaultKDFParams())

	plaintext := []byte("testing encryption")
	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	invalidKey := DeriveKey([]byte("invalid_password"), salt, kdfParams)

	if _, err := Decrypt(invalidKey, ciphertext); err == nil {
		t.Error("expected err, but got nil")
	}
}

func TestCiphertextTamper(t *testing.T) {
	key := DeriveKey(password, salt, kdfParams)

	plaintext := []byte("testing encryption")
	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	tampered := bytes.Clone(ciphertext)
	tampered[len(tampered)-1] ^= 0xFF

	if _, err := Decrypt(key, tampered); err == nil {
		t.Error("expected err, but got nil")
	}
}

func TestCiphertextTooShort(t *testing.T) {
	key := DeriveKey(password, salt, kdfParams)

	ciphertext := []byte("1234")
	if _, err := Decrypt(key, ciphertext); err == nil {
		t.Error("expected err, but got nil")
	}
}

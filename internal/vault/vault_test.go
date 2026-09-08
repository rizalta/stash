package vault

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestCreateAndOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.json")
	password := []byte("correct-password")

	v, err := Create(path, password)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	opened, err := Open(path, password)
	if err != nil {
		t.Fatalf("failed to open vault with correct password: %v", err)
	}
	if !bytes.Equal(opened.dek, v.dek) {
		t.Error("dek mismatch between create and open")
	}

	if _, err := Open(path, []byte("wrong-password")); !errors.Is(err, ErrWrongPassword) {
		t.Errorf("expected ErrWrongPassword, got %v", err)
	}

	if _, err := Create(path, password); !errors.Is(err, ErrVaultExists) {
		t.Errorf("expected ErrVaultExists, got %v", err)
	}
}

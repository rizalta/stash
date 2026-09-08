package vault

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

var password = []byte("correct-password")

func newTestVault(t *testing.T) *Vault {
	path := filepath.Join(t.TempDir(), "vault.json")

	v, err := Create(path, password)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	return v
}

func TestCreateAndOpen(t *testing.T) {
	v := newTestVault(t)
	path := v.path

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

func TestAddReopen(t *testing.T) {
	v := newTestVault(t)

	e := Entry{
		Title:    "Test",
		Username: "test_user",
		Password: "user_password_test",
		URL:      "https://test.com",
	}

	id, err := v.Add(e)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	path := v.path
	opened, err := Open(path, password)
	if err != nil {
		t.Fatalf("failed to open vault with correct password: %v", err)
	}

	re, err := opened.Get(id)
	if err != nil {
		t.Fatalf("failed to retrieve entry from opened vault: %v", err)
	}

	if !reflect.DeepEqual(e, re) {
		t.Error("retrieved entry is not same as the added")
	}
}

func TestUpdateGet(t *testing.T) {
	v := newTestVault(t)

	e := Entry{
		Title:    "Test",
		Username: "test_user",
		Password: "user_password_test",
		URL:      "https://test.com",
	}

	id, err := v.Add(e)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	e.Password = "updated_user_password"

	if err := v.Update(id, e); err != nil {
		t.Fatalf("failed to update entry: %v", err)
	}

	re, err := v.Get(id)
	if err != nil {
		t.Fatalf("failed to retrieve entry from vault: %v", err)
	}

	if !reflect.DeepEqual(re, e) {
		t.Fatalf("retrieved entry is not same as the updated")
	}

	entries := v.List()
	if len(entries) != 1 {
		t.Errorf("expected num of entries 1, got %d", len(entries))
	}
}

func TestDeleteGet(t *testing.T) {
	v := newTestVault(t)

	e := Entry{
		Title:    "Test",
		Username: "test_user",
		Password: "user_password_test",
		URL:      "https://test.com",
	}

	id, err := v.Add(e)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	if err := v.Delete(id); err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	if _, err := v.Get(id); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("expected error %v, got %v", ErrEntryNotFound, err)
	}

	entries := v.List()

	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

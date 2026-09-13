package vault

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

var password = []byte("correct-password")

func newTestVault(t *testing.T) (*Vault, string) {
	path := filepath.Join(t.TempDir(), "vault.db")
	ctx := context.Background()

	err := Create(ctx, path, password)
	if err != nil {
		t.Fatalf("failed to create vault: %v", err)
	}

	v, err := Open(ctx, path, password)
	if err != nil {
		t.Fatalf("failed to open vault: %v", err)
	}

	t.Cleanup(func() { _ = v.Close() })

	return v, path
}

func TestCreateAndOpen(t *testing.T) {
	_, path := newTestVault(t)
	ctx := context.Background()

	if _, err := Open(ctx, path, []byte("wrong-password")); !errors.Is(err, ErrWrongPassword) {
		t.Errorf("expected ErrWrongPassword, got %v", err)
	}

	if err := Create(ctx, path, password); !errors.Is(err, ErrVaultExists) {
		t.Errorf("expected ErrVaultExists, got %v", err)
	}
}

func TestAddReopen(t *testing.T) {
	v, path := newTestVault(t)
	ctx := context.Background()

	e := Entry{
		Title:    "Test",
		Username: "test_user",
		Password: "user_password_test",
		URL:      "https://test.com",
	}

	id, err := v.Add(ctx, e)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	opened, err := Open(ctx, path, password)
	if err != nil {
		t.Fatalf("failed to reopen the vault: %v", err)
	}
	defer func() { _ = opened.Close() }()

	re, err := opened.Get(ctx, id)
	if err != nil {
		t.Fatalf("failed to retrieve entry from opened vault: %v", err)
	}

	if !reflect.DeepEqual(e, re) {
		t.Error("retrieved entry is not same as the added")
	}
}

func TestUpdateGet(t *testing.T) {
	v, _ := newTestVault(t)
	ctx := context.Background()

	e := Entry{
		Title:    "Test",
		Username: "test_user",
		Password: "user_password_test",
		URL:      "https://test.com",
	}

	id, err := v.Add(ctx, e)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	e.Password = "updated_user_password"

	if err := v.Update(ctx, id, e); err != nil {
		t.Fatalf("failed to update entry: %v", err)
	}

	re, err := v.Get(ctx, id)
	if err != nil {
		t.Fatalf("failed to retrieve entry from vault: %v", err)
	}

	if !reflect.DeepEqual(re, e) {
		t.Fatalf("retrieved entry is not same as the updated")
	}

	entries, err := v.List(ctx)
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected num of entries 1, got %d", len(entries))
	}
}

func TestDeleteGet(t *testing.T) {
	v, _ := newTestVault(t)
	ctx := context.Background()

	e := Entry{
		Title:    "Test",
		Username: "test_user",
		Password: "user_password_test",
		URL:      "https://test.com",
	}

	id, err := v.Add(ctx, e)
	if err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	if err := v.Delete(ctx, id); err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	if _, err := v.Get(ctx, id); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("expected error %v, got %v", ErrEntryNotFound, err)
	}

	entries, err := v.List(ctx)
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestList(t *testing.T) {
	v, _ := newTestVault(t)
	ctx := context.Background()

	entries := []Entry{
		{Title: "aaa", Username: "user-1", Password: "password-1", URL: "https://aaa.com", Notes: "note-1"},
		{Title: "bbb", Username: "user-2", Password: "password-2", URL: "https://bbb.com", Notes: "note-2"},
		{Title: "ccc", Username: "user-3", Password: "password-3", URL: "https://ccc.com", Notes: "note-3"},
	}

	for _, e := range entries {
		if _, err := v.Add(ctx, e); err != nil {
			t.Fatalf("failed to add entry: %v", err)
		}
	}

	listed, err := v.List(ctx)
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(listed) != len(entries) {
		t.Errorf("expected %d entries, got %d", len(entries), len(listed))
	}

	for i, em := range listed {
		if em.Title != entries[i].Title {
			t.Errorf("entry %d title mismatch: expected %s, got %s", i, entries[i].Title, em.Title)
		}
	}
}

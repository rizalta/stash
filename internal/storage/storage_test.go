package storage

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/rizalta/stash/internal/crypto"
)

var header = Header{
	Version:      1,
	KDFSalt:      []byte("test-salt"),
	KDFParams:    crypto.DefaultKDFParams(),
	EncryptedDEK: []byte("test-encrypted-dek"),
	Verifier:     []byte("test-verifier"),
}

func newTestStorage(t *testing.T) *Storage {
	path := filepath.Join(t.TempDir(), "vault.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open test storage: %v", err)
	}
	if err := s.Init(context.Background(), header); err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestInitAndLoadHeader(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	h, err := s.LoadHeader(ctx)
	if err != nil {
		t.Fatalf("failed to load header: %v", err)
	}

	if !reflect.DeepEqual(h, header) {
		t.Error("loaded header different from the expeced header")
	}
}

func TestAddAndGetEntry(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	te := EntryRecord{
		ID:         "test-1",
		Title:      "test-title",
		Ciphertext: []byte("test-ciphertext"),
		ModifiedAt: time.Now().Unix(),
		Deleted:    false,
	}

	if err := s.AddEntry(ctx, te); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	ge, err := s.GetEntry(ctx, te.ID)
	if err != nil {
		t.Fatalf("failed to get entry: %v", err)
	}

	if !reflect.DeepEqual(ge, te) {
		t.Error("retrieved entry not same as added entry")
	}
}

func TestGetEntryNotFound(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	if _, err := s.GetEntry(ctx, "non-existent-id"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListEntries(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	entries := []EntryRecord{
		{ID: "test-1", Title: "title-1", Ciphertext: []byte("cipher-1"), ModifiedAt: time.Now().Unix()},
		{ID: "test-2", Title: "title-2", Ciphertext: []byte("cipher-2"), ModifiedAt: time.Now().Unix()},
		{ID: "test-3", Title: "title-3", Ciphertext: []byte("cipher-3"), ModifiedAt: time.Now().Unix()},
	}

	for _, e := range entries {
		if err := s.AddEntry(ctx, e); err != nil {
			t.Fatalf("failed to add entry: %v", err)
		}
	}

	listed, err := s.ListEntries(ctx)
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(listed) != len(entries) {
		t.Errorf("expected %d entries, got %d", len(entries), len(listed))
	}

	for i, e := range listed {
		if e.ID != entries[i].ID || e.Title != entries[i].Title || e.ModifiedAt != entries[i].ModifiedAt {
			t.Errorf("entry %d mismatch: got %+v", i, e)
		}
	}
}

func TestUpdateEntry(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	te := EntryRecord{
		ID:         "test-1",
		Title:      "original-title",
		Ciphertext: []byte("original-ciphertext"),
		ModifiedAt: time.Now().Unix(),
	}

	if err := s.AddEntry(ctx, te); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	te.Title = "updated-title"
	te.Ciphertext = []byte("updated-ciphertext")
	te.ModifiedAt = time.Now().Unix()

	if err := s.UpdateEntry(ctx, te); err != nil {
		t.Fatalf("failed to update entry: %v", err)
	}

	ge, err := s.GetEntry(ctx, te.ID)
	if err != nil {
		t.Fatalf("failed to get entry: %v", err)
	}

	if !reflect.DeepEqual(ge, te) {
		t.Error("retrieved entry not same as updated entry")
	}
}

func TestUpdateEntryNotFound(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	te := EntryRecord{
		ID:         "non-existent-id",
		Title:      "title",
		Ciphertext: []byte("ciphertext"),
		ModifiedAt: time.Now().Unix(),
	}

	if err := s.UpdateEntry(ctx, te); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetDeleted(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	te := EntryRecord{
		ID:         "test-1",
		Title:      "test-title",
		Ciphertext: []byte("test-ciphertext"),
		ModifiedAt: time.Now().Unix(),
	}

	if err := s.AddEntry(ctx, te); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	if err := s.SetDeleted(ctx, te.ID, time.Now().Unix()); err != nil {
		t.Fatalf("failed to delete entry: %v", err)
	}

	if _, err := s.GetEntry(ctx, te.ID); err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSetDeletedNotFound(t *testing.T) {
	s := newTestStorage(t)
	ctx := context.Background()

	if err := s.SetDeleted(ctx, "non-existent-id", time.Now().Unix()); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

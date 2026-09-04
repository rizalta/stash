package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rizalta/stash/internal/crypto"
)

func TestSaveLoad(t *testing.T) {
	entries := []EntryRecord{
		{
			ID:         "id1",
			Ciphertext: []byte("data1, xaisj, ufjref"),
			ModifiedAt: 1234,
			Deleted:    false,
		}, {
			ID:         "id2",
			Ciphertext: []byte("data2, 2329482, jjskajc"),
			ModifiedAt: 567821,
			Deleted:    true,
		}, {
			ID:         "id13",
			Ciphertext: []byte("...."),
			ModifiedAt: 783787328,
			Deleted:    false,
		},
	}

	vf := VaultFile{
		Version:      1,
		KDFSalt:      []byte("random_salt"),
		KDFParams:    crypto.DefaultKDFParams(),
		EncryptedDEK: []byte("some_encrypted_dek"),
		Verifier:     []byte("vault-ok"),
		Entries:      entries,
	}

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.json")

	if err := SaveVaultFile(path, &vf); err != nil {
		t.Fatalf("failed to save the vault file: %v", err)
	}

	loadedVf, err := LoadVaultFile(path)
	if err != nil {
		t.Fatalf("failed to load the vault file: %v", err)
	}

	if !reflect.DeepEqual(loadedVf, &vf) {
		t.Error("loaded vf not same as the saved vf")
	}
}

func TestLoadInvalidPath(t *testing.T) {
	invalidPath := "invalid_path"
	if _, err := LoadVaultFile(invalidPath); err == nil {
		t.Errorf("expected err, got nil")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	invalidJSONData := []byte("not_a_json_at_all")
	tempFile := filepath.Join(t.TempDir(), "test.json")

	if err := os.WriteFile(tempFile, invalidJSONData, 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if _, err := LoadVaultFile(tempFile); err == nil {
		t.Errorf("expected err, got nil")
	}
}

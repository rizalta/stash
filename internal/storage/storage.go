package storage

import (
	"encoding/json"
	"os"

	"github.com/rizalta/stash/internal/crypto"
)

type VaultFile struct {
	Version      int              `json:"version"`
	KDFSalt      []byte           `json:"kdf_salt"`
	KDFParams    crypto.KDFParams `json:"kdf_params"`
	EncryptedDEK []byte           `json:"encrypted_dek"`
	Verifier     []byte           `json:"verifier"`
	Entries      []EntryRecord    `json:"entries"`
}

type EntryRecord struct {
	ID         string `json:"id"`
	Ciphertext []byte `json:"ciphertext"`
	ModifiedAt int64  `json:"modified_at"`
	Deleted    bool   `json:"deleted"`
}

func LoadVaultFile(path string) (*VaultFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var vf VaultFile
	if err := json.Unmarshal(data, &vf); err != nil {
		return nil, err
	}

	return &vf, nil
}

func SaveVaultFile(path string, vf *VaultFile) error {
	data, err := json.MarshalIndent(vf, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

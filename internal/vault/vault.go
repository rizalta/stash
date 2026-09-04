package vault

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"github.com/rizalta/stash/internal/crypto"
	"github.com/rizalta/stash/internal/storage"
)

const verifierStr = "vault-ok"

var (
	ErrWrongPassword = errors.New("vault: wrong master password")
	ErrVaultExists   = errors.New("vault: file already exists")
)

type Entry struct {
	Title    string
	Username string
	Password string
	URL      string
	Notes    string
}

type Vault struct {
	path string
	dek  []byte
	file *storage.VaultFile
}

func Create(path string, password []byte) (*Vault, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, ErrVaultExists
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("vault: generating salt: %w", err)
	}

	kdfParams := crypto.DefaultKDFParams()
	kek := crypto.DeriveKey(password, salt, kdfParams)

	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("vault: generating dek: %w", err)
	}

	encryptedDek, err := crypto.Encrypt(kek, dek)
	if err != nil {
		return nil, fmt.Errorf("vault: encrypting dek: %w", err)
	}

	verifier, err := crypto.Encrypt(dek, []byte(verifierStr))
	if err != nil {
		return nil, fmt.Errorf("vault: encrypting verifier: %w", err)
	}

	vf := &storage.VaultFile{
		Version:      1,
		KDFSalt:      salt,
		KDFParams:    kdfParams,
		EncryptedDEK: encryptedDek,
		Verifier:     verifier,
		Entries:      []storage.EntryRecord{},
	}

	if err := storage.SaveVaultFile(path, vf); err != nil {
		return nil, fmt.Errorf("vault: saving vault file: %w", err)
	}

	v := &Vault{
		path: path,
		dek:  dek,
		file: vf,
	}

	return v, nil
}

func Open(path string, password []byte) (*Vault, error) {
	vf, err := storage.LoadVaultFile(path)
	if err != nil {
		return nil, fmt.Errorf("vault: loading vault file: %w", err)
	}

	kek := crypto.DeriveKey(password, vf.KDFSalt, vf.KDFParams)
	dek, err := crypto.Decrypt(kek, vf.EncryptedDEK)
	if err != nil {
		return nil, ErrWrongPassword
	}

	verifier, err := crypto.Decrypt(dek, vf.Verifier)
	if err != nil {
		return nil, ErrWrongPassword
	}

	if string(verifier) != verifierStr {
		return nil, ErrWrongPassword
	}

	v := &Vault{
		path: path,
		dek:  dek,
		file: vf,
	}
	return v, nil
}

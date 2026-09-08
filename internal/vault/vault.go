package vault

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rizalta/stash/internal/crypto"
	"github.com/rizalta/stash/internal/storage"
)

const verifierStr = "vault-ok"

var (
	ErrWrongPassword = errors.New("vault: wrong master password")
	ErrVaultExists   = errors.New("vault: file already exists")
	ErrEntryNotFound = errors.New("vault: entry not found")
	ErrCorruptEntry  = errors.New("vault: entry corrupt or tampered")
)

type Entry struct {
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Notes    string `json:"notes"`
}

type EntryMeta struct {
	ID         string
	Title      string
	ModifiedAt int64
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

func (v *Vault) Add(e Entry) (string, error) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", fmt.Errorf("vault: encoding entry: %w", err)
	}

	encryptedData, err := crypto.Encrypt(v.dek, data)
	if err != nil {
		return "", fmt.Errorf("vault: encrypting entry: %w", err)
	}

	id := uuid.NewString()
	record := storage.EntryRecord{
		ID:         id,
		Title:      e.Title,
		Ciphertext: encryptedData,
		ModifiedAt: time.Now().Unix(),
		Deleted:    false,
	}

	newEntries := append(v.file.Entries, record)
	tempFile := *v.file
	tempFile.Entries = newEntries

	if err := storage.SaveVaultFile(v.path, &tempFile); err != nil {
		return "", fmt.Errorf("vault: saving vault file: %w", err)
	}

	v.file.Entries = newEntries

	return id, nil
}

func (v *Vault) Get(id string) (Entry, error) {
	for _, er := range v.file.Entries {
		if er.ID == id {
			if er.Deleted {
				return Entry{}, ErrEntryNotFound
			}

			data, err := crypto.Decrypt(v.dek, er.Ciphertext)
			if err != nil {
				return Entry{}, fmt.Errorf("%w: %w", ErrCorruptEntry, err)
			}

			e := Entry{}
			if err := json.Unmarshal(data, &e); err != nil {
				return Entry{}, fmt.Errorf("%w: %w", ErrCorruptEntry, err)
			}

			return e, nil
		}
	}

	return Entry{}, ErrEntryNotFound
}

func (v *Vault) List() []EntryMeta {
	ret := []EntryMeta{}
	for _, er := range v.file.Entries {
		if !er.Deleted {
			ret = append(ret, EntryMeta{
				ID:         er.ID,
				Title:      er.Title,
				ModifiedAt: er.ModifiedAt,
			})
		}
	}

	return ret
}

func (v *Vault) Update(id string, e Entry) error {
	idx := -1
	for i, er := range v.file.Entries {
		if er.ID == id {
			if !er.Deleted {
				idx = i
			}
			break
		}
	}

	if idx == -1 {
		return ErrEntryNotFound
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("vault: encoding entry: %w", err)
	}

	encryptedData, err := crypto.Encrypt(v.dek, data)
	if err != nil {
		return fmt.Errorf("vault: encrypting entry: %w", err)
	}

	record := storage.EntryRecord{
		ID:         id,
		Title:      e.Title,
		Ciphertext: encryptedData,
		ModifiedAt: time.Now().Unix(),
		Deleted:    false,
	}

	newEntries := make([]storage.EntryRecord, len(v.file.Entries))
	copy(newEntries, v.file.Entries)
	newEntries[idx] = record

	tempFile := *v.file
	tempFile.Entries = newEntries

	if err := storage.SaveVaultFile(v.path, &tempFile); err != nil {
		return fmt.Errorf("vault: saving vault file: %w", err)
	}

	v.file.Entries = newEntries

	return nil
}

func (v *Vault) Delete(id string) error {
	idx := -1
	for i, er := range v.file.Entries {
		if er.ID == id {
			if !er.Deleted {
				idx = i
			}
			break
		}
	}

	if idx == -1 {
		return ErrEntryNotFound
	}

	record := v.file.Entries[idx]
	record.Deleted = true
	record.ModifiedAt = time.Now().Unix()

	newEntries := make([]storage.EntryRecord, len(v.file.Entries))
	copy(newEntries, v.file.Entries)
	newEntries[idx] = record

	tempFile := *v.file
	tempFile.Entries = newEntries

	if err := storage.SaveVaultFile(v.path, &tempFile); err != nil {
		return fmt.Errorf("vault: saving vault file: %w", err)
	}

	v.file.Entries = newEntries

	return nil
}

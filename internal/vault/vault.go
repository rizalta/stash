package vault

import (
	"context"
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

type Vault struct {
	dek     []byte
	storage *storage.Storage
}

func Create(ctx context.Context, path string, password []byte) error {
	if _, err := os.Stat(path); err == nil {
		return ErrVaultExists
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("vault: generating salt: %w", err)
	}

	kdfParams := crypto.DefaultKDFParams()
	kek := crypto.DeriveKey(password, salt, kdfParams)

	dek := make([]byte, 32)
	if _, err := rand.Read(dek); err != nil {
		return fmt.Errorf("vault: generating dek: %w", err)
	}

	encryptedDek, err := crypto.Encrypt(kek, dek)
	if err != nil {
		return fmt.Errorf("vault: encrypting dek: %w", err)
	}

	verifier, err := crypto.Encrypt(dek, []byte(verifierStr))
	if err != nil {
		return fmt.Errorf("vault: encrypting verifier: %w", err)
	}

	h := storage.Header{
		Version:      1,
		KDFSalt:      salt,
		KDFParams:    kdfParams,
		EncryptedDEK: encryptedDek,
		Verifier:     verifier,
	}

	s, err := storage.Open(path)
	if err != nil {
		return fmt.Errorf("vault: opening storage: %w", err)
	}
	if err := s.Init(ctx, h); err != nil {
		return fmt.Errorf("vault: initializing vault: %w", err)
	}

	if err := s.Close(); err != nil {
		return fmt.Errorf("vault: closing storage: %w", err)
	}

	return nil
}

func Open(ctx context.Context, path string, password []byte) (v *Vault, err error) {
	s, err := storage.Open(path)
	if err != nil {
		return nil, fmt.Errorf("vault: loading storage: %w", err)
	}
	defer func() {
		if err != nil {
			_ = s.Close()
		}
	}()

	h, err := s.LoadHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("vault: loading header: %w", err)
	}

	kek := crypto.DeriveKey(password, h.KDFSalt, h.KDFParams)
	dek, err := crypto.Decrypt(kek, h.EncryptedDEK)
	if err != nil {
		return nil, ErrWrongPassword
	}

	verifier, err := crypto.Decrypt(dek, h.Verifier)
	if err != nil {
		return nil, ErrWrongPassword
	}

	if string(verifier) != verifierStr {
		return nil, ErrWrongPassword
	}

	return &Vault{
		dek:     dek,
		storage: s,
	}, nil
}

func (v *Vault) Close() error {
	return v.storage.Close()
}

func (v *Vault) Add(ctx context.Context, e Entry) (string, error) {
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

	if err := v.storage.AddEntry(ctx, record); err != nil {
		return "", fmt.Errorf("vault: adding entry: %w", err)
	}

	return id, nil
}

func (v *Vault) Get(ctx context.Context, id string) (Entry, error) {
	record, err := v.storage.GetEntry(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return Entry{}, ErrEntryNotFound
		}
		return Entry{}, fmt.Errorf("vault: getting entry: %w", err)
	}

	data, err := crypto.Decrypt(v.dek, record.Ciphertext)
	if err != nil {
		return Entry{}, fmt.Errorf("%w: %w", ErrCorruptEntry, err)
	}

	e := Entry{}
	if err := json.Unmarshal(data, &e); err != nil {
		return Entry{}, fmt.Errorf("%w: %w", ErrCorruptEntry, err)
	}

	return e, nil
}

func (v *Vault) List(ctx context.Context) ([]storage.EntryMeta, error) {
	entries, err := v.storage.ListEntries(ctx)
	if err != nil {
		return []storage.EntryMeta{}, fmt.Errorf("vault: listing entries: %w", err)
	}

	return entries, nil
}

func (v *Vault) Update(ctx context.Context, id string, e Entry) error {
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

	if err := v.storage.UpdateEntry(ctx, record); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrEntryNotFound
		}
		return fmt.Errorf("vault: updating entry: %w", err)
	}

	return nil
}

func (v *Vault) Delete(ctx context.Context, id string) error {
	if err := v.storage.SetDeleted(ctx, id, time.Now().Unix()); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrEntryNotFound
		}
		return fmt.Errorf("vault: deleting an entry: %w", err)
	}

	return nil
}

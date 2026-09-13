package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rizalta/stash/db"
	"github.com/rizalta/stash/internal/crypto"
	"github.com/rizalta/stash/internal/repo"
	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("storage: entry not found")

type Storage struct {
	queries *repo.Queries
	conn    *sql.DB
}

func Open(path string) (*Storage, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage: opening db: %w", err)
	}

	return &Storage{
		queries: repo.New(conn),
		conn:    conn,
	}, nil
}

func (s *Storage) Close() error {
	return s.conn.Close()
}

type Header struct {
	Version      int
	KDFSalt      []byte
	KDFParams    crypto.KDFParams
	EncryptedDEK []byte
	Verifier     []byte
}

type EntryRecord struct {
	ID         string
	Title      string
	Ciphertext []byte
	ModifiedAt int64
	Deleted    bool
}

type EntryMeta struct {
	ID         string
	Title      string
	ModifiedAt int64
}

func (s *Storage) Init(ctx context.Context, h Header) error {
	if _, err := s.conn.ExecContext(ctx, db.Schema); err != nil {
		return fmt.Errorf("storage: creating schema: %w", err)
	}

	params := repo.InitMetaParams{
		Version:      int64(h.Version),
		KdfSalt:      h.KDFSalt,
		KdfTime:      int64(h.KDFParams.Time),
		KdfMemory:    int64(h.KDFParams.Memory),
		KdfThreads:   int64(h.KDFParams.Threads),
		KdfKeyLen:    int64(h.KDFParams.KeyLen),
		EncryptedDek: h.EncryptedDEK,
		Verifier:     h.Verifier,
	}

	if err := s.queries.InitMeta(ctx, params); err != nil {
		return fmt.Errorf("storage: initializing metadata: %w", err)
	}

	return nil
}

func (s *Storage) LoadHeader(ctx context.Context) (Header, error) {
	meta, err := s.queries.GetMeta(ctx)
	if err != nil {
		return Header{}, fmt.Errorf("storage: loading header: %w", err)
	}

	kdfParams := crypto.KDFParams{
		Time:    uint32(meta.KdfTime),
		Memory:  uint32(meta.KdfMemory),
		Threads: uint8(meta.KdfThreads),
		KeyLen:  uint32(meta.KdfKeyLen),
	}

	return Header{
		Version:      int(meta.Version),
		KDFSalt:      meta.KdfSalt,
		KDFParams:    kdfParams,
		EncryptedDEK: meta.EncryptedDek,
		Verifier:     meta.Verifier,
	}, nil
}

func (s *Storage) AddEntry(ctx context.Context, e EntryRecord) error {
	params := repo.InsertEntryParams{
		ID:         e.ID,
		Title:      e.Title,
		Ciphertext: e.Ciphertext,
		ModifiedAt: e.ModifiedAt,
	}

	if err := s.queries.InsertEntry(ctx, params); err != nil {
		return fmt.Errorf("storage: inserting entry: %w", err)
	}

	return nil
}

func (s *Storage) GetEntry(ctx context.Context, id string) (EntryRecord, error) {
	row, err := s.queries.GetEntry(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EntryRecord{}, ErrNotFound
		}
		return EntryRecord{}, fmt.Errorf("storage: getting entry: %w", err)
	}

	return EntryRecord{
		ID:         row.ID,
		Title:      row.Title,
		Ciphertext: row.Ciphertext,
		ModifiedAt: row.ModifiedAt,
		Deleted:    row.Deleted != 0,
	}, nil
}

func (s *Storage) ListEntries(ctx context.Context) ([]EntryMeta, error) {
	rows, err := s.queries.ListEntries(ctx)
	if err != nil {
		return []EntryMeta{}, fmt.Errorf("storage: listing entries: %w", err)
	}

	entries := []EntryMeta{}
	for _, row := range rows {
		entries = append(entries, EntryMeta{
			ID:         row.ID,
			Title:      row.Title,
			ModifiedAt: row.ModifiedAt,
		})
	}

	return entries, nil
}

func (s *Storage) UpdateEntry(ctx context.Context, e EntryRecord) error {
	params := repo.UpdateEntryParams{
		ID:         e.ID,
		Title:      e.Title,
		Ciphertext: e.Ciphertext,
		ModifiedAt: e.ModifiedAt,
	}

	if affected, err := s.queries.UpdateEntry(ctx, params); err != nil {
		return fmt.Errorf("storage: updating entry: %w", err)
	} else if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Storage) SetDeleted(ctx context.Context, id string, modifiedAt int64) error {
	if affected, err := s.queries.SetDeleted(ctx, repo.SetDeletedParams{
		ID:         id,
		ModifiedAt: modifiedAt,
	}); err != nil {
		return fmt.Errorf("storage: deleting entry: %w", err)
	} else if affected == 0 {
		return ErrNotFound
	}

	return nil
}

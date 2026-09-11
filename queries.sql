-- name: InitMeta :exec
INSERT INTO vault_meta (
  id,
  version,
  kdf_salt,
  kdf_time,
  kdf_memory,
  kdf_threads,
  kdf_key_len,
  encrypted_dek,
  verifier
)
VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetMeta :one
SELECT * FROM vault_meta WHERE id = 1;

-- name: InsertEntry :exec
INSERT INTO entries (
  id,
  title,
  ciphertext,
  modified_at,
  deleted
)
VALUES (?, ?, ?, ?, 0);

-- name: GetEntry :one
SELECT * FROM entries WHERE id = ? AND deleted = 0;

-- name: ListEntries :many
SELECT id, title, modified_at FROM entries WHERE deleted = 0;

-- name: UpdateEntry :exec
UPDATE entries
SET title = ?, ciphertext = ?, modified_at = ?
WHERE id = ? AND deleted = 0;

-- name: SetDeleted :exec
UPDATE entries
SET deleted = 1
WHERE id = ? AND deleted = 0;

-- name: EntryExists :one
SELECT
  EXISTS (
    SELECT 1
    FROM entries
    WHERE id = ? AND deleted = 0
  );

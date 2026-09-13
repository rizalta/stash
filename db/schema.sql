CREATE TABLE IF NOT EXISTS vault_meta (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  version INTEGER NOT NULL,
  kdf_salt BLOB NOT NULL,
  kdf_time INTEGER NOT NULL,
  kdf_memory INTEGER NOT NULL,
  kdf_threads INTEGER NOT NULL,
  kdf_key_len INTEGER NOT NULL,
  encrypted_dek BLOB NOT NULL,
  verifier BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS entries (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  ciphertext BLOB NOT NULL,
  modified_at INTEGER NOT NULL,
  deleted INTEGER NOT NULL DEFAULT 0
);

package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, registers as "sqlite"
)

// SQLite is the production Store, backed by an FTS5-enabled SQLite database.
type SQLite struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite index at path and applies the schema.
//
// Pragmas applied:
//   - journal_mode=WAL: concurrent readers while indexing writes.
//   - synchronous=NORMAL: safe with WAL, much faster than FULL for bulk writes.
//   - temp_store=MEMORY: keep FTS5 sort/merge scratch in RAM.
func Open(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %q: %w", path, err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA temp_store=MEMORY;",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("pragma %q: %w", pragma, err)
		}
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

// UpsertFile inserts or updates a file's metadata and indexed content.
//
// TODO M1:
//
//	BEGIN
//	  INSERT INTO files (path, mtime_ns, size, sha256, indexed_at)
//	  VALUES (?, ?, ?, ?, ?)
//	  ON CONFLICT(path) DO UPDATE SET
//	      mtime_ns=excluded.mtime_ns,
//	      size=excluded.size,
//	      sha256=excluded.sha256,
//	      indexed_at=excluded.indexed_at
//	  RETURNING id;
//
//	  DELETE FROM files_fts WHERE rowid = :id;
//	  INSERT INTO files_fts(rowid, path, content) VALUES (:id, :path, :content);
//	COMMIT
//
// Use s.db.BeginTx(ctx, nil) so Ctrl-C aborts cleanly.
func (s *SQLite) UpsertFile(ctx context.Context, f File, content []byte) error {
	return fmt.Errorf("not implemented yet -- TODO M1")
}

// DeleteFile removes a file (used when a path disappears between indexes).
//
// TODO M3.
func (s *SQLite) DeleteFile(ctx context.Context, path string) error {
	return fmt.Errorf("not implemented yet -- TODO M3")
}

// GetFile returns the metadata row for path, or sql.ErrNoRows if absent.
//
// TODO M4: needed for incremental re-index (skip files where mtime is unchanged).
func (s *SQLite) GetFile(ctx context.Context, path string) (File, error) {
	return File{}, fmt.Errorf("not implemented yet -- TODO M4")
}

// Search runs a BM25-ranked FTS5 query.
//
// TODO M2:
//
//	SELECT path,
//	       bm25(files_fts) AS score,
//	       snippet(files_fts, 1, '>>', '<<', '...', 16) AS snippet
//	  FROM files_fts
//	 WHERE files_fts MATCH ?
//	 ORDER BY score ASC
//	 LIMIT ?;
//
// Note on the trigram tokenizer: queries shorter than 3 characters won't
// match anything (FTS5 needs at least one full trigram). Document this in
// the search command help text.
func (s *SQLite) Search(ctx context.Context, query string, limit int) ([]SearchHit, error) {
	return nil, fmt.Errorf("not implemented yet -- TODO M2")
}

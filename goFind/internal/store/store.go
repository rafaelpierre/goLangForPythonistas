// Package store defines the persistence interface used by the indexer
// (writes) and the searcher (reads), plus the SQLite/FTS5 implementation.
//
// The interface lives next to the implementation here to keep the package
// self-contained, but the rest of the codebase depends on the *interface*,
// not the concrete type. That's what makes the indexer and the TUI testable
// with an in-memory fake -- see patterns/04_interfaceDI/.
package store

import (
	"context"
	"time"
)

// File is one row from the files table -- enough metadata to decide
// whether a path needs to be re-indexed on the next pass.
type File struct {
	ID        int64
	Path      string
	ModTime   time.Time
	Size      int64
	SHA256    string
	IndexedAt time.Time
}

// SearchHit is one BM25-ranked match returned by Search.
type SearchHit struct {
	Path    string
	Score   float64 // raw bm25() value; lower (more negative) = better
	Snippet string  // contextual excerpt with the match marked
}

// Store is the persistence contract. Accept this interface; return the
// concrete *SQLite. ("Accept interfaces, return structs.")
type Store interface {
	UpsertFile(ctx context.Context, f File, content []byte) error
	DeleteFile(ctx context.Context, path string) error
	GetFile(ctx context.Context, path string) (File, error)
	Search(ctx context.Context, query string, limit int) ([]SearchHit, error)
	Close() error
}

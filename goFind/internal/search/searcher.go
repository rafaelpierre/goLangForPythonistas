// Package search exposes the read-only query interface consumed by the
// search command and (in M5) the Bubble Tea TUI. Both will depend on
// this interface, not on *store.SQLite -- which is the whole point: when
// you wire up the TUI, no business logic moves.
package search

import (
	"context"

	"gofind/internal/store"
)

// Searcher is the read-only contract.
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]store.SearchHit, error)
}

// *store.SQLite already satisfies this interface (it has the right
// Search method signature). You don't need a separate adapter type --
// just pass your *store.SQLite where a Searcher is expected.

# goFind

Capstone project: a `grep`-substitute that indexes a codebase into SQLite
(FTS5 trigram tokenizer) and serves BM25-ranked search.

## Stack decisions

- **SQLite driver:** `modernc.org/sqlite` (pure Go, no CGO required).
- **Tokenizer:** FTS5 `trigram` -- works well for code identifiers and
  substring queries; no language-specific stemming.
- **CLI:** stdlib `flag` with a hand-rolled subcommand dispatcher.
- **Bonus TUI (M5):** Bubble Tea, reusing the same `Searcher` interface.

## Layout

```
goFind/
  go.mod
  cmd/gofind/main.go              # subcommand dispatch + signal handling
  internal/store/
    store.go                      # Store interface + types
    sqlite.go                     # *SQLite implementation
    schema.go                     # FTS5 trigram + files table DDL
  internal/index/
    indexer.go                    # Indexer + functional options + worker pool
    walker.go                     # WalkDir + binary detection + ignore globs
  internal/search/
    searcher.go                   # read-only Searcher interface
```

## Bootstrap

```
cd goFind
go get modernc.org/sqlite
go mod tidy
go build ./...
```

## Run

```
gofind index ./some/repo            # M1
gofind search "user repository"     # M2
gofind stats                        # M3
```

## Milestones

- [ ] **M1 — Index command.** Walker + worker pool + gitignore + binary skip
      + FTS5 trigram persist. Ctrl-C must abort cleanly without corrupting
      the DB.
      Patterns: concurrency, context, functional options, graceful shutdown.

- [ ] **M2 — Search command.** BM25-ranked results with snippet highlighting,
      `--json` and human output, `--limit` flag.
      Patterns: interfaces (Searcher), error wrapping at boundaries.

- [ ] **M3 — Stats + clean.** Row counts, on-disk size, drop deleted files.
      Patterns: SQLite specifics, idempotency.

- [ ] **M4 — Incremental re-index.** mtime-based `--update` first, then
      `watch` with fsnotify.
      Patterns: more concurrency + cancellation, decision-by-comparison.

- [ ] **M5 (bonus) — TUI.** Bubble Tea live search-as-you-type, preview pane,
      Enter to open `$EDITOR`. Reuses the same `Searcher` interface as M2 --
      if anything from `cmd/gofind/` has to move into a shared package to
      make M5 work, the abstraction was wrong.

## Implementation guide

Every public function with a `// TODO` comment has a sketch of what it should
do. Implement them in the order below -- each step gives you something
runnable to validate before moving on. The order is chosen so you always have
a working binary and a way to test the layer you just wrote.

### Phase A: get a write path working (no walker yet)

The point of this phase is to prove your `Store` works in isolation, before
adding concurrency on top.

1. **`internal/store/sqlite.go` — `UpsertFile`.** Implement the BEGIN /
   INSERT-OR-REPLACE / DELETE-FROM-FTS / INSERT-INTO-FTS / COMMIT sketch in
   the doc comment. Use `s.db.BeginTx(ctx, nil)` so cancellation works.
2. **Smoke test it from a tiny `_test.go` next to it.** Open an in-memory DB
   (`store.Open(":memory:")`), `UpsertFile` one fake file, then query the raw
   `files_fts` table directly with `s.db.QueryContext` and assert you get
   exactly one row back. This is the first time you'll know the schema is
   right.
3. **`internal/store/sqlite.go` — `Search`.** Use the SQL sketched in the
   doc comment. Extend the same test: insert a file containing `"GetUserByID"`,
   call `Search(ctx, "UserBy", 10)`, assert one hit. This proves the trigram
   tokenizer is working end to end.

After Phase A you have a tested storage layer. Everything that follows is
plumbing on top.

### Phase B: walk + index (M1)

4. **`internal/index/walker.go` — `walk()`.** Implement the
   `filepath.WalkDir` producer in the doc comment. No goroutines yet --
   write to `out chan<- string` synchronously. Easier to debug.
5. **`internal/index/indexer.go` — `Index()`.** Start single-threaded:
   one consumer goroutine reading from `paths`, calling `UpsertFile` for
   each. Get the whole pipeline working sequentially before adding
   concurrency. Aggregate `Stats` as you go.
6. **Add the worker pool.** Replace the single consumer with a
   `for i := 0; i < opts.Concurrency; i++` loop using `sync.WaitGroup`.
   Stats updates now need a mutex (or atomics, or a results channel
   aggregated in main goroutine -- pick one and own the choice).
7. **Wire `cmd/gofind/main.go` — `runIndex`.** Replace the placeholder
   with `store.Open` -> `index.New` -> `idx.Index` -> print the `Stats`.
   You should now have a working `gofind index <path>`.
8. **Test Ctrl-C.** Run `gofind index` against a large directory and hit
   Ctrl-C halfway through. The DB file should still open cleanly afterwards
   and `gofind index` should be re-runnable. If it isn't, your transaction
   boundaries are wrong.

### Phase C: search (M2)

9. **Wire `cmd/gofind/main.go` — `runSearch`.** Add `-db`, `-limit`, `-json`
   flags. Open the store, call `Search`, format the results. Two output
   modes: human (path, score, snippet on one line) and `--json` (one JSON
   object per result, newline-delimited -- not a JSON array, easier to pipe).
10. **Document the trigram minimum.** Queries shorter than 3 characters
    return zero hits; print a friendly message instead of an empty result.

### Phase D: stats + housekeeping (M3)

11. **`internal/store/sqlite.go` — `DeleteFile`.**
    `DELETE FROM files WHERE path = ?` then
    `DELETE FROM files_fts WHERE rowid = ?` in a transaction.
12. **Wire `runStats`.** `SELECT COUNT(*), SUM(size) FROM files`,
    `os.Stat(dbPath)` for on-disk size, last `MAX(indexed_at)`.
13. **Add `gofind clean`** as a fourth subcommand: walk the `files` table,
    drop rows whose `path` no longer exists on disk.

### Phase E: incremental re-index (M4)

14. **`internal/store/sqlite.go` — `GetFile`.** Single-row read; return
    `sql.ErrNoRows` unwrapped (callers will use `errors.Is`).
15. **Add an `--update` flag to `index`.** Before calling `UpsertFile`,
    look up the previous `mtime_ns` and `size`; skip if unchanged.
16. **`gofind watch`** as a fifth subcommand using `github.com/fsnotify/fsnotify`.
    Debounce events (200-500ms) so a save burst doesn't trigger 20 re-indexes.

### Phase F: TUI bonus (M5)

17. **Add `cmd/gofind/tui.go`** with a Bubble Tea program. Model holds the
    `search.Searcher`, current query, and result slice. `Update` runs the
    search on every keystroke (or debounced), `View` renders results.
    If you find yourself reaching into `internal/store` from the TUI,
    your interface is too narrow -- fix `Searcher` instead.
18. **Preview pane.** When the user moves the selection, read the file and
    show ~30 lines around the first match. (`bufio.Scanner` is fine here;
    don't load whole files.)

## Definition of done

- `go build ./...` succeeds with no warnings.
- `go vet ./...` clean.
- `go test ./...` has at least one test per package.
- `gofind index .` on this very repo finishes in under 5 seconds.
- `gofind search "<term>"` returns ranked results with snippets.
- Ctrl-C during indexing leaves a usable DB.

## What you'll have learned by the end

- A real Go module with internal packages and a clean public surface.
- A bounded worker-pool pipeline with backpressure and graceful shutdown.
- SQLite + FTS5 + BM25 from Go (and why trigram beats word tokenizers
  for code).
- Interface-based DI proven by the second consumer (the TUI) showing up.
- The full subcommand-CLI pattern with stdlib only.

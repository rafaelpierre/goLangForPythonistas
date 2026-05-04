// Package index walks a directory tree and feeds eligible source files
// into a Store. It is the worker-pool side of the project; the read side
// lives in internal/search.
package index

import (
	"context"
	"fmt"

	"gofind/internal/store"
)

// Options configures an Indexer. Use the With... helpers below.
//
// Functional options pattern -- see patterns/01_functionalOptions/.
type Options struct {
	Concurrency int
	MaxFileSize int64
	IgnoreGlobs []string
}

func defaultOptions() Options {
	return Options{
		Concurrency: 8,
		MaxFileSize: 1 << 20, // 1 MiB
		IgnoreGlobs: []string{".git", "node_modules", "vendor", "*.pb.go"},
	}
}

type Option func(*Options)

func WithConcurrency(n int) Option   { return func(o *Options) { o.Concurrency = n } }
func WithMaxFileSize(n int64) Option { return func(o *Options) { o.MaxFileSize = n } }
func WithIgnoreGlobs(globs ...string) Option {
	return func(o *Options) { o.IgnoreGlobs = append(o.IgnoreGlobs, globs...) }
}

// Indexer depends on the Store *interface*, so a unit test can hand it
// an in-memory fake without ever touching SQLite.
type Indexer struct {
	store store.Store
	opts  Options
}

func New(s store.Store, opts ...Option) *Indexer {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &Indexer{store: s, opts: o}
}

// Stats summarizes a single Index() run.
type Stats struct {
	Files   int
	Bytes   int64
	Skipped int
	Errors  int
}

// Index walks root and indexes every eligible file into the store.
//
// TODO M1: implement the worker-pool pipeline. A reasonable skeleton:
//
//	paths := make(chan string, opts.Concurrency)
//	var wg sync.WaitGroup
//
//	// producer
//	go func() {
//	    defer close(paths)
//	    walk(ctx, root, opts, paths) // see walker.go
//	}()
//
//	// consumers
//	var stats Stats
//	var mu sync.Mutex
//	for i := 0; i < opts.Concurrency; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        for p := range paths {
//	            if ctx.Err() != nil { return }
//	            // read file, hash it, build store.File, call UpsertFile.
//	            // Update stats under the mutex, or use atomics, or send
//	            // results on a channel and aggregate in main goroutine.
//	        }
//	    }()
//	}
//	wg.Wait()
//	return stats, ctx.Err()
//
// Don't forget:
//   - Honor ctx.Done() in BOTH producer and consumers (Ctrl-C must abort).
//   - Skip binaries (see walker.IsBinary).
//   - Skip files larger than opts.MaxFileSize.
//   - Skip paths matching opts.IgnoreGlobs.
//   - Wrap errors with file path context: fmt.Errorf("index %s: %w", path, err).
//
// Optional: try golang.org/x/sync/errgroup instead of sync.WaitGroup for
// nicer error propagation.
func (idx *Indexer) Index(ctx context.Context, root string) (Stats, error) {
	_ = root
	return Stats{}, fmt.Errorf("not implemented yet -- TODO M1")
}

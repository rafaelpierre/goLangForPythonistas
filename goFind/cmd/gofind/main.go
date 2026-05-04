// gofind: code search via SQLite FTS5 + BM25 ranking.
//
// This file is the entrypoint. Its job is *only* subcommand dispatch and
// signal handling -- all real logic lives under internal/.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const usage = `gofind: code search via SQLite FTS5 + BM25 ranking

Usage:
  gofind <command> [arguments]

Commands:
  index    Index a directory of source files
  search   Search the index with a BM25-ranked query
  stats    Show index statistics

Use "gofind <command> -h" for command-specific help.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	// Cancel the root context on Ctrl-C / SIGTERM so long-running commands
	// (indexing, watch) can shut down gracefully without corrupting the DB.
	// See patterns/06_gracefulShutdown/.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "index":
		err = runIndex(ctx, args)
	case "search":
		err = runSearch(ctx, args)
	case "stats":
		err = runStats(ctx, args)
	case "-h", "--help", "help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "gofind: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "gofind %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func runIndex(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	dbPath := fs.String("db", ".gofind.db", "path to the SQLite index database")
	concurrency := fs.Int("concurrency", 8, "number of indexer workers")
	maxSize := fs.Int64("max-size", 1<<20, "skip files larger than this many bytes")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: gofind index [flags] <path>")
	}
	root := fs.Arg(0)

	// TODO M1:
	//   1. s, err := store.Open(*dbPath); defer s.Close()
	//   2. idx := index.New(s,
	//          index.WithConcurrency(*concurrency),
	//          index.WithMaxFileSize(*maxSize),
	//      )
	//   3. stats, err := idx.Index(ctx, root)
	//   4. fmt.Printf("indexed %d files (%d bytes), %d skipped, %d errors\n", ...)
	_ = root
	_ = dbPath
	_ = concurrency
	_ = maxSize
	return fmt.Errorf("not implemented yet -- see TODO in cmd/gofind/main.go")
}

func runSearch(ctx context.Context, args []string) error {
	// TODO M2:
	//   - flags: -db, -limit, -json
	//   - open store, run searcher.Search(ctx, query, limit)
	//   - print path  score  snippet  per result (or JSON if -json)
	_ = ctx
	_ = args
	return fmt.Errorf("not implemented yet -- M2")
}

func runStats(ctx context.Context, args []string) error {
	// TODO M3:
	//   - open store, query COUNT(*) from files, sum(size), DB file size on disk.
	_ = ctx
	_ = args
	return fmt.Errorf("not implemented yet -- M3")
}

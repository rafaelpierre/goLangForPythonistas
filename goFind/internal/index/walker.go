package index

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IsBinary heuristically decides whether a file is binary by reading the
// first 512 bytes and checking for a NUL byte. This is what git uses
// internally, and it is good enough for code search -- no one wants to
// grep through a JPEG.
func IsBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	var buf [512]byte
	n, err := f.Read(buf[:])
	if err != nil && n == 0 {
		return false, err
	}
	return bytes.IndexByte(buf[:n], 0) >= 0, nil
}

// shouldSkip reports whether a directory entry should be skipped during
// the walk -- based on .gitignore-style globs and well-known noise dirs.
//
// For M1, simple filepath.Match against each segment is fine. For real
// .gitignore semantics (M3+), look at github.com/sabhiram/go-gitignore --
// but only adopt it once you understand what it is doing.
func shouldSkip(d fs.DirEntry, relPath string, globs []string) bool {
	name := d.Name()
	for _, g := range globs {
		// Match against the basename (e.g. "*.pb.go") ...
		if ok, _ := filepath.Match(g, name); ok {
			return true
		}
		// ... and any path segment (e.g. ".git", "node_modules").
		for _, seg := range strings.Split(relPath, string(filepath.Separator)) {
			if seg == g {
				return true
			}
		}
	}
	return false
}

// walk is the producer side of the indexing pipeline: it walks root and
// sends every eligible file path on out, until ctx is cancelled or the
// walk completes.
//
// TODO M1: implement using filepath.WalkDir. Roughly:
//
//	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
//	    if err != nil { return err }
//	    if ctx.Err() != nil { return ctx.Err() }
//	    rel, _ := filepath.Rel(root, path)
//	    if shouldSkip(d, rel, opts.IgnoreGlobs) {
//	        if d.IsDir() { return fs.SkipDir }
//	        return nil
//	    }
//	    if d.IsDir() { return nil }
//	    info, err := d.Info()
//	    if err != nil { return err }
//	    if info.Size() > opts.MaxFileSize { return nil }
//	    if bin, _ := IsBinary(path); bin { return nil }
//	    select {
//	    case out <- path:
//	    case <-ctx.Done():
//	        return ctx.Err()
//	    }
//	    return nil
//	})

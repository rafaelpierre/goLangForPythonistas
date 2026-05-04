// Topic: sync.Once, sync.Pool, and sync.Map
//
// The sync package has three types beyond WaitGroup and Mutex that show up
// constantly in production code but are easy to misuse:
//
//   sync.Once   -- run initialization exactly once, even under concurrent access.
//                  Python analog: module-level initialization (Python modules are
//                  singletons, but Go packages are not). Use Once for lazy,
//                  thread-safe initialization of expensive resources.
//
//   sync.Pool   -- a cache of reusable objects to reduce GC pressure.
//                  Python analog: there's no direct equivalent; Python relies on
//                  the GC. Use Pool for short-lived allocations in hot paths:
//                  byte buffers, JSON encoders, DB connections (though db/sql
//                  manages its own pool).
//
//   sync.Map    -- a concurrent map. NOT a drop-in for map+Mutex in all cases.
//                  Better than a plain map + RWMutex when: keys are written once
//                  and read many times, or when disjoint goroutines access disjoint
//                  keys (e.g., a per-goroutine cache). Worse when the write rate
//                  is high and uniform. Python analog: threading.local() or a
//                  plain dict with a Lock.
//
// Anti-pattern (from Python habits): treating sync.Once like a decorator or
// context manager. It's a value type -- embed it in a struct alongside the
// thing it initializes.
//
// Run: go run 01_sync_patterns.go

//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ===========================================================================
// 1. sync.Once: lazy, thread-safe singleton initialization
// ===========================================================================

// Config represents an expensive-to-load config (e.g., from a remote store).
type Config struct {
	APIURL  string
	Timeout time.Duration
}

// configHolder uses Once to initialize config exactly once, even if Load is
// called concurrently from hundreds of goroutines.
type configHolder struct {
	once   sync.Once
	config *Config
	err    error
}

func (h *configHolder) Load() (*Config, error) {
	h.once.Do(func() {
		// This function runs exactly once, no matter how many goroutines call Load.
		// All concurrent callers block here until the first one finishes.
		fmt.Println("  [configHolder] loading config (should print once)")
		time.Sleep(20 * time.Millisecond) // simulate remote fetch
		h.config = &Config{
			APIURL:  "https://api.example.com",
			Timeout: 5 * time.Second,
		}
		// In real code: h.err = some_fetch_error if the fetch failed.
		// Callers check both h.config and h.err after Load returns.
	})
	return h.config, h.err
}

// Global holder -- in practice this is often a package-level var.
var globalConfig configHolder

func demonstrateOnce() {
	fmt.Println("=== sync.Once: concurrent config load ===")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg, err := globalConfig.Load()
			if err != nil {
				fmt.Printf("  goroutine %d: error: %v\n", id, err)
				return
			}
			fmt.Printf("  goroutine %d: got config url=%s\n", id, cfg.APIURL)
		}(i)
	}
	wg.Wait()
	fmt.Println("  (config was loaded exactly once)")
}

// ===========================================================================
// 2. sync.Pool: recycled byte buffers to reduce allocations
//
// Real-world use: fmt/log internals, HTTP servers that format response bodies,
// JSON serializers -- anywhere you build a byte buffer on every request.
// ===========================================================================

// bufPool is a pool of *bytes.Buffer. Get() returns a recycled buffer (or a
// new one if the pool is empty). Put() returns it for future reuse.
// The GC may clear the pool under memory pressure -- that's intentional.
var bufPool = sync.Pool{
	New: func() any {
		// This runs only when the pool is empty.
		return new(bytes.Buffer)
	},
}

// formatJSON simulates building a JSON-ish response body.
// Without the pool, every call would allocate a new bytes.Buffer.
func formatJSON(id int, name string) string {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset() // IMPORTANT: always reset before use -- the buffer may have old data
	defer bufPool.Put(buf)

	fmt.Fprintf(buf, `{"id":%d,"name":%q,"ts":%d}`, id, name, time.Now().UnixMilli())
	return buf.String()
}

func demonstratePool() {
	fmt.Println("\n=== sync.Pool: byte buffer recycling ===")
	var wg sync.WaitGroup
	results := make([]string, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = formatJSON(idx, fmt.Sprintf("user-%d", idx))
		}(i)
	}
	wg.Wait()
	for _, r := range results {
		fmt.Println(" ", r)
	}

	// NOTE: sync.Pool is NOT a connection pool or a fixed-size cache.
	// For DB connections, use database/sql (manages its own pool).
	// For fixed caches, use a map+Mutex or a third-party LRU.
}

// ===========================================================================
// 3. sync.Map: concurrent map for read-heavy, write-once workloads
//
// Use case: a registry of metric counters keyed by name, written once on first
// use and read thousands of times per second thereafter.
// ===========================================================================

// MetricRegistry stores named counters using sync.Map.
type MetricRegistry struct {
	counters sync.Map // map[string]*int64  (values are pointers to allow atomic increment)
}

func (r *MetricRegistry) Increment(name string) {
	// LoadOrStore atomically: if the key exists, return its value;
	// otherwise, store the new value and return it.
	actual, _ := r.counters.LoadOrStore(name, new(int64))
	ptr := actual.(*int64)
	// Use atomic to increment safely -- sync.Map protects the MAP (key/value
	// storage), not the VALUES themselves.
	// We do a simple non-atomic increment here for readability; in production
	// use sync/atomic.AddInt64(ptr, 1).
	*ptr++
}

func (r *MetricRegistry) Get(name string) (int64, bool) {
	val, ok := r.counters.Load(name)
	if !ok {
		return 0, false
	}
	return *val.(*int64), true
}

func (r *MetricRegistry) Print() {
	r.counters.Range(func(key, value any) bool {
		fmt.Printf("  %s = %d\n", key.(string), *value.(*int64))
		return true // return false to stop iteration (like Python's break)
	})
}

func demonstrateSyncMap() {
	fmt.Println("\n=== sync.Map: concurrent metric registry ===")
	var registry MetricRegistry
	var wg sync.WaitGroup

	metrics := []string{"http.requests", "db.queries", "cache.hits", "http.requests", "db.queries"}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each goroutine randomly increments one of the known metrics.
			name := metrics[rand.Intn(len(metrics))]
			registry.Increment(name)
		}()
	}
	wg.Wait()

	fmt.Println("  final counts:")
	registry.Print()

	// When NOT to use sync.Map:
	// If your map has lots of writes (e.g., a cache with frequent invalidation),
	// a plain map + sync.RWMutex is usually faster because sync.Map uses
	// two internal maps and has overhead on the write path.
}

// ===========================================================================
// 4. Common pitfall: copying a sync type
//
// Once, Mutex, WaitGroup, Pool, Map -- NONE of these may be copied after first use.
// If you embed them in a struct, always pass the struct by pointer.
// ===========================================================================

func demonstrateCopyPitfall() {
	fmt.Println("\n=== Copy pitfall: always pass sync types by pointer ===")

	type Counter struct {
		mu    sync.Mutex
		value int
	}

	// CORRECT: pass by pointer
	inc := func(c *Counter) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.value++
	}

	c := &Counter{}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			inc(c)
		}()
	}
	wg.Wait()
	fmt.Printf("  counter value: %d (should be 5)\n", c.value)

	// WRONG (do not do this -- the go vet tool will catch it):
	// inc2 := func(c Counter) { // value receiver copies the Mutex!
	//     c.mu.Lock()           // locks a COPY of the mutex, not the original
	//     ...
	// }
	fmt.Println("  'go vet' catches mutex-copy bugs -- run it as part of CI")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	demonstrateOnce()
	demonstratePool()
	demonstrateSyncMap()
	demonstrateCopyPitfall()

	// TODO 1: Add an error path to configHolder.Load: on the FIRST call only,
	// simulate a 30% chance of the fetch failing (return a non-nil error and a
	// nil Config). On subsequent calls, Once still prevents re-running the init
	// -- which means a failed initialization is permanent. Demonstrate this by
	// calling Load 5 times concurrently when the "random fail" is forced on.
	// Then discuss: what would you do differently if you need to RETRY on failure?
	// (Hint: don't use Once -- use a mutex and a "loaded" bool instead.)

	// TODO 2: Benchmark formatJSON with and without the pool. Write two versions:
	//   BenchmarkFormatJSONPool   -- uses bufPool (current implementation)
	//   BenchmarkFormatJSONAlloc  -- allocates a new bytes.Buffer every call
	// Run with: go test -bench=. -benchmem ./07_syncPatterns/
	// (You'll need to move the benchmarks to a _test.go file to run them.)
	// What does the allocs/op column tell you?

	// TODO 3: Replace the non-atomic *ptr++ in MetricRegistry.Increment with a
	// proper atomic increment using sync/atomic. Import "sync/atomic" and use
	// atomic.AddInt64(ptr, 1). Then stress-test with 1000 goroutines each
	// incrementing the same metric 100 times -- the final count should be 100000.
	// With the current non-atomic version it will be less due to data races
	// (run with 'go run -race 01_sync_patterns.go' to see the race detector fire).

	// STRETCH: Implement a generic, thread-safe LRU cache using sync.Mutex and
	// a map + doubly-linked list (or use container/list from stdlib). The API:
	//   type LRU[K comparable, V any] struct { ... }
	//   func NewLRU[K comparable, V any](capacity int) *LRU[K, V]
	//   func (c *LRU[K, V]) Get(key K) (V, bool)
	//   func (c *LRU[K, V]) Put(key K, value V)
	// This combines sync patterns (Mutex) with generics (next exercise).
}

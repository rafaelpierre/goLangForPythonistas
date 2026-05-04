// Topic: Generics (Go 1.18+)
//
// Before generics, Go code that needed to work on multiple types had two options:
//   1. Duplicate the function for each type (verbose, unmaintainable)
//   2. Use 'any' (interface{}) and type-assert at runtime (loses compile-time safety)
//
// Python analog: Python is dynamically typed, so functions are naturally generic.
// In Go, generics give you the same flexibility WITH compile-time type checking.
// Think of it as Python duck typing, but the compiler verifies the "duck" at
// build time rather than at runtime.
//
// Syntax:
//   func Map[T, U any](slice []T, fn func(T) U) []U
//                ^^^^  type parameters, like Python TypeVar
//
// Constraints: a type parameter can require the type to satisfy an interface.
//   func Sum[T constraints.Ordered](slice []T) T
//            ^^^^^^^^^^^^^^^^^^^  T must support <, >, <=, >=, ==
//
// When to use generics:
//   - Utility functions that operate on containers (map, filter, reduce, Set, Stack)
//   - Type-safe wrappers (Result[T], Option[T], Pair[A, B])
//   - Data structures (linked list, tree, queue) that should work for any element type
//
// When NOT to use generics:
//   - When an interface is enough (e.g., io.Reader works fine without generics)
//   - When the behavior differs per type (that's what interfaces are for)
//   - "Just in case" -- generics add cognitive overhead; prefer concrete types until
//     the duplication actually hurts
//
// Run: go run 01_generics.go

//go:build ignore

package main

import (
	"cmp"
	"fmt"
	"strings"
)

// ===========================================================================
// 1. Generic slice utilities
//    (Go 1.21 added slices.Map/Filter/etc in x/exp -- roll your own here
//     to understand how they work)
// ===========================================================================

// Map transforms []T into []U using a function. Like Python's map().
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter returns elements of slice that satisfy the predicate. Like Python's filter().
func Filter[T any](slice []T, pred func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if pred(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce folds a slice into a single value. Like Python's functools.reduce().
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range slice {
		acc = fn(acc, v)
	}
	return acc
}

// Contains reports whether needle is in haystack. Works for any comparable type.
// comparable is a built-in constraint: any type that supports == and !=.
func Contains[T comparable](haystack []T, needle T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// Keys returns the keys of a map as a slice. Order is not guaranteed (maps are unordered).
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ===========================================================================
// 2. cmp.Ordered constraint: numeric and string operations
//
// cmp.Ordered (from the "cmp" package, Go 1.21) includes all types that
// support <, >, <=, >=. Equivalent to Python's requirement for __lt__ etc.
// ===========================================================================

// Min returns the smaller of two ordered values.
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two ordered values.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp restricts v to [lo, hi].
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	return Max(lo, Min(hi, v))
}

// ===========================================================================
// 3. Generic Result type (like Rust's Result<T,E> or Haskell's Either)
//
// Python analog: a tuple (value, error) or a Union type hint. Go already
// uses the (T, error) multiple-return pattern everywhere, but a Result type
// is useful when you need to pass a future/deferred result through a channel.
// ===========================================================================

// Result holds either a success value or an error, not both.
type Result[T any] struct {
	value T
	err   error
}

// Ok wraps a successful value.
func Ok[T any](v T) Result[T] {
	return Result[T]{value: v}
}

// Fail wraps an error.
func Fail[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// Unwrap returns the value or panics if there is an error.
// Use only when you are CERTAIN the result is Ok (like Rust's unwrap()).
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("Result.Unwrap: called on error result: %v", r.err))
	}
	return r.value
}

// Get returns (value, error) -- the idiomatic Go pattern.
func (r Result[T]) Get() (T, error) {
	return r.value, r.err
}

// IsOk reports whether the result holds a value.
func (r Result[T]) IsOk() bool { return r.err == nil }

// ===========================================================================
// 4. Generic Set
//
// Python analog: set(). Go's built-in map can be used as a set
// (map[T]struct{}), but wrapping it in a typed struct is cleaner.
// ===========================================================================

// Set is a generic, unordered collection of unique values.
type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{})}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

func (s *Set[T]) Add(v T)          { s.m[v] = struct{}{} }
func (s *Set[T]) Remove(v T)       { delete(s.m, v) }
func (s *Set[T]) Contains(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int          { return len(s.m) }

func (s *Set[T]) Slice() []T {
	out := make([]T, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}
	return out
}

// Union returns a new Set containing elements in either s or other.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for k := range s.m {
		result.Add(k)
	}
	for k := range other.m {
		result.Add(k)
	}
	return result
}

// Intersection returns a new Set containing only elements in BOTH s and other.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for k := range s.m {
		if other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

// ===========================================================================
// Main
// ===========================================================================

func main() {
	fmt.Println("=== Slice utilities ===")

	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}

	doubled := Map(nums, func(n int) int { return n * 2 })
	fmt.Println("doubled:", doubled)

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println("evens:", evens)

	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Println("sum:", sum)

	words := []string{"go", "python", "rust", "java"}
	upper := Map(words, strings.ToUpper)
	fmt.Println("upper:", upper)

	hasGo := Contains(words, "go")
	hasRuby := Contains(words, "ruby")
	fmt.Printf("has 'go': %v, has 'ruby': %v\n", hasGo, hasRuby)

	fmt.Println("\n=== Min / Max / Clamp ===")
	fmt.Println("Min(3, 7):", Min(3, 7))
	fmt.Println("Max(3.14, 2.72):", Max(3.14, 2.72))
	fmt.Println("Max(\"apple\", \"banana\"):", Max("apple", "banana"))
	fmt.Println("Clamp(15, 0, 10):", Clamp(15, 0, 10))
	fmt.Println("Clamp(-3, 0, 10):", Clamp(-3, 0, 10))

	fmt.Println("\n=== Result[T] ===")
	safeDiv := func(a, b float64) Result[float64] {
		if b == 0 {
			return Fail[float64](fmt.Errorf("division by zero"))
		}
		return Ok(a / b)
	}

	r1 := safeDiv(10, 3)
	if v, err := r1.Get(); err == nil {
		fmt.Printf("10/3 = %.4f\n", v)
	}

	r2 := safeDiv(5, 0)
	if _, err := r2.Get(); err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("\n=== Set[T] ===")
	a := NewSet("go", "python", "rust")
	b := NewSet("python", "rust", "java", "c++")

	fmt.Println("a:", a.Slice())
	fmt.Println("b:", b.Slice())
	fmt.Println("a contains 'go':", a.Contains("go"))
	fmt.Println("union len:", a.Union(b).Len())
	fmt.Println("intersection:", a.Intersection(b).Slice())

	intSet := NewSet(1, 2, 3, 2, 1) // duplicates dropped
	fmt.Println("int set len (should be 3):", intSet.Len())

	fmt.Println("\n=== Keys ===")
	inventory := map[string]int{"apples": 5, "bananas": 3, "cherries": 12}
	fmt.Println("keys:", Keys(inventory))

	// TODO 1: Implement 'GroupBy[T any, K comparable](slice []T, key func(T) K) map[K][]T'
	// that partitions a slice into groups by the result of key(). Example:
	//   words := []string{"go", "python", "rust", "get", "put"}
	//   GroupBy(words, func(w string) int { return len(w) })
	//   // => map[2:["go"] 3:["get","put"] 4:["rust"] 6:["python"]]
	// Then call it in main and print the result.

	// TODO 2: Implement 'Chunk[T any](slice []T, size int) [][]T' that splits
	// a slice into sub-slices of at most 'size' elements. The last chunk may
	// be smaller. Example: Chunk([]int{1,2,3,4,5}, 2) => [[1,2],[3,4],[5]].
	// Return an error (or panic?) if size <= 0. Justify your choice.

	// TODO 3: Add a 'Difference' method to Set[T] that returns elements in s
	// but NOT in other (like Python's set.difference()). Then add a 'SymmetricDiff'
	// method that returns elements in either set but not both. Write a small
	// main demo that verifies correctness with integer sets.

	// TODO 4: Implement 'Must[T any](v T, err error) T' -- a helper that panics
	// if err != nil, otherwise returns v. This lets you write:
	//   cfg := Must(loadConfig("app.yaml"))
	// instead of:
	//   cfg, err := loadConfig("app.yaml")
	//   if err != nil { panic(err) }
	// When is Must appropriate? (Hint: same rule as panic -- program startup only.)

	// STRETCH: Implement a generic, bounded Stack[T any] with:
	//   func NewStack[T any](capacity int) *Stack[T]
	//   func (s *Stack[T]) Push(v T) error   // error if full
	//   func (s *Stack[T]) Pop() (T, error)  // error if empty
	//   func (s *Stack[T]) Peek() (T, bool)  // (top, false if empty)
	//   func (s *Stack[T]) Len() int
	// Make it safe for concurrent use with a sync.Mutex. Then write a small
	// table-driven test in a _test.go file. (You won't be able to use
	// //go:build ignore on the test file -- put both files in the same directory
	// and run `go test ./08_generics/` -- but wait, there's no go.mod! What
	// happens? Investigate and document the workaround.)
}

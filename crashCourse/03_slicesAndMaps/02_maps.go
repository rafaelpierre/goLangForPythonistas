// Topic: Maps
//
// Python dict vs Go map:
//   Python: d = {"key": "value"}  -- any hashable key, any value, ordered (3.7+)
//   Go:     m := map[string]int{"a": 1}  -- typed keys AND values, unordered
//
// Key ideas:
//   - Maps are reference types (like Python dicts). Passing to a function
//     lets the function mutate the original -- no copy is made.
//   - Reading a missing key returns the zero value (NOT a KeyError like Python).
//     Use the two-value form to distinguish "missing" from "zero":
//       val, ok := m[key]
//   - Delete with delete(m, key) -- like Python's del d[key].
//   - Iterating order is intentionally randomized per run (do not rely on it).
//   - make(map[K]V) or a literal; never use a nil map for writes.
//
// Run: go run 02_maps.go

//go:build ignore

package main

import (
	"fmt"
	"sort"
)

func main() {
	// Map literal
	population := map[string]int{
		"São Paulo":  12_325_000,
		"Rio":        6_748_000,
		"Brasília":   3_094_000,
	}

	// Read a value
	fmt.Println("São Paulo population:", population["São Paulo"])

	// Two-value form: check existence (Python: d.get(key) or key in d)
	val, ok := population["Manaus"]
	if !ok {
		fmt.Println("Manaus not in map; zero value was:", val)
	}

	// Add / update (same syntax as Python d[key] = value)
	population["Manaus"] = 2_063_000
	fmt.Println("After insert, Manaus:", population["Manaus"])

	// Delete (Python: del d[key])
	delete(population, "Brasília")
	fmt.Println("After delete:", population)

	// Iterate -- order is random; sort keys if you need determinism
	cities := make([]string, 0, len(population))
	for city := range population {
		cities = append(cities, city)
	}
	sort.Strings(cities)
	for _, city := range cities {
		fmt.Printf("  %-15s %d\n", city, population[city])
	}

	// GOTCHA: writing to a nil map panics.
	// var bad map[string]int  -- this is nil
	// bad["key"] = 1          -- PANIC: assignment to entry in nil map
	// Always initialize: bad := make(map[string]int)

	// Maps as frequency counters (classic Python Counter pattern)
	words := []string{"go", "python", "go", "rust", "go", "python"}
	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++ // zero value (0) is returned for missing keys, so this is safe
	}
	fmt.Println("word frequencies:", freq)

	// TODO 1: Write a function wordCount(s string) map[string]int that splits
	// a sentence into words and returns a frequency map.
	// Test: wordCount("the cat sat on the mat") should give
	//       map[cat:1 mat:1 on:1 sat:1 the:2]
	// Hint: use strings.Fields(s) to split on whitespace.

	// TODO 2: Write a function invertMap(m map[string]int) map[int]string
	// that swaps keys and values. What happens if two keys have the same value?
	// Print a note about it in a comment.

	// TODO 3: Given a slice of strings, use a map to deduplicate it and return
	// a slice of unique values. Order does not matter.
	// Hint: map[string]struct{} is the idiomatic "set" in Go.
}

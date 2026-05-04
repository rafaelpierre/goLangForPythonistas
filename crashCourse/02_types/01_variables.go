// Topic: Variables, Types, and Zero Values
//
// Python lets you write: x = 42  (dynamic, no declaration needed)
// Go requires you to declare the type -- but often infers it for you.
//
// Key ideas:
//   - Every type has a "zero value" (the default when you don't initialize).
//     int -> 0, float64 -> 0.0, string -> "", bool -> false, pointer -> nil
//   - Two declaration styles:
//       var name string = "alice"   // explicit type
//       name := "alice"             // short declaration (type inferred, inside func only)
//   - Constants use 'const'. Unlike Python, Go constants are truly compile-time.
//   - Go is STATICALLY typed: you cannot assign an int to a string variable.
//
// Run: go run 01_variables.go

//go:build ignore

package main

import "fmt"

func main() {
	// --- Zero values ---
	// Declare variables without initializing them. Print their zero values.
	var i int
	var f float64
	var s string
	var flag bool
	fmt.Printf("Zero values: int=%d, float64=%f, string=%q, bool=%t\n", i, f, s, flag)

	// --- Short declaration (most common inside functions) ---
	city := "São Paulo"
	year := 2026
	fmt.Printf("city=%s, year=%d\n", city, year)

	// --- Constants ---
	const maxRetries = 3
	const pi = 3.14159
	fmt.Printf("maxRetries=%d, pi=%f\n", maxRetries, pi)

	// --- Multiple assignment (like Python tuple unpacking, but explicit) ---
	x, y := 10, 20
	fmt.Printf("x=%d, y=%d\n", x, y)

	// --- Type conversions are EXPLICIT in Go (no implicit coercion) ---
	// Python:  total = 1 + 2.5  (works silently)
	// Go:      you must convert:
	count := 7
	ratio := float64(count) / 2.0
	fmt.Printf("ratio=%f\n", ratio)

	// TODO 1: Declare a variable 'temperature' of type float64 using 'var'.
	// Do NOT initialize it. Print it and observe the zero value.

	var temperature float64
	fmt.Printf("Temperature=%f", temperature)

	// TODO 2: Use a short declaration to store your name in a variable,
	// then print: "Hello, my name is <name>"

	name := "Rafael"
	fmt.Printf("Hello, my name is %s\n", name)

	// TODO 3: Declare two int variables 'a' and 'b', assign them any values,
	// then swap them IN ONE LINE using multiple assignment (no temp variable).
	// Print both before and after the swap.

	var a, b int = 1, 2
	fmt.Printf("a: %d, b: %d\n", a, b)
	a, b = b, a
	fmt.Printf("a: %d, b: %d\n", a, b)
}

// Topic: Error Handling
//
// Python raises exceptions; callers catch them with try/except.
// Go returns errors as plain values -- the caller MUST check them explicitly.
// This is not a limitation; it makes error paths visible and intentional.
//
// Key ideas:
//   - 'error' is a built-in interface: type error interface { Error() string }
//   - Return (result, error) from functions that can fail. The convention is:
//     if err != nil, the result value is meaningless.
//   - fmt.Errorf("...: %w", err) wraps an error (adds context while preserving
//     the original for errors.Is / errors.As inspection).
//   - errors.Is(err, target) checks if target is anywhere in the error chain.
//     Python equivalent: isinstance(exc, SomeException)
//   - errors.As(err, &target) extracts a concrete error type from the chain.
//   - Sentinel errors (var ErrNotFound = errors.New("not found")) are the
//     Go equivalent of custom exception types for known conditions.
//   - panic/recover exist but are NOT for normal error handling. Use them only
//     for truly unrecoverable situations (programmer bugs, not runtime failures).
//
// Run: go run 01_errors.go

//go:build ignore

package main

import (
	"errors"
	"fmt"
	"strconv"
)

// --- Sentinel error: a known, named error value ---
// Python equivalent: class NotFoundError(Exception): pass
var ErrNotFound = errors.New("not found")
var ErrPermission = errors.New("permission denied")

// --- Custom error type: carries extra context ---
// Python equivalent: class ValidationError(Exception): def __init__(self, field, msg): ...
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Message)
}

// --- Functions that return errors ---

func findUser(id int) (string, error) {
	users := map[int]string{1: "alice", 2: "bob"}
	name, ok := users[id]
	if !ok {
		// Return the sentinel -- callers can check with errors.Is
		return "", ErrNotFound
	}
	return name, nil
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Message: "must be non-negative"}
	}
	if age > 150 {
		return &ValidationError{Field: "age", Message: "unrealistically large"}
	}
	return nil // nil == no error
}

// Wrapping errors with context (use %w to preserve the chain)
func loadUser(id int) (string, error) {
	name, err := findUser(id)
	if err != nil {
		// Wrap: adds "loadUser:" prefix while keeping ErrNotFound checkable.
		return "", fmt.Errorf("loadUser(id=%d): %w", id, err)
	}
	return name, nil
}

func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parsePositiveInt: %w", err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("parsePositiveInt: value must be positive, got %d", n)
	}
	return n, nil
}

func main() {
	// --- Basic error check ---
	name, err := findUser(1)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("found user:", name)
	}

	// Missing user
	_, err = findUser(99)
	if err != nil {
		fmt.Println("error:", err)
	}

	// --- errors.Is: check for a specific sentinel anywhere in the chain ---
	_, err = loadUser(99) // wrapped error
	fmt.Println("raw error:", err)
	fmt.Println("is ErrNotFound?", errors.Is(err, ErrNotFound)) // true -- unwraps chain

	// --- errors.As: extract a concrete type from the chain ---
	err = validateAge(-5)
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		fmt.Printf("Validation problem -- field: %s, message: %s\n", valErr.Field, valErr.Message)
	}

	err = validateAge(25)
	fmt.Println("valid age error:", err) // nil

	// --- Wrapping strconv errors ---
	n, err := parsePositiveInt("abc")
	if err != nil {
		fmt.Println("parse error:", err)
	}
	n, err = parsePositiveInt("-3")
	if err != nil {
		fmt.Println("parse error:", err)
	}
	n, err = parsePositiveInt("42")
	if err == nil {
		fmt.Println("parsed:", n)
	}

	// --- ANTI-PATTERN: do not use panic for normal error handling ---
	// Python engineers sometimes reach for panic like they'd raise an exception.
	// In Go, panic means "the program is in an unrecoverable state." Use it for:
	//   - Programmer errors (nil pointer you should have checked)
	//   - Initialization failures that make the program unusable
	// NOT for: user input errors, network failures, file not found, etc.

	// TODO 1: Write a function 'divide(a, b float64) (float64, error)'.
	// Return a descriptive error if b == 0. Use it and handle the error.

	// TODO 2: Define a sentinel 'ErrEmptyInput = errors.New("empty input")'.
	// Write a function 'normalize(s string) (string, error)' that returns
	// ErrEmptyInput if s is blank (after trimming), otherwise returns
	// strings.ToLower(strings.TrimSpace(s)).
	// Verify with errors.Is.

	// TODO 3: Write a function 'mustParseInt(s string) int' that panics
	// with a descriptive message if the string is not a valid integer.
	// This is the correct use of panic: it is a programmer error to call
	// mustParseInt with a non-integer string (e.g. from a hardcoded config).
	// Use recover() in a test to prove the panic can be caught.
}

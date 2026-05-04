// Topic: Error Wrapping and Sentinel Hierarchies
//
// You saw basic errors in crashCourse/06_errors. This exercise goes deeper:
//   - Building a layered error hierarchy for a real service
//   - Wrapping errors across multiple call-stack layers so the chain is inspectable
//   - errors.Is for sentinel matching, errors.As for type extraction
//   - Collecting multiple errors with a custom multi-error type
//   - When to wrap vs when to return a new error
//
// Python analog: a custom exception hierarchy plus chained exceptions (raise X from Y).
// Go's chain is explicit -- every layer must call fmt.Errorf("%w", err) to preserve it.
// Unlike Python, there are no stack traces attached automatically; the chain IS the trace.
//
// Real-world use: any service layer (HTTP handler -> business logic -> repository -> DB)
// where you want the caller to distinguish "not found" from "permission denied" from
// "connection failed" without parsing error strings.
//
// Run: go run 01_error_wrapping.go

//go:build ignore

package main

import (
	"errors"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Sentinel error hierarchy
//
// Convention: define package-level sentinel vars that callers check with errors.Is.
// These act like Python's custom exception classes, but lighter.
// ---------------------------------------------------------------------------

var (
	// Base domain errors -- used as-is or wrapped with additional context
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrConflict   = errors.New("conflict")
	ErrBadRequest = errors.New("bad request")

	// Storage-layer sentinel -- the service layer wraps this; callers should
	// not depend on ErrDBUnavailable leaking out of the storage layer.
	ErrDBUnavailable = errors.New("database unavailable")
)

// ---------------------------------------------------------------------------
// Typed errors for richer context
// ---------------------------------------------------------------------------

// NotFoundError carries the resource type and ID so the caller can build a
// useful API response without parsing error strings.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %q not found", e.Resource, e.ID)
}

// Unwrap makes NotFoundError part of the ErrNotFound sentinel chain.
// errors.Is(err, ErrNotFound) returns true when err is a *NotFoundError.
func (e *NotFoundError) Unwrap() error { return ErrNotFound }

// ValidationError collects field-level validation failures.
type ValidationError struct {
	Fields map[string]string // field name -> problem description
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for field, msg := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

func (e *ValidationError) Unwrap() error { return ErrBadRequest }

// ---------------------------------------------------------------------------
// Multi-error: collect several independent errors
//
// Useful when you want to validate a whole struct at once rather than
// returning on the first failure. Python equivalent: accumulate into a list.
// ---------------------------------------------------------------------------

type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	msgs := make([]string, len(m.Errors))
	for i, e := range m.Errors {
		msgs[i] = e.Error()
	}
	return fmt.Sprintf("%d error(s): %s", len(m.Errors), strings.Join(msgs, "; "))
}

// Unwrap returns the slice so that errors.Is / errors.As can walk all of them.
// This is the Go 1.20+ multi-error unwrap protocol: return []error.
func (m *MultiError) Unwrap() []error { return m.Errors }

// appendError adds err to a *MultiError, creating one if dst is nil.
// Returns nil if err is nil (so callers can chain: dst = appendError(dst, validate(x))).
func appendError(dst *MultiError, err error) *MultiError {
	if err == nil {
		return dst
	}
	if dst == nil {
		dst = &MultiError{}
	}
	dst.Errors = append(dst.Errors, err)
	return dst
}

// ---------------------------------------------------------------------------
// Fake layered service: storage -> repository -> service
// ---------------------------------------------------------------------------

// --- Storage layer: speaks DB ---

type userRecord struct {
	id    string
	email string
	role  string
}

var fakeDB = map[string]userRecord{
	"u1": {id: "u1", email: "alice@example.com", role: "admin"},
	"u2": {id: "u2", email: "bob@example.com", role: "viewer"},
}

func dbFetchUser(id string) (userRecord, error) {
	rec, ok := fakeDB[id]
	if !ok {
		// Storage layer uses its own sentinel -- the repo layer will decide
		// whether to expose this or translate it.
		return userRecord{}, fmt.Errorf("dbFetchUser(%q): %w", id, ErrDBUnavailable)
	}
	return rec, nil
}

// --- Repository layer: translates storage errors into domain errors ---

type User struct {
	ID    string
	Email string
	Role  string
}

func getUser(id string) (User, error) {
	if id == "" {
		return User{}, fmt.Errorf("getUser: %w", &ValidationError{
			Fields: map[string]string{"id": "must not be empty"},
		})
	}

	rec, err := dbFetchUser(id)
	if err != nil {
		if errors.Is(err, ErrDBUnavailable) {
			// Translate storage error -> domain error, but keep the original
			// wrapped so that the storage-layer details are preserved for logs.
			return User{}, fmt.Errorf("getUser(%q): %w: %w", id, &NotFoundError{Resource: "user", ID: id}, err)
		}
		return User{}, fmt.Errorf("getUser(%q): %w", id, err)
	}

	return User{ID: rec.id, Email: rec.email, Role: rec.role}, nil
}

// --- Service layer: enforces business rules ---

func deleteUser(requesterID, targetID string) error {
	requester, err := getUser(requesterID)
	if err != nil {
		return fmt.Errorf("deleteUser: look up requester: %w", err)
	}

	if requester.Role != "admin" {
		return fmt.Errorf("deleteUser: user %q is not an admin: %w", requesterID, ErrForbidden)
	}

	_, err = getUser(targetID)
	if err != nil {
		return fmt.Errorf("deleteUser: look up target: %w", err)
	}

	// (would actually delete here)
	fmt.Printf("user %q deleted by %q\n", targetID, requesterID)
	return nil
}

func createUser(id, email, role string) error {
	var errs *MultiError

	if id == "" {
		errs = appendError(errs, fmt.Errorf("id: must not be empty"))
	}
	if !strings.Contains(email, "@") {
		errs = appendError(errs, fmt.Errorf("email: %q is not a valid email", email))
	}
	if role != "admin" && role != "viewer" {
		errs = appendError(errs, fmt.Errorf("role: %q is unknown; must be admin or viewer", role))
	}

	if errs != nil {
		return fmt.Errorf("createUser: %w", errs)
	}

	fmt.Printf("user created: id=%s email=%s role=%s\n", id, email, role)
	return nil
}

func main() {
	fmt.Println("=== getUser: not found ===")
	_, err := getUser("u99")
	fmt.Println("raw error:", err)

	// errors.Is walks the whole chain, including through fmt.Errorf %w wrappers
	fmt.Println("is ErrNotFound?", errors.Is(err, ErrNotFound))
	fmt.Println("is ErrDBUnavailable?", errors.Is(err, ErrDBUnavailable))

	// errors.As extracts the first *NotFoundError anywhere in the chain
	var nfe *NotFoundError
	if errors.As(err, &nfe) {
		fmt.Printf("  -> resource=%q id=%q\n", nfe.Resource, nfe.ID)
	}

	fmt.Println("\n=== getUser: validation failure ===")
	_, err = getUser("")
	fmt.Println("raw error:", err)
	fmt.Println("is ErrBadRequest?", errors.Is(err, ErrBadRequest))

	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("  -> fields=%v\n", ve.Fields)
	}

	fmt.Println("\n=== deleteUser: permission denied ===")
	err = deleteUser("u2", "u1") // bob (viewer) tries to delete alice
	fmt.Println("raw error:", err)
	fmt.Println("is ErrForbidden?", errors.Is(err, ErrForbidden))

	fmt.Println("\n=== deleteUser: success ===")
	err = deleteUser("u1", "u2") // alice (admin) deletes bob
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("\n=== createUser: multi-error ===")
	err = createUser("", "not-an-email", "superuser")
	fmt.Println("raw error:", err)

	var me *MultiError
	if errors.As(err, &me) {
		fmt.Printf("  -> %d individual errors:\n", len(me.Errors))
		for i, e := range me.Errors {
			fmt.Printf("     [%d] %v\n", i, e)
		}
	}

	// TODO 1: Add an ErrRateLimit sentinel and a RateLimitError struct that
	// carries a RetryAfter time.Duration. Make RateLimitError.Unwrap() return
	// ErrRateLimit. Add it as a possible return from a new function
	// checkRateLimit(userID string) error that rate-limits "u2" (returns the
	// error) but allows "u1". Verify with errors.Is and errors.As.

	// TODO 2: Add a function 'updateUser(id, email, role string) error' that
	// calls getUser, validates the new values with appendError, and returns a
	// combined error. It should be possible to check errors.Is(err, ErrNotFound)
	// to distinguish "user does not exist" from "validation failed".

	// TODO 3: Write a helper 'firstOfType[T error](err error) (T, bool)' using
	// generics (Go 1.18+) that wraps errors.As for any error type T. Then use
	// it instead of the manual 'var nfe *NotFoundError; errors.As(...)' pattern.
}

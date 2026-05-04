// Topic: context.Context Propagation
//
// context.Context is Go's standard mechanism for:
//   1. Cancellation: tell downstream calls "the caller gave up, stop working"
//   2. Deadlines/timeouts: "stop working after N milliseconds"
//   3. Request-scoped values: carry a trace ID or auth token through a call chain
//      without threading it as an explicit parameter everywhere
//
// Python analog:
//   - Cancellation: asyncio.CancelledError / task.cancel() -- but that only
//     works in async code. Go's context works across regular function calls.
//   - Request-scoped values: threading.local() or contextvars.ContextVar --
//     same idea, different mechanism.
//
// The rule: if a function does I/O, calls another service, or can take a
// non-trivial amount of time, its FIRST parameter should be ctx context.Context.
// This is a strong convention, not a language rule.
//
// Anti-pattern (from Python habits): storing a context in a struct field and
// reusing it across multiple requests. A context is per-request/per-call -- it
// dies when the call dies.
//
// Run: go run 01_context_propagation.go

//go:build ignore

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// ---------------------------------------------------------------------------
// Request-scoped value keys
//
// Use unexported concrete types (not plain strings) as keys to prevent
// collisions with other packages putting values in the same context.
// ---------------------------------------------------------------------------

type contextKey string

const (
	keyTraceID contextKey = "traceID"
	keyUserID  contextKey = "userID"
)

func withTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyTraceID, id)
}

func traceIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(keyTraceID).(string)
	return id, ok
}

func withUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

func userIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(keyUserID).(string)
	return id, ok
}

// logf prefixes every log line with the trace ID from context.
// This is how distributed tracing correlation IDs work in practice.
func logf(ctx context.Context, format string, args ...any) {
	traceID, _ := traceIDFromContext(ctx)
	fmt.Printf("[trace=%s] "+format+"\n", append([]any{traceID}, args...)...)
}

// ---------------------------------------------------------------------------
// Fake layered service: handler -> service -> repository -> "database"
// ---------------------------------------------------------------------------

// dbQuery simulates a slow database query that respects cancellation.
// Real database drivers (pgx, database/sql) accept a context and cancel
// the in-flight query when it's done.
func dbQuery(ctx context.Context, query string) (string, error) {
	// Simulate variable latency
	latency := time.Duration(rand.Intn(200)+50) * time.Millisecond
	logf(ctx, "dbQuery: starting %q (will take ~%v)", query, latency)

	select {
	case <-time.After(latency):
		// Query "completed" before cancellation
		logf(ctx, "dbQuery: %q completed", query)
		return "result-of-" + query, nil
	case <-ctx.Done():
		// Context was cancelled or timed out before the query finished.
		// ctx.Err() returns context.Canceled or context.DeadlineExceeded.
		logf(ctx, "dbQuery: %q aborted: %v", query, ctx.Err())
		return "", fmt.Errorf("dbQuery %q: %w", query, ctx.Err())
	}
}

// userRepository is the data-access layer.
func userRepository(ctx context.Context, userID string) (map[string]string, error) {
	logf(ctx, "userRepository: fetching user %q", userID)

	result, err := dbQuery(ctx, "SELECT * FROM users WHERE id="+userID)
	if err != nil {
		return nil, fmt.Errorf("userRepository: %w", err)
	}
	return map[string]string{"id": userID, "data": result}, nil
}

// userService is the business logic layer.
func userService(ctx context.Context, userID string) (map[string]string, error) {
	logf(ctx, "userService: processing request for user %q", userID)

	// Check context before doing expensive work (eager cancellation check)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("userService: context already done: %w", err)
	}

	user, err := userRepository(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("userService: %w", err)
	}

	logf(ctx, "userService: enriching user data")
	// Simulate some extra work (e.g., calling another service for permissions)
	_, err = dbQuery(ctx, "SELECT * FROM permissions WHERE user_id="+userID)
	if err != nil {
		return nil, fmt.Errorf("userService: permissions: %w", err)
	}

	return user, nil
}

// httpHandler simulates an HTTP request handler.
// In a real net/http handler, r.Context() provides the per-request context.
// The server cancels it when the client disconnects.
func httpHandler(requestUserID string, timeout time.Duration) {
	// Attach request-scoped values: in practice from middleware
	ctx := context.Background()
	ctx = withTraceID(ctx, fmt.Sprintf("trace-%04d", rand.Intn(10000)))
	ctx = withUserID(ctx, requestUserID)

	// Set a deadline for the whole request
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // always defer cancel to release resources even on success

	logf(ctx, "httpHandler: starting (timeout=%v)", timeout)

	user, err := userService(ctx, requestUserID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logf(ctx, "httpHandler: request timed out: %v", err)
			fmt.Println("-> HTTP 503 Service Unavailable (timeout)")
		} else if errors.Is(err, context.Canceled) {
			logf(ctx, "httpHandler: request cancelled: %v", err)
			fmt.Println("-> HTTP 499 Client Closed Request")
		} else {
			logf(ctx, "httpHandler: internal error: %v", err)
			fmt.Println("-> HTTP 500 Internal Server Error")
		}
		return
	}

	logf(ctx, "httpHandler: success: %v", user)
	fmt.Println("-> HTTP 200 OK")
}

// ---------------------------------------------------------------------------
// Manual cancellation demo
// ---------------------------------------------------------------------------

// longJob does work in a loop and checks context cancellation on each iteration.
func longJob(ctx context.Context, name string) error {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("longJob %q cancelled after %d iterations: %w", name, i, ctx.Err())
		default:
			// not cancelled yet -- do a unit of work
			logf(ctx, "longJob %q: iteration %d", name, i)
			time.Sleep(50 * time.Millisecond)
			if i >= 4 {
				logf(ctx, "longJob %q: finished naturally", name)
				return nil
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Case 1: request completes within timeout ===")
	httpHandler("u1", 500*time.Millisecond)

	fmt.Println("\n=== Case 2: request times out ===")
	httpHandler("u2", 100*time.Millisecond)

	fmt.Println("\n=== Case 3: manual cancellation ===")
	ctx := context.Background()
	ctx = withTraceID(ctx, "trace-manual")
	ctx, cancel := context.WithCancel(ctx)

	// Cancel after 120ms -- before longJob's 4th iteration at 50ms * 4 = 200ms
	go func() {
		time.Sleep(120 * time.Millisecond)
		logf(ctx, "main: cancelling context")
		cancel()
	}()

	err := longJob(ctx, "etl-job")
	if err != nil {
		fmt.Println("longJob error:", err)
		fmt.Println("is Canceled?", errors.Is(err, context.Canceled))
	}

	fmt.Println("\n=== Case 4: value propagation ===")
	baseCtx := context.Background()
	baseCtx = withTraceID(baseCtx, "trace-abc123")
	baseCtx = withUserID(baseCtx, "u42")

	// Derived contexts inherit values from their parent
	childCtx, childCancel := context.WithTimeout(baseCtx, 1*time.Second)
	defer childCancel()

	if tid, ok := traceIDFromContext(childCtx); ok {
		fmt.Println("trace ID visible in child context:", tid)
	}
	if uid, ok := userIDFromContext(childCtx); ok {
		fmt.Println("user ID visible in child context:", uid)
	}

	// TODO 1: Write a function 'fetchParallel(ctx context.Context, ids []string)'
	// that launches one goroutine per ID (calling dbQuery for each), collects
	// results into a slice, and cancels remaining goroutines as soon as ANY
	// goroutine fails. Use a context derived with WithCancel passed to each
	// goroutine. Hint: errgroup from golang.org/x/sync is the stdlib-adjacent
	// tool for this -- but try implementing it manually with channels first.

	// TODO 2: Add a 'WithRequestID' function that embeds a UUID-like string
	// in the context. Add a middleware-style function:
	//   func attachRequestID(next func(ctx context.Context)) func(ctx context.Context)
	// that wraps the handler, generating and attaching a fresh request ID before
	// calling next. Call it from main and verify the ID appears in logf output.

	// TODO 3: Refactor longJob to accept a 'checkInterval time.Duration' parameter
	// and replace the hard-coded 50ms sleep. Then demonstrate that a very small
	// checkInterval (5ms) makes the job more responsive to cancellation than a
	// large one (200ms) -- the job should stop sooner after cancel() is called.

	// STRETCH: Implement a simple retry helper:
	//   func retry(ctx context.Context, maxAttempts int, fn func(ctx context.Context) error) error
	// It should stop early if the context is cancelled between attempts.
	// Between attempts, sleep an exponentially increasing delay starting at 10ms.
	// Make sure the sleep itself is cancellable (use select + time.After, not time.Sleep).
}

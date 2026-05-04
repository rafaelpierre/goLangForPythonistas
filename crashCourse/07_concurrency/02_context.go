// Topic: context.Context -- Cancellation and Deadlines
//
// Python has no direct equivalent. asyncio has task cancellation, but it is
// implicit. Go's context.Context is EXPLICIT: every function that can be
// cancelled or timed out receives a ctx as its FIRST parameter.
//
// Key ideas:
//   - context.Background() is the root context (used in main, tests, top-level).
//   - context.WithCancel(parent) returns a child ctx + a cancel() function.
//     Call cancel() when you're done (always defer cancel()).
//   - context.WithTimeout / WithDeadline: automatically cancel after a duration.
//   - ctx.Done() returns a channel that is closed when the context is cancelled.
//   - ctx.Err() returns context.Canceled or context.DeadlineExceeded.
//   - Pass context as the FIRST argument; never store it in a struct.
//
// Real-world use: HTTP handlers receive a ctx tied to the request lifetime.
// If the client disconnects, ctx is cancelled, and all downstream calls
// (DB queries, RPCs) should also abort -- no wasted work.
//
// Run: go run 02_context.go

//go:build ignore

package main

import (
	"context"
	"fmt"
	"time"
)

// Simulates a slow database query. It respects ctx -- if cancelled, it stops early.
func slowQuery(ctx context.Context, query string) (string, error) {
	// Simulate work that takes 200ms
	select {
	case <-time.After(200 * time.Millisecond):
		return fmt.Sprintf("result of %q", query), nil
	case <-ctx.Done():
		// ctx was cancelled or timed out before we finished
		return "", ctx.Err()
	}
}

// Simulates a pipeline stage that passes ctx down the call chain.
func fetchData(ctx context.Context) (string, error) {
	// Every downstream call gets the same ctx
	result, err := slowQuery(ctx, "SELECT * FROM users")
	if err != nil {
		return "", fmt.Errorf("fetchData: %w", err)
	}
	return result, nil
}

// Worker that runs until its context is cancelled.
func backgroundWorker(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %q stopping: %v\n", name, ctx.Err())
			return
		case <-time.After(50 * time.Millisecond):
			fmt.Printf("worker %q: tick\n", name)
		}
	}
}

func main() {
	fmt.Println("=== 1. context.WithTimeout -- query finishes in time ===")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel() // always defer cancel even when timeout is set -- cleans up resources

	result, err := fetchData(ctx)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("success:", result)
	}

	fmt.Println("\n=== 2. context.WithTimeout -- query too slow ===")
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer shortCancel()

	result, err = fetchData(shortCtx)
	if err != nil {
		fmt.Println("timed out:", err)
		// Inspect error type
		if err.Error() == "fetchData: "+context.DeadlineExceeded.Error() {
			fmt.Println("confirmed: DeadlineExceeded")
		}
	} else {
		fmt.Println("success:", result)
	}

	fmt.Println("\n=== 3. context.WithCancel -- manual cancellation ===")
	workerCtx, workerCancel := context.WithCancel(context.Background())

	go backgroundWorker(workerCtx, "processor")

	// Let it run for a bit, then cancel
	time.Sleep(180 * time.Millisecond)
	fmt.Println("cancelling worker...")
	workerCancel()

	// Give the goroutine time to print its stop message
	time.Sleep(20 * time.Millisecond)
	fmt.Println("main done")

	fmt.Println("\n=== 4. ctx.Value -- passing request-scoped data ===")
	// Use a typed key to avoid collisions (string keys collide across packages)
	type ctxKey string
	const requestIDKey ctxKey = "requestID"

	reqCtx := context.WithValue(context.Background(), requestIDKey, "req-42")

	// Downstream function retrieves it
	printRequestID := func(ctx context.Context) {
		if id, ok := ctx.Value(requestIDKey).(string); ok {
			fmt.Println("handling request:", id)
		}
	}
	printRequestID(reqCtx)

	// NOTE: ctx.Value is for request-scoped metadata (request IDs, auth tokens).
	// Do NOT use it to pass optional function parameters -- that is an anti-pattern.

	// TODO 1: Write a function 'retryWithContext(ctx context.Context, fn func() error,
	// maxAttempts int) error' that retries fn up to maxAttempts times, but stops
	// immediately if ctx is cancelled. Use select with ctx.Done() between retries.

	// TODO 2: Simulate an HTTP handler: create a context with a 300ms timeout,
	// then concurrently run two "queries" (each takes a random time 100-400ms).
	// Use errgroup or goroutines + channels to collect results. Cancel all work
	// if any one of them fails or times out.

	// TODO 3: Implement a function 'withRetryDeadline(ctx context.Context,
	// deadline time.Duration, work func(context.Context) error) error'.
	// It creates a child context with the given deadline, calls work, and
	// propagates the error. The key insight: cancelling the parent also
	// cancels the child -- demonstrate this.
}

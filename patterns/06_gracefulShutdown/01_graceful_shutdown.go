// Topic: Graceful Shutdown
//
// A production server must handle OS signals (SIGINT/Ctrl-C, SIGTERM from
// Kubernetes) without dropping in-flight requests or leaving open connections.
// The pattern is always the same:
//   1. Create a root context cancelled when a shutdown signal arrives.
//   2. Pass that context to all long-running goroutines and servers.
//   3. After signalling, wait (with a timeout) for goroutines to finish.
//   4. Exit cleanly.
//
// Python analog: signal.signal(signal.SIGTERM, handler) + asyncio.run() with
// a custom event loop shutdown sequence. Go's approach is more explicit but
// also more composable -- you can cancel the root context from anywhere.
//
// This file simulates an HTTP-like server that accepts "requests" (from a
// goroutine), processes them with some latency, and shuts down cleanly.
//
// Run: go run 01_graceful_shutdown.go
// Press Ctrl-C to trigger the shutdown signal.
// The program also auto-triggers a simulated SIGINT after 2 seconds so you
// can watch the shutdown sequence without pressing anything.
//
//go:build ignore

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ---------------------------------------------------------------------------
// Simulated request handler
// ---------------------------------------------------------------------------

// handleRequest simulates a slow HTTP handler. It respects ctx so it can be
// cancelled mid-flight. In real net/http this is r.Context() being cancelled.
func handleRequest(ctx context.Context, id int) error {
	latency := time.Duration(rand.Intn(800)+200) * time.Millisecond
	fmt.Printf("  [req %d] started (will take %v)\n", id, latency)

	select {
	case <-time.After(latency):
		fmt.Printf("  [req %d] completed\n", id)
		return nil
	case <-ctx.Done():
		fmt.Printf("  [req %d] cancelled after partial work: %v\n", id, ctx.Err())
		return ctx.Err()
	}
}

// ---------------------------------------------------------------------------
// Simulated server: accepts work, tracks in-flight requests
// ---------------------------------------------------------------------------

// Server models a long-running service. In real code, replace this with
// net/http.Server (which has built-in Shutdown support).
type Server struct {
	requestsIn chan int   // channel of incoming request IDs
	shutdown   chan struct{} // closed to signal no new requests accepted
	wg         sync.WaitGroup
	mu         sync.Mutex
	inFlight   int
}

func NewServer() *Server {
	return &Server{
		requestsIn: make(chan int, 16),
		shutdown:   make(chan struct{}),
	}
}

// Serve drains requestsIn, handling each request in a goroutine.
// ctx is the server's lifetime context -- cancelled on shutdown signal.
func (s *Server) Serve(ctx context.Context) {
	fmt.Println("server: accepting requests")
	for {
		select {
		case id, ok := <-s.requestsIn:
			if !ok {
				return // channel closed
			}
			s.wg.Add(1)
			s.mu.Lock()
			s.inFlight++
			s.mu.Unlock()

			go func(reqID int) {
				defer s.wg.Done()
				defer func() {
					s.mu.Lock()
					s.inFlight--
					s.mu.Unlock()
				}()
				// Each request gets the server's context.
				// When ctx is cancelled, in-flight requests are notified.
				handleRequest(ctx, reqID)
			}(id)

		case <-s.shutdown:
			fmt.Println("server: stop accepting new requests")
			return
		}
	}
}

// Shutdown stops accepting new requests and waits for in-flight ones to finish,
// up to the given deadline. Returns an error if the drain timed out.
func (s *Server) Shutdown(ctx context.Context) error {
	close(s.shutdown) // stop Serve() from accepting more requests

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("server: all in-flight requests finished")
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		remaining := s.inFlight
		s.mu.Unlock()
		return fmt.Errorf("server: shutdown timed out with %d requests still in flight", remaining)
	}
}

// ---------------------------------------------------------------------------
// Simulated load generator
// ---------------------------------------------------------------------------

// sendRequests generates request IDs and submits them to the server.
// Stops when stopCh is closed (caller's signal to stop sending).
func sendRequests(srv *Server, stopCh <-chan struct{}) {
	id := 1
	for {
		select {
		case <-stopCh:
			fmt.Println("load-gen: stopping")
			return
		case <-time.After(time.Duration(rand.Intn(200)+100) * time.Millisecond):
			select {
			case srv.requestsIn <- id:
				fmt.Printf("load-gen: submitted request %d\n", id)
				id++
			default:
				fmt.Printf("load-gen: queue full, dropped request %d\n", id)
				id++
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Main: signal handling + graceful shutdown sequence
// ---------------------------------------------------------------------------

func main() {
	rand.Seed(time.Now().UnixNano())

	// rootCtx is cancelled when we want to stop in-flight request processing.
	// This is separate from the shutdown timeout context below.
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	srv := NewServer()

	// Capture OS signals. signal.NotifyContext is the idiomatic Go 1.16+ approach:
	// it returns a context that is cancelled when the listed signals arrive.
	sigCtx, sigStop := signal.NotifyContext(rootCtx, syscall.SIGINT, syscall.SIGTERM)
	defer sigStop()

	// Start the server in a goroutine
	var serverWg sync.WaitGroup
	serverWg.Add(1)
	go func() {
		defer serverWg.Done()
		srv.Serve(sigCtx) // Serve stops when sigCtx is cancelled or shutdown fires
	}()

	// Start the load generator
	stopLoad := make(chan struct{})
	go sendRequests(srv, stopLoad)

	// Simulate auto-shutdown after 2 seconds so this file is runnable
	// unattended. In production this goroutine wouldn't exist -- the signal
	// would come from Ctrl-C or Kubernetes.
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\nmain: simulating SIGINT (auto-triggered for demo)")
		// Send to the process's own signal channel to trigger the same
		// path as a real OS signal.
		self, err := os.FindProcess(os.Getpid())
		if err == nil {
			self.Signal(syscall.SIGINT)
		}
	}()

	// ---- Block until signal arrives ----
	<-sigCtx.Done()
	fmt.Printf("\nmain: got signal (%v) -- starting graceful shutdown\n", sigCtx.Err())

	// 1. Stop accepting new requests immediately.
	close(stopLoad) // tell load generator to stop

	// 2. Cancel root context: in-flight requests see their context cancelled.
	//    They may or may not finish depending on how they check ctx.Done().
	rootCancel()

	// 3. Wait for in-flight requests to drain, with a hard deadline.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()

	fmt.Println("main: waiting for in-flight requests to drain (3s max)...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("main: shutdown error: %v\n", err)
		os.Exit(1)
	}

	// 4. Wait for the Serve goroutine itself to exit.
	serverWg.Wait()

	fmt.Println("main: shutdown complete. Goodbye.")

	// TODO 1: The current Server does not distinguish "request cancelled by
	// shutdown" from "request that errored for business reasons". Add a
	// Metrics struct with InFlight, Completed, Cancelled, and Errored int64
	// counters. Increment them atomically (use sync/atomic or a mutex).
	// Print the metrics summary during shutdown, after Shutdown() returns.

	// TODO 2: Add a second background goroutine: a "health ticker" that prints
	// "server healthy: N in-flight" every 500ms. It should stop cleanly when
	// the root context is cancelled (use a select loop, not time.Sleep).
	// Make sure it exits before os.Exit is called by using a WaitGroup.

	// TODO 3: Replace the simulated auto-SIGINT with a real interactive shutdown:
	// remove the auto-trigger goroutine, then run the file with:
	//   go run 01_graceful_shutdown.go
	// and press Ctrl-C. Verify the shutdown sequence runs and prints the
	// "shutdown complete" message before the process exits.

	// STRETCH: In real net/http servers, use srv.Shutdown(ctx) on a
	// *http.Server directly. Replace this entire simulated server with a
	// real net/http.Server that serves GET /slow (sleeps 500ms then returns
	// "ok"). Wire up the same graceful shutdown sequence using http.Server.Shutdown.
	// Test by curling http://localhost:8080/slow while pressing Ctrl-C.
}

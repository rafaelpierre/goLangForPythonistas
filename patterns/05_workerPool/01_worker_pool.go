// Topic: Worker Pool with Backpressure and Graceful Shutdown
//
// You've seen a basic worker pool in crashCourse/07_concurrency. This exercise
// builds a production-grade version with:
//   - Backpressure: the job queue has a bounded buffer; senders block when full
//   - Result collection: workers emit typed results through a separate channel
//   - Graceful shutdown: context cancellation stops workers cleanly, no goroutine leaks
//   - Error propagation: each result carries either a value or an error
//
// Python analog:
//   concurrent.futures.ThreadPoolExecutor with submit() returning Futures.
//   Go's equivalent is a set of goroutines reading from a shared jobs channel.
//   Key difference: in Go you design the backpressure explicitly (buffered channels);
//   in Python the executor queues unboundedly by default.
//
// Real-world use: processing a Kafka topic with N parallel workers, batch image
// resizing, parallel API enrichment (rate-limited calls to an external service).
//
// Run: go run 01_worker_pool.go

//go:build ignore

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

// Job represents a unit of work. In real code this might be a message from
// a queue, a row from a database cursor, or an item from an API response.
type Job struct {
	ID      int
	Payload string
}

// Result pairs a completed job with either a value or an error.
type Result struct {
	JobID  int
	Output string
	Err    error
}

// ---------------------------------------------------------------------------
// Simulated "external call" that the workers perform
//
// This mimics calling a rate-limited external API. It has variable latency
// and a configurable failure rate to make the demo interesting.
// ---------------------------------------------------------------------------

func processJob(ctx context.Context, job Job) (string, error) {
	// Respect context cancellation before starting
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("job %d: context done before start: %w", job.ID, err)
	}

	latency := time.Duration(rand.Intn(80)+20) * time.Millisecond

	select {
	case <-time.After(latency):
		// Simulate a 15% failure rate
		if rand.Intn(100) < 15 {
			return "", fmt.Errorf("job %d: external API returned 500", job.ID)
		}
		return fmt.Sprintf("processed(%s)", job.Payload), nil
	case <-ctx.Done():
		return "", fmt.Errorf("job %d: cancelled during processing: %w", job.ID, ctx.Err())
	}
}

// ---------------------------------------------------------------------------
// Worker Pool
// ---------------------------------------------------------------------------

// Pool manages a fixed number of goroutines that drain a jobs channel.
type Pool struct {
	workers  int
	jobsCh   chan Job
	resultCh chan Result
	wg       sync.WaitGroup
}

// NewPool creates a Pool with the given number of workers and job-queue depth.
// jobQueueDepth is the backpressure knob: a small value means producers block
// sooner when workers can't keep up; a large value allows bursts but uses memory.
func NewPool(workers, jobQueueDepth int) *Pool {
	return &Pool{
		workers:  workers,
		jobsCh:   make(chan Job, jobQueueDepth),
		resultCh: make(chan Result, workers*2), // result buffer proportional to workers
	}
}

// Start launches the worker goroutines. The pool stops when ctx is cancelled
// OR when all submitted jobs are drained and the pool is closed.
func (p *Pool) Start(ctx context.Context) {
	for id := 1; id <= p.workers; id++ {
		p.wg.Add(1)
		go p.runWorker(ctx, id)
	}

	// Close resultCh once all workers exit, so the result consumer's range-loop ends.
	go func() {
		p.wg.Wait()
		close(p.resultCh)
	}()
}

func (p *Pool) runWorker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled: drain any remaining jobs with a cancellation result
			// rather than leaving them unacknowledged.
			for job := range p.jobsCh {
				p.resultCh <- Result{
					JobID: job.ID,
					Err:   fmt.Errorf("worker %d: pool shutting down: %w", id, ctx.Err()),
				}
			}
			return
		case job, ok := <-p.jobsCh:
			if !ok {
				// jobsCh was closed: no more work, exit cleanly.
				return
			}
			output, err := processJob(ctx, job)
			p.resultCh <- Result{JobID: job.ID, Output: output, Err: err}
		}
	}
}

// Submit sends a job to the pool. Blocks if the job queue is full (backpressure).
// Returns an error if the context is cancelled while waiting.
func (p *Pool) Submit(ctx context.Context, job Job) error {
	select {
	case p.jobsCh <- job:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("pool.Submit: job %d dropped: %w", job.ID, ctx.Err())
	}
}

// Close signals that no more jobs will be submitted.
// Workers finish their current job, drain remaining queued jobs, then exit.
// Always call Close after all Submit calls are done.
func (p *Pool) Close() {
	close(p.jobsCh)
}

// Results returns the channel of results for the consumer to read.
// The channel is closed automatically once all workers have exited.
func (p *Pool) Results() <-chan Result {
	return p.resultCh
}

// ---------------------------------------------------------------------------
// Main: demonstrate the pool with two scenarios
// ---------------------------------------------------------------------------

func runPool(ctx context.Context, totalJobs, workers, queueDepth int, label string) {
	fmt.Printf("\n=== %s: %d jobs, %d workers, queue depth %d ===\n",
		label, totalJobs, workers, queueDepth)

	pool := NewPool(workers, queueDepth)
	pool.Start(ctx)

	// Producer goroutine: submit jobs; this decouples production speed from
	// processing speed and lets the queue absorb bursts.
	go func() {
		for i := 1; i <= totalJobs; i++ {
			job := Job{ID: i, Payload: fmt.Sprintf("item-%03d", i)}
			if err := pool.Submit(ctx, job); err != nil {
				fmt.Printf("submit failed: %v\n", err)
				break
			}
		}
		pool.Close() // signal: no more jobs
	}()

	// Consumer: collect results. This runs in the main goroutine.
	var (
		succeeded, failed int
		cancelledByCtx    int
	)
	for result := range pool.Results() {
		if result.Err != nil {
			if errors.Is(result.Err, context.Canceled) || errors.Is(result.Err, context.DeadlineExceeded) {
				cancelledByCtx++
			} else {
				failed++
				fmt.Printf("  job %d failed: %v\n", result.JobID, result.Err)
			}
		} else {
			succeeded++
		}
	}

	fmt.Printf("results: %d succeeded, %d failed, %d cancelled by context\n",
		succeeded, failed, cancelledByCtx)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Scenario 1: all jobs complete -- long timeout
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	runPool(ctx1, 20, 4, 8, "normal run")

	// Scenario 2: context cancelled mid-way -- some jobs will be dropped
	ctx2, cancel2 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel2()
	runPool(ctx2, 30, 3, 5, "deadline exceeded mid-run")

	fmt.Println("\ndone.")

	// TODO 1: Add a retry mechanism inside processJob: if the external API
	// fails (non-context error), retry up to 2 times with a 20ms delay
	// between attempts. The delay must be cancellable (use select + time.After,
	// not time.Sleep). After all retries are exhausted, return the last error.

	// TODO 2: Add a Stats struct with fields Submitted, Succeeded, Failed,
	// Cancelled int64. Use sync/atomic to increment them safely from multiple
	// goroutines (or use a mutex -- pick one, explain the trade-off in a comment).
	// Print final stats after runPool returns.

	// TODO 3: Add a WithRateLimit(rps int) option to Pool that uses a time.Ticker
	// to gate how many jobs per second workers are STARTED (not submitted).
	// Apply the rate limit inside runWorker before calling processJob.
	// This models an API rate limiter: you can queue more jobs than the rate
	// allows, but workers won't exceed rps calls per second to the external API.

	// STRETCH: Implement a two-stage pipeline:
	//   Stage 1 (fetcher pool, 3 workers): reads IDs from a source channel, fetches
	//     "raw" data via processJob, emits raw results to a middle channel.
	//   Stage 2 (transformer pool, 2 workers): reads from the middle channel,
	//     transforms the raw string (e.g., strings.ToUpper), emits to the final
	//     results channel.
	// This is the fan-out/fan-in pipeline pattern used in ETL and stream processing.
}

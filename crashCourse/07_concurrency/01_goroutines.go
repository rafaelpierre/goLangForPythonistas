// Topic: Goroutines and Channels
//
// Python concurrency options:
//   - threading: real threads, but GIL limits CPU parallelism
//   - asyncio: cooperative, single-threaded, needs async/await everywhere
//   - multiprocessing: true parallelism, heavy processes
//
// Go's model:
//   - goroutines: lightweight (2 KB stack), multiplexed over OS threads by the
//     Go scheduler. Starting one is as cheap as 'go fn()'.
//   - channels: typed conduits for communication between goroutines.
//     "Do not communicate by sharing memory; share memory by communicating."
//   - sync.WaitGroup: wait for a group of goroutines to finish (like asyncio.gather).
//   - select: wait on multiple channels simultaneously (like asyncio.wait/select).
//
// Run: go run 01_goroutines.go

//go:build ignore

package main

import (
	"fmt"
	"sync"
	"time"
)

// --- 1. Basic goroutine + WaitGroup ---

func greet(name string, wg *sync.WaitGroup) {
	defer wg.Done() // signals "I'm done" when this function returns
	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("Hello from goroutine: %s\n", name)
}

// --- 2. Channel basics ---

// producer sends values into the channel, then closes it.
func producer(ch chan<- int, n int) { // chan<- means "send-only"
	for i := 0; i < n; i++ {
		ch <- i // send i into channel; blocks if channel buffer is full
	}
	close(ch) // signal to receivers that no more values are coming
}

// --- 3. Fan-out worker pool ---
// Real-world use: process a queue of jobs with N parallel workers.
// In Python you'd use concurrent.futures.ThreadPoolExecutor or asyncio.gather.

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs { // range over a channel: receives until it's closed
		// Simulate processing: square the job number
		result := job * job
		results <- result
		fmt.Printf("worker %d: job=%d result=%d\n", id, job, result)
	}
}

// --- 4. select: multiplex channels ---

func ticker(ch chan<- string, msg string, interval time.Duration, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			fmt.Printf("ticker(%q) stopping\n", msg)
			return
		case <-time.After(interval):
			ch <- msg
		}
	}
}

func main() {
	fmt.Println("=== 1. Basic goroutines with WaitGroup ===")
	var wg sync.WaitGroup
	names := []string{"alice", "bob", "carol", "dave"}
	for _, name := range names {
		wg.Add(1)
		go greet(name, &wg) // 'go' keyword launches a goroutine
	}
	wg.Wait() // block until all Done() calls balance the Add() calls
	fmt.Println("all goroutines finished")

	fmt.Println("\n=== 2. Buffered channel (producer/consumer) ===")
	ch := make(chan int, 5) // buffered: producer can send 5 before blocking
	go producer(ch, 8)
	for v := range ch { // receive until channel is closed
		fmt.Printf("received: %d\n", v)
	}

	fmt.Println("\n=== 3. Worker pool (fan-out) ===")
	jobs := make(chan int, 10)
	results := make(chan int, 10)

	var poolWg sync.WaitGroup
	const numWorkers = 3
	for id := 1; id <= numWorkers; id++ {
		poolWg.Add(1)
		go worker(id, jobs, results, &poolWg)
	}

	// Send 9 jobs
	for j := 1; j <= 9; j++ {
		jobs <- j
	}
	close(jobs) // no more jobs; workers will drain and exit

	// Close results once all workers are done
	go func() {
		poolWg.Wait()
		close(results)
	}()

	// Collect results
	var total int
	for r := range results {
		total += r
	}
	fmt.Println("sum of squares:", total)

	fmt.Println("\n=== 4. select: first channel wins ===")
	fast := make(chan string, 1)
	slow := make(chan string, 1)

	go func() {
		time.Sleep(20 * time.Millisecond)
		fast <- "fast result"
	}()
	go func() {
		time.Sleep(100 * time.Millisecond)
		slow <- "slow result"
	}()

	// Wait for whichever arrives first (like asyncio.wait with FIRST_COMPLETED)
	select {
	case msg := <-fast:
		fmt.Println("got:", msg)
	case msg := <-slow:
		fmt.Println("got:", msg)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("timeout")
	}

	// TODO 1: Launch 5 goroutines, each computing fib(n) for n in [30..34].
	// Collect results through a channel and print them in order of arrival.
	// Observe that output order may not match input order.

	// TODO 2: Implement a simple rate limiter using a ticker channel:
	// process 10 "requests" but only allow 2 per second. Use time.Tick
	// as a token source and a for loop over the requests.

	// TODO 3: Add a timeout to findUser from the errors exercise by wrapping
	// it in a goroutine that sends its result on a channel. Use select with
	// time.After(50 * time.Millisecond) to return an error if it's too slow.
}

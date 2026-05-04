// Topic: defer, init, and Go Program Structure
//
// Python has context managers (with/as) and __init__ module-level code.
// Go has 'defer' for cleanup and 'init()' for package-level setup.
//
// Key ideas about defer:
//   - 'defer fn()' schedules fn to run when the surrounding FUNCTION returns,
//     regardless of how it returns (normal, error, panic).
//   - Deferred calls stack in LIFO order (last defer = first to run).
//   - Arguments to a deferred function are evaluated IMMEDIATELY at the defer
//     statement, not when the deferred call executes.
//   - Classic uses: closing files/connections, unlocking mutexes, logging.
//   - Python equivalent: 'finally' block or a context manager __exit__.
//
// Key ideas about init():
//   - A package can have one or more init() functions; they run automatically
//     before main(), after all imports are initialized.
//   - Use for registering drivers, validating config, setting up global state.
//   - Python equivalent: module-level code that runs on import.
//
// Run: go run 01_defer_and_init.go

//go:build ignore

package main

import (
	"fmt"
	"os"
)

// --- init() runs before main() ---
var appVersion string

func init() {
	appVersion = "1.0.0"
	fmt.Println("[init] package initialized, version:", appVersion)
}

// --- defer basics ---

func deferOrder() {
	fmt.Println("start of deferOrder")
	defer fmt.Println("defer 1 (registered first, runs last)")
	defer fmt.Println("defer 2")
	defer fmt.Println("defer 3 (registered last, runs first)")
	fmt.Println("end of deferOrder body")
}

// --- defer for resource cleanup (the most common real-world use) ---

func writeToFile(name, content string) error {
	f, err := os.CreateTemp("", name)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close() // guaranteed to close even if we return early due to an error
	// Python equivalent:
	//   with open(name, 'w') as f:
	//       f.write(content)

	_, err = fmt.Fprintln(f, content)
	if err != nil {
		return fmt.Errorf("write to file: %w", err)
	}
	fmt.Println("wrote to:", f.Name())
	return nil
}

// --- defer argument evaluation is immediate ---

func deferArgEval() {
	x := 0
	// The value of x (0) is captured NOW, not when the defer fires.
	defer fmt.Println("deferred x (captured at defer time):", x)
	x = 42
	fmt.Println("x at end of function:", x)
	// Output order: "x at end..." then "deferred x... 0"
}

// --- defer with named return values (advanced) ---
// This lets a deferred function modify the return value -- use sparingly.

func safeDiv(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurred (e.g. integer division by zero),
			// we catch it here and convert it to an error.
			err = fmt.Errorf("recovered panic: %v", r)
		}
	}()
	result = a / b // panics if b == 0
	return
}

// --- Measuring function duration with defer (useful for profiling) ---

func measured(name string) func() {
	// Returns a function that prints elapsed time when called.
	// Usage: defer measured("myOperation")()
	//        Note the extra () -- it calls measured() immediately to capture
	//        the start time, and defers the RETURNED func.
	start := make(chan struct{})
	close(start) // just to force the compiler to keep start in scope
	_ = start
	fmt.Printf("[timer] %s started\n", name)
	return func() {
		fmt.Printf("[timer] %s done\n", name)
	}
}

func slowOperation() {
	defer measured("slowOperation")()
	// do work...
	fmt.Println("  doing slow work...")
}

func main() {
	fmt.Println("main starts, version:", appVersion)

	fmt.Println("\n--- defer order (LIFO) ---")
	deferOrder()

	fmt.Println("\n--- defer for file cleanup ---")
	if err := writeToFile("example", "hello from Go"); err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("\n--- defer argument evaluation ---")
	deferArgEval()

	fmt.Println("\n--- recover from panic via defer ---")
	result, err := safeDiv(10, 2)
	fmt.Printf("10/2 = %d, err = %v\n", result, err)
	result, err = safeDiv(10, 0)
	fmt.Printf("10/0 = %d, err = %v\n", result, err)

	fmt.Println("\n--- timing with defer ---")
	slowOperation()

	// TODO 1: Write a function 'withLock(mu *sync.Mutex, fn func())' that
	// locks mu, defers mu.Unlock(), then calls fn. This is the idiomatic
	// mutex pattern in Go.

	// TODO 2: Write a function 'traceCall(name string) func()' that prints
	// "entering <name>" immediately and returns a func that prints
	// "leaving <name>". Use it with 'defer traceCall("myFunc")()' to trace
	// entry and exit of any function.

	// TODO 3: Demonstrate that defer runs even when a function returns early.
	// Write a function that opens a "connection" (just a struct with an
	// IsOpen bool field), defers its Close(), then returns early half the time
	// based on a parameter. Verify Close is always called.
}

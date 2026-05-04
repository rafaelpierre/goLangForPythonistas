# Go Crash Course for Pythonistas

Work through the topics in this order. Each folder has one or two files.
Read the comment block at the top of each file before writing any code.

## Learning Path

1. **intro/**                -- Hello World; getting the toolchain working
   - `01_hello.go`

2. **types/**                -- Variables, zero values, type system, functions
   - `01_variables.go`      Zero values, short declaration, constants, type conversion
   - `02_functions.go`      Multiple returns, variadic, first-class functions

3. **slicesAndMaps/**        -- Go's workhorse data structures
   - `01_slices.go`         Slice internals, append, copy, range (vs Python lists)
   - `02_maps.go`           Map literals, two-value lookup, frequency counters

4. **structs/**              -- Data modeling and pointers
   - `01_structs.go`        Structs, methods, value vs pointer receivers, embedding
   - `02_pointers.go`       & and *, when to use pointers, nil gotchas

5. **interfaces/**           -- Go's polymorphism (implicit, not declared)
   - `01_interfaces.go`     Interface definition, Stringer, type assertions, type switch

6. **errors/**               -- Explicit error handling (no exceptions)
   - `01_errors.go`         error interface, sentinel errors, custom types, wrapping,
                             errors.Is / errors.As, panic vs error

7. **concurrency/**          -- Goroutines, channels, context
   - `01_goroutines.go`     go keyword, WaitGroup, buffered channels, worker pool, select
   - `02_context.go`        context.WithTimeout, WithCancel, WithValue; cancellation chains

8. **toolchain/**            -- Idiomatic Go mechanics
   - `01_defer_and_init.go` defer (LIFO, arg eval, recover), init(), resource cleanup

9. **testing/**              -- Built-in testing, table-driven tests, benchmarks
   - `01_tests.go`          Implementation to test
   - `01_tests_test.go`     Table tests, subtests, benchmarks (go test -v .)

## Quick Reference

| Task                        | Command                                      |
|-----------------------------|----------------------------------------------|
| Run a single file           | `go run types/01_variables.go`               |
| Run all files in a folder   | `go run ./concurrency/`  (multi-file pkgs)   |
| Run tests                   | `go test ./testing/`                         |
| Run tests verbose           | `go test -v ./testing/`                      |
| Run benchmarks              | `go test -bench=. ./testing/`                |
| Format code                 | `gofmt -w <file>`                            |
| Vet for common mistakes     | `go vet ./...`                               |

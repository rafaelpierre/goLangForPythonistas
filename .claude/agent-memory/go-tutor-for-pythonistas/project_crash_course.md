---
name: Course Structure (Crash Course + Patterns)
description: Topics covered in crashCourse/ and patterns/ -- file layout, what has been taught, and what TODOs remain
type: project
---

## crashCourse/ (created 2026-05-04)

Subfolders use NN_topicName convention. All exercise files use //go:build ignore.

1. 01_intro/01_hello.go               -- Hello World (pre-existing)
2. 02_types/01_variables.go           -- zero values, short declaration, type conversion
3. 02_types/02_functions.go           -- multiple returns, variadic, first-class functions
4. 03_slicesAndMaps/01_slices.go      -- slice internals, append, copy, range
5. 03_slicesAndMaps/02_maps.go        -- map literals, two-value lookup, frequency counters
6. 04_structs/01_structs.go           -- structs, methods, value vs pointer receivers, embedding
7. 04_structs/02_pointers.go          -- & and *, nil, pointer-to-struct pattern
8. 05_interfaces/01_interfaces.go     -- implicit interfaces, Stringer, type switch
9. 06_errors/01_errors.go             -- error interface, sentinels, wrapping, errors.Is/As
10. 07_concurrency/01_goroutines.go   -- goroutines, WaitGroup, channels, worker pool, select
11. 07_concurrency/02_context.go      -- context.WithTimeout/Cancel/Value, cancellation chains
12. 08_toolchain/01_defer_and_init.go -- defer (LIFO, arg eval, recover), init()
13. 09_testing/01_tests.go            -- FizzBuzz, Reverse, Contains, WordFrequency implementations
14. 09_testing/01_tests_test.go       -- table-driven tests, subtests, benchmarks

## patterns/ (created 2026-05-04)

Intermediate material. Same NN_topicName convention, same //go:build ignore files.
Prerequisite: finish crashCourse/ first.

1. 01_functionalOptions/01_functional_options.go -- WithXxx constructor pattern; HTTPClient example
2. 02_errorWrapping/01_error_wrapping.go          -- sentinel hierarchies, NotFoundError, MultiError
3. 03_contextPropagation/01_context_propagation.go -- context through handler->service->repo chain
4. 04_interfaceDI/01_interface_di.go              -- OrderService wired with OrderStore+Notifier interfaces
5. 05_workerPool/01_worker_pool.go                -- bounded pool with backpressure, graceful drain
6. 06_gracefulShutdown/01_graceful_shutdown.go    -- signal.NotifyContext, Shutdown with drain timeout
7. 07_syncPatterns/01_sync_patterns.go            -- sync.Once (config), sync.Pool (buffers), sync.Map (metrics)
8. 08_generics/01_generics.go                     -- Map/Filter/Reduce, Set[T], Result[T], cmp.Ordered

**Why:** User completed crashCourse setup and requested intermediate design/implementation patterns.
**How to apply:** Check which TODOs the user has completed before designing follow-up exercises.
All TODOs were unfinished as of 2026-05-04 (files freshly created).

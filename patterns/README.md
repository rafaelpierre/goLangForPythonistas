# patterns — Intermediate Go: Design & Implementation Patterns

Prerequisites: finish all of `crashCourse/` first.

Each subfolder is a standalone topic. Files run via:

    go run patterns/NN_topic/NN_file.go

All files use `//go:build ignore` so they coexist without a go.mod.

---

## Recommended order

| # | Folder                  | What you'll learn                                              |
|---|-------------------------|----------------------------------------------------------------|
| 1 | `01_functionalOptions`  | The functional-options constructor pattern (`WithXxx` funcs)  |
| 2 | `02_errorWrapping`      | Sentinel hierarchies, multi-error, deep `errors.Is/As` chains |
| 3 | `03_contextPropagation` | Threading `context.Context` through real call chains          |
| 4 | `04_interfaceDI`        | Dependency injection via interfaces; swappable backends       |
| 5 | `05_workerPool`         | Fan-out/fan-in pipeline with backpressure and shutdown        |
| 6 | `06_gracefulShutdown`   | OS signal handling + context cancellation for servers         |
| 7 | `07_syncPatterns`       | `sync.Once`, `sync.Pool`, `sync.Map` — when and why           |
| 8 | `08_generics`           | Type-safe containers and helpers with Go 1.18+ generics       |

---

## When to start this

After you can:
- Write a struct with methods and a pointer receiver without looking it up
- Handle errors with `fmt.Errorf("%w", err)` and `errors.Is`
- Launch goroutines and wait on them with `sync.WaitGroup`
- Satisfy a Go interface and use it to swap implementations in tests

If any of that is fuzzy, revisit the relevant `crashCourse/` module first.

---

## A note on "design patterns"

Classic OOP patterns (Singleton, Abstract Factory, Template Method) map poorly to Go.
Go's answer is usually: define a small interface, pass it as a parameter, done.
These exercises cover patterns that are actually idiomatic in Go production code —
not patterns translated mechanically from Java textbooks.

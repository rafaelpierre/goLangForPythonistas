# golearn

Personal workspace for learning Go, framed for someone coming from Python.

## How to use this repo

Each exercise is a standalone, runnable Go file. From the repo root:

```
go run crashCourse/01_intro/01_hello.go
go run crashCourse/02_types/01_variables.go
go run patterns/01_functionalOptions/01_functional_options.go
# ...etc
```

Every exercise file carries a `//go:build ignore` tag. That tag tells the Go
toolchain "this file is a standalone script, not part of a package build", which
lets multiple `package main` files coexist in the same folder without colliding
on `func main()`. Without it, `go build ./...` and the LSP would complain about
"main redeclared". You can ignore that tag for now — just know that's why it's
there.

## Recommended learning order

Work through the folders in this order. Each one builds on the previous.

### Crash Course (`crashCourse/`)

1. **`crashCourse/01_intro/`** — Smoke test. The classic Hello World, just to
   confirm your `go` toolchain works.

2. **`crashCourse/02_types/`** — Variables, zero values, the two declaration
   styles (`var` vs `:=`), constants, explicit type conversion, and functions
   (multiple return values, variadics, first-class functions). The biggest
   mental shift from Python: static types and zero values instead of `None`.

3. **`crashCourse/03_slicesAndMaps/`** — Slices (Python's `list`, but typed and
   backed by an array) and maps (Python's `dict`, but typed). Key gotchas:
   slice-mutation aliasing through the backing array, the two-value map lookup
   (`val, ok := m[key]`), and why writing to a nil map panics.

4. **`crashCourse/04_structs/`** — Structs, methods, value vs pointer receivers,
   and embedding (composition over inheritance). Then pointers explicitly:
   `&` and `*`, why you'd use a pointer, and how Go avoids the C footguns.

5. **`crashCourse/05_interfaces/`** — Implicit interface satisfaction (Go's take
   on duck typing), `Stringer`, type assertions, and type switches. This is
   how you write testable Go: define an interface, swap a real DB for a fake.

6. **`crashCourse/06_errors/`** — Errors as values, not exceptions. `error`
   interface, sentinel errors, `fmt.Errorf("...: %w", err)` wrapping, and
   `errors.Is` / `errors.As`. Also covers why `panic`/`recover` is NOT for
   normal error handling.

7. **`crashCourse/07_concurrency/`** — Goroutines, channels, `sync.WaitGroup`,
   `select`, worker pools — and then `context.Context` for cancellation and
   timeouts. The mental model is very different from `asyncio`: there's no
   `async`/`await`, every goroutine is preemptible, and you communicate by
   passing values through channels.

8. **`crashCourse/08_toolchain/`** — `defer` (LIFO ordering, immediate argument
   evaluation, `recover`) and package `init()` functions. The Go equivalents
   of `with`/`finally` and module-level setup.

9. **`crashCourse/09_testing/`** — The built-in `testing` package, table-driven
   tests, subtests with `t.Run`, and benchmarks. No pytest needed — the
   toolchain has it all. Run with `go test ./crashCourse/09_testing/`.

### Patterns (`patterns/`)

Start this after you can write error-handling, interfaces, and basic goroutines
without looking them up. These exercises cover design and implementation patterns
that show up constantly in real Go production code.

| # | Folder                      | Pattern                                                |
|---|-----------------------------|--------------------------------------------------------|
| 1 | `01_functionalOptions/`     | Functional options constructor (`WithXxx` functions)   |
| 2 | `02_errorWrapping/`         | Sentinel hierarchies, multi-error, deep chain walking  |
| 3 | `03_contextPropagation/`    | Threading context through real call chains             |
| 4 | `04_interfaceDI/`           | Dependency injection via interfaces, swappable fakes   |
| 5 | `05_workerPool/`            | Fan-out/fan-in pipeline with backpressure and shutdown |
| 6 | `06_gracefulShutdown/`      | OS signal handling + context cancellation for servers  |
| 7 | `07_syncPatterns/`          | `sync.Once`, `sync.Pool`, `sync.Map` — when and why   |
| 8 | `08_generics/`              | Type-safe containers and helpers (Go 1.18+ generics)   |

See `patterns/README.md` for the full description and prerequisites.

## See also

- `crashCourse/README.md` — quick-reference command table for the crash course.
- `patterns/README.md` — pattern list, prerequisites, and recommended order for
  the intermediate material.

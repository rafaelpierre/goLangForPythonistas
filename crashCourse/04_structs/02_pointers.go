// Topic: Pointers
//
// Python hides pointers entirely -- everything is an object reference under the hood.
// Go exposes them explicitly, which gives you precise control over copying vs sharing.
//
// Key ideas:
//   - &x  gives you the memory address of x  (a *T -- "pointer to T")
//   - *p  dereferences p to read/write the value it points to
//   - new(T) allocates a zero-value T and returns *T (rare in practice)
//   - Go has NO pointer arithmetic (unlike C). Pointers are safe.
//   - The compiler often auto-dereferences for method calls on structs:
//     if p is *Point, p.X works -- you don't need (*p).X
//   - nil is the zero value of any pointer type. Dereferencing nil panics.
//   - When to use a pointer:
//       1. You need to mutate the value from a function.
//       2. The struct is large and copying would be wasteful.
//       3. You need to represent "absent" (pointer can be nil; value cannot).
//
// Run: go run 02_pointers.go

//go:build ignore

package main

import "fmt"

// Without a pointer: the function gets a COPY -- changes don't affect the original.
// Python analogy: integers are immutable in Python; passing one to a function
// cannot change the caller's variable. In Go, ALL types work this way by default.
func doubleByValue(n int) {
	n *= 2
	fmt.Println("inside doubleByValue, n =", n)
}

// With a pointer: the function receives the address and mutates through it.
func doubleByPointer(n *int) {
	*n *= 2 // dereference and assign
}

// Useful pattern: optional config value using a pointer to distinguish
// "not set" (nil) from "set to zero".
type Config struct {
	Timeout *int // nil means "use the default"; 0 is a valid explicit value
}

func effectiveTimeout(cfg Config) int {
	if cfg.Timeout == nil {
		return 30 // default
	}
	return *cfg.Timeout
}

// Pointer to struct -- the idiomatic way to share a struct across functions.
type Server struct {
	host string
	port int
}

func (s *Server) setPort(p int) { // pointer receiver -- mutates the server
	s.port = p
}

func (s Server) addr() string { // value receiver -- read-only
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func main() {
	// Value semantics
	x := 10
	doubleByValue(x)
	fmt.Println("after doubleByValue, x =", x) // still 10

	// Pointer semantics
	doubleByPointer(&x)
	fmt.Println("after doubleByPointer, x =", x) // now 20

	// & and * basics
	y := 42
	ptr := &y
	fmt.Printf("y=%d, ptr=%p, *ptr=%d\n", y, ptr, *ptr)
	*ptr = 100
	fmt.Println("after *ptr=100, y =", y)

	// Pointer to struct (auto-dereferencing for field access)
	s := &Server{host: "localhost", port: 8080}
	fmt.Println("addr:", s.addr())
	s.setPort(9090) // Go automatically passes s as *Server to pointer receiver
	fmt.Println("new addr:", s.addr())

	// Optional value pattern
	zero := 0
	cfgExplicit := Config{Timeout: &zero}
	cfgDefault := Config{} // Timeout is nil

	fmt.Println("explicit timeout:", effectiveTimeout(cfgExplicit)) // 0
	fmt.Println("default timeout:", effectiveTimeout(cfgDefault))   // 30

	// GOTCHA: returning a pointer to a local variable is SAFE in Go.
	// The compiler will heap-allocate it. (In C this would be undefined behavior.)
	newInt := func(n int) *int { return &n }
	p := newInt(99)
	fmt.Println("heap-allocated int:", *p)

	// GOTCHA: nil pointer dereference panics.
	// var nilPtr *int
	// fmt.Println(*nilPtr)  -- PANIC: runtime error: invalid memory address

	// TODO 1: Write a function 'increment(n *int)' that adds 1 to the value
	// at the pointer. Call it three times on the same variable and print the result.

	// TODO 2: Write a function 'swap(a, b *int)' that swaps two integers
	// via pointers. Verify it works (unlike doubleByValue above).

	// TODO 3: Create a struct 'Node' with fields Value int and Next *Node.
	// Build a small linked list of 3 nodes manually (no append, just &Node{...}).
	// Walk it with a for loop and print each value.
}

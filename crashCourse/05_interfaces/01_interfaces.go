// Topic: Interfaces
//
// Python uses duck typing: "if it has a .quack() method, it's a duck."
// Go uses IMPLICIT interfaces: a type satisfies an interface automatically
// if it implements all the methods -- no 'implements' keyword needed.
//
// Key ideas:
//   - An interface is a set of method signatures. Any type that implements
//     ALL those methods satisfies the interface -- no declaration required.
//   - The empty interface (interface{} or 'any') accepts any type.
//     Python equivalent: just 'object' or no type annotation at all.
//   - Type assertions: val.(ConcreteType) -- like Python's isinstance/cast.
//   - Type switches: idiomatic way to handle multiple concrete types.
//   - Interfaces enable testability: swap a real DB for an in-memory fake
//     without changing the code that uses it.
//
// Real-world use: define a storage interface; tests use an in-memory impl;
// production uses a Postgres impl. Same interface, different behavior.
//
// Run: go run 01_interfaces.go

//go:build ignore

package main

import (
	"fmt"
	"math"
	"strings"
)

// --- Define an interface ---
// Python equivalent: class Shape(Protocol): def area(self) -> float: ...
type Shape interface {
	Area() float64
	Perimeter() float64
}

// --- Types that satisfy Shape (no explicit declaration) ---

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

// A function that works with ANY Shape -- it doesn't know or care about
// the concrete type. This is polymorphism in Go.
func printShape(s Shape) {
	fmt.Printf("%T  area=%.2f  perimeter=%.2f\n", s, s.Area(), s.Perimeter())
}

// --- The Stringer interface (from the fmt package) ---
// If your type implements String() string, fmt.Println will use it automatically.
// Python equivalent: __str__ / __repr__

type Temperature struct {
	Celsius float64
}

func (t Temperature) String() string {
	return fmt.Sprintf("%.1f°C (%.1f°F)", t.Celsius, t.Celsius*9/5+32)
}

// --- Multiple interfaces ---
// A type can satisfy many interfaces at once.

type Writer interface {
	Write(data string)
}

type Closer interface {
	Close()
}

type WriteCloser interface { // interface embedding
	Writer
	Closer
}

type LogWriter struct {
	log []string
}

func (lw *LogWriter) Write(data string) {
	lw.log = append(lw.log, data)
}

func (lw *LogWriter) Close() {
	fmt.Println("LogWriter closed with", len(lw.log), "entries:", strings.Join(lw.log, ", "))
}

// --- Type assertion and type switch ---
func describe(i interface{}) { // 'interface{}' == 'any' -- accepts any type
	switch v := i.(type) {
	case int:
		fmt.Printf("int: %d\n", v)
	case string:
		fmt.Printf("string: %q (len=%d)\n", v, len(v))
	case Shape:
		fmt.Printf("Shape with area=%.2f\n", v.Area())
	default:
		fmt.Printf("unknown type: %T\n", v)
	}
}

func main() {
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 3, Height: 4},
	}

	for _, s := range shapes {
		printShape(s)
	}

	// Stringer -- fmt picks it up automatically
	t := Temperature{Celsius: 100}
	fmt.Println(t)

	// Interface as variable -- holds any Shape
	var s Shape = Circle{Radius: 1}
	fmt.Println("area:", s.Area())

	// Type assertion: extract the concrete type
	c, ok := s.(Circle)
	if ok {
		fmt.Println("It's a Circle with radius:", c.Radius)
	}

	// Type switch
	describe(42)
	describe("hello")
	describe(Circle{Radius: 3})
	describe(3.14)

	// WriteCloser
	var wc WriteCloser = &LogWriter{}
	wc.Write("first entry")
	wc.Write("second entry")
	wc.Close()

	// TODO 1: Add a Triangle type with base and height fields.
	// Implement Area() (0.5*base*height) and Perimeter() -- you'll need
	// all three side lengths, so add those as fields too.
	// Add a Triangle to the shapes slice and re-run.

	// TODO 2: Define a 'Discounter' interface with a single method:
	//   Discount(price float64) float64
	// Implement it for two types: PercentOff (e.g. 10% off) and FlatOff
	// (e.g. $5 off). Write a function applyDiscount(d Discounter, price float64)
	// that prints the original and discounted prices.

	// TODO 3: Define a simple 'Storage' interface:
	//   Get(key string) (string, bool)
	//   Set(key, value string)
	// Implement it with an InMemoryStore (backed by a map). Write a function
	// that accepts Storage and uses Get/Set. This is the pattern you'd use
	// to swap in a real database later.
}

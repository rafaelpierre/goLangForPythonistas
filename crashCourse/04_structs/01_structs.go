// Topic: Structs and Methods
//
// Python uses classes with __init__, self, and inheritance.
// Go uses structs (data) + methods (behavior attached to a type).
// There are NO classes, NO inheritance -- only composition.
//
// Key ideas:
//   - A struct is a named collection of fields (like a Python dataclass or namedtuple).
//   - Methods are functions with a RECEIVER: func (r ReceiverType) MethodName() ...
//   - Value receiver: operates on a copy. Pointer receiver: mutates the original.
//     Rule of thumb: use pointer receivers when you need to mutate state OR the
//     struct is large enough that copying is expensive.
//   - Embedding (not inheritance): one struct can embed another to "inherit" its
//     fields and methods -- but it is composition, not IS-A.
//   - Exported names start with an uppercase letter (Python: no real equivalent,
//     just convention with underscores for "private").
//
// Run: go run 01_structs.go

//go:build ignore

package main

import (
	"fmt"
	"math"
)

// --- Basic struct ---
// Python equivalent:
//   @dataclass
//   class Point:
//       x: float
//       y: float

type Point struct {
	X, Y float64 // Exported fields (uppercase) -- visible outside the package
}

// Method with a VALUE receiver -- gets a copy; cannot mutate the original.
// Naming convention: single-letter or short abbreviation of the type name.
func (p Point) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", p.X, p.Y)
}

func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// --- Struct with behavior that mutates state -> pointer receiver ---
type Counter struct {
	count int // unexported (lowercase) -- only visible within this package
	step  int
}

// Constructor function -- Go has no __init__. Convention: func NewXxx(...) *Xxx
func NewCounter(step int) *Counter {
	return &Counter{step: step} // count defaults to 0 (zero value)
}

// Pointer receiver: modifies the original Counter, not a copy.
// Python equivalent: def increment(self): self.count += self.step
func (c *Counter) Increment() {
	c.count += c.step
}

func (c *Counter) Reset() {
	c.count = 0
}

func (c Counter) Value() int { // value receiver is fine here -- just reading
	return c.count
}

// --- Embedding (composition, not inheritance) ---
type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " makes a sound"
}

type Dog struct {
	Animal        // Embedded -- Dog "inherits" Name and Speak()
	Breed  string
}

// Dog can override Speak by defining its own method.
func (d Dog) Speak() string {
	return d.Name + " says: Woof!"
}

func main() {
	// Struct literal (positional -- fragile, avoid for large structs)
	p1 := Point{1.0, 2.0}
	// Struct literal (named fields -- preferred)
	p2 := Point{X: 4.0, Y: 6.0}

	fmt.Println("p1:", p1)
	fmt.Println("p2:", p2.String())
	fmt.Printf("distance: %.4f\n", p1.Distance(p2))

	// Constructor pattern
	c := NewCounter(5)
	c.Increment()
	c.Increment()
	c.Increment()
	fmt.Println("counter value:", c.Value()) // 15
	c.Reset()
	fmt.Println("after reset:", c.Value()) // 0

	// Embedding
	d := Dog{
		Animal: Animal{Name: "Rex"},
		Breed:  "Labrador",
	}
	fmt.Println(d.Speak())       // Dog's own Speak()
	fmt.Println(d.Name)          // promoted from Animal -- no d.Animal.Name needed
	fmt.Println(d.Animal.Speak()) // explicitly calling Animal's Speak()

	// Struct comparison (if all fields are comparable)
	a := Point{1, 2}
	b := Point{1, 2}
	fmt.Println("a == b:", a == b) // true

	// TODO 1: Define a Rectangle struct with Width and Height float64 fields.
	// Add methods:
	//   Area() float64
	//   Perimeter() float64
	//   Scale(factor float64)   -- this one must mutate, so use pointer receiver
	// Create one, scale it, and print area and perimeter.

	// TODO 2: Define a Stack[T] struct (use any for T if generics feel new)
	// with methods Push(val), Pop() (val, bool), Len() int.
	// Hint: embed a []T slice as a field.

	// TODO 3: Create a Circle struct that embeds Point (its center) and has
	// a Radius field. Add an Area() method. Demonstrate that you can access
	// the center coordinates directly via c.X and c.Y (promotion).
}

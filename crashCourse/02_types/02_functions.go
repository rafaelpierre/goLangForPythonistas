// Topic: Functions
//
// Python functions return None by default and can return anything dynamically.
// Go functions must declare their parameter types AND return types up front.
//
// Key ideas:
//   - Multiple return values are idiomatic (especially for (result, error) pairs).
//   - Named return values exist but use sparingly (mostly for documentation).
//   - Functions are first-class values -- you can pass them around.
//   - Variadic functions: func sum(nums ...int) mirrors Python's *args.
//
// Run: go run 02_functions.go

//go:build ignore

package main

import (
	"fmt"
	"strings"
)

// Basic function: explicit parameter and return types.
// Python equivalent: def add(a: int, b: int) -> int: return a + b
func add(a, b int) int { // 'a, b int' is shorthand when both are the same type
	return a + b
}

// Multiple return values -- the Go way to handle results + errors.
// Python equivalent: return value, error  (but Go enforces you handle it)
func divide(a, b float64) (float64, error) {
	if b == 0 {
		// We return the zero value for float64 (0.0) alongside the error.
		// More on errors in the errors/ topic.
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil // nil means "no error"
}

// Variadic function: accepts zero or more ints.
// Python equivalent: def total(*nums): return sum(nums)
func total(nums ...int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

// Functions as values: accept a function as a parameter.
// Python equivalent: def apply(fn, items): return [fn(x) for x in items]
func applyToAll(words []string, fn func(string) string) []string {
	result := make([]string, len(words))
	for i, w := range words {
		result[i] = fn(w)
	}
	return result
}

// 01. clamp function
func clamp(value, min, max int) (int, error) {
	if min >= max {
		return 0, fmt.Errorf("min must be less than max")
	}

	result := 0

	if value <= min {
		result = min
	} else if value >= max {
		result = max
	} else {
		result = value
	}

	return result, nil
}

// 2. string reverse

func reverse(text string) string {
	r := []rune(text)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	return string(r)
}

// 3. minMax

func minMax(nums ...int) (int, int) {

	if len(nums) == 0 {
		return 0, 0
	}

	min, max := nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}

	return min, max
}

func main() {
	// Basic call
	fmt.Println("3 + 4 =", add(3, 4))

	// Multiple return values -- you MUST handle (or explicitly discard) each one.
	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("10 / 3 = %.4f\n", result)
	}

	// Deliberately trigger the error case
	_, err = divide(5, 0)
	if err != nil {
		fmt.Println("Got expected error:", err)
	}

	// Variadic call
	fmt.Println("total(1,2,3,4,5) =", total(1, 2, 3, 4, 5))

	// Spread a slice into a variadic function (like Python's *args unpacking)
	nums := []int{10, 20, 30}
	fmt.Println("total from slice =", total(nums...))

	// Anonymous function (closure) assigned to a variable
	double := func(n int) int { return n * 2 }
	fmt.Println("double(7) =", double(7))

	// Pass a function as an argument
	words := []string{"hello", "world", "go"}
	upper := applyToAll(words, strings.ToUpper)
	fmt.Println("uppercased:", upper)

	// TODO 1: Write a function 'clamp(value, min, max int) int' that returns
	// value clamped to [min, max]. Call it and print the result.

	clamped, error := clamp(20, 10, 15)
	fmt.Printf("Clamped: %d, error: %s\n", clamped, error)

	// TODO 2: Write an anonymous function that takes a string and returns
	// it reversed. Assign it to a variable and test it with "golang".

	reversed := reverse("golang")
	fmt.Printf("reversed: %s\n", reversed)

	// TODO 3: Write a function 'minMax(nums ...int) (int, int)' that returns
	// the minimum and maximum of its arguments. Handle the empty case by
	// returning 0, 0.

	minVal, maxVal := minMax(1, 2, 9, 30, -2)
	fmt.Printf("min: %d, max: %d\n", minVal, maxVal)
}

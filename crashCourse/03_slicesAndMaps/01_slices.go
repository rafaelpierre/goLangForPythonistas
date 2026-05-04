// Topic: Slices
//
// Python list vs Go slice:
//   Python: nums = [1, 2, 3]  -- dynamic, heterogeneous, reference semantics
//   Go:     nums := []int{1, 2, 3}  -- dynamic length, but HOMOGENEOUS (typed)
//
// Key ideas:
//   - A slice is a VIEW into an underlying array (header = pointer + len + cap).
//   - Arrays ([3]int) have fixed length and are VALUE types -- rarely used directly.
//   - make([]T, len, cap) is the idiomatic way to pre-allocate.
//   - append() returns a new slice header; you MUST use the return value.
//   - Slicing a slice shares the backing array -- mutations affect the original!
//   - 'range' over a slice gives (index, value) -- like Python's enumerate().
//
// Run: go run 01_slices.go

//go:build ignore

package main

import "fmt"

func main() {
	// Slice literal (most common)
	fruits := []string{"apple", "banana", "cherry"}
	fmt.Println("fruits:", fruits)
	fmt.Printf("len=%d, cap=%d\n", len(fruits), cap(fruits))

	// append -- always assign back! (Python: list.append mutates in-place)
	fruits = append(fruits, "date")
	fmt.Println("after append:", fruits)

	// Slicing: [low:high] -- high is exclusive, just like Python
	fmt.Println("fruits[1:3]:", fruits[1:3])

	// GOTCHA: the sub-slice shares backing memory.
	sub := fruits[1:3]
	sub[0] = "BLUEBERRY" // this also changes fruits[1]!
	fmt.Println("after mutating sub-slice:")
	fmt.Println("  sub:", sub)
	fmt.Println("  fruits:", fruits) // notice fruits[1] changed

	// To avoid this, copy explicitly:
	safeSub := make([]string, 2)
	copy(safeSub, fruits[1:3])
	safeSub[0] = "grape"
	fmt.Println("safeSub:", safeSub)
	fmt.Println("fruits unchanged:", fruits)

	// make() -- pre-allocate when you know the size (avoids repeated realloc)
	scores := make([]int, 5) // len=5, all zeros
	fmt.Println("zero scores:", scores)

	// range -- like Python's enumerate()
	// Python: for i, v in enumerate(fruits): ...
	for i, v := range fruits {
		fmt.Printf("  [%d] %s\n", i, v)
	}

	// If you only need the index: for i := range fruits { ... }
	// If you only need the value: for _, v := range fruits { ... }
	// The blank identifier '_' discards the value (like _ in Python unpacking).

	// 2D slice (slice of slices -- like a list of lists in Python)
	matrix := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	fmt.Println("matrix[1][2]:", matrix[1][2]) // 6

	// TODO 1: Create a slice of 5 ints using make(). Fill it with the squares
	// of indices 0..4 using a regular for loop (not range). Print it.

	// TODO 2: Write a function filterEven(nums []int) []int that returns a
	// new slice containing only the even numbers. Call it and print the result.
	// Hint: start with var result []int (nil slice), then append to it.

	// TODO 3: Demonstrate the copy-vs-share gotcha intentionally:
	// create a slice, take a sub-slice, modify the sub-slice, and print
	// both to prove they share memory. Then fix it with copy().
	_ = matrix
	_ = scores
}

// Topic: Testing in Go
//
// Python uses pytest (or unittest). Go has a built-in 'testing' package.
// No test framework needed -- the toolchain handles everything.
//
// Key ideas:
//   - Test files end in _test.go and are only compiled during 'go test'.
//   - Test functions have signature: func TestXxx(t *testing.T)
//   - t.Error / t.Errorf: mark failure and continue.
//   - t.Fatal / t.Fatalf: mark failure and stop the current test.
//   - Table-driven tests are the Go idiom for parameterized tests.
//     Python equivalent: @pytest.mark.parametrize
//   - Subtests via t.Run("name", func(t *testing.T) {...}) -- nest them freely.
//   - Benchmarks: func BenchmarkXxx(b *testing.B) { for i := 0; i < b.N; i++ {...} }
//
// This file contains the IMPLEMENTATION. The tests live in 01_tests_test.go.
// Run: go test ./testing/  (from crashCourse/) or go test . (from this directory)
//
// NOTE: This file is 'package main' so you can also run it standalone,
// but in real projects put library code in its own package (not 'main').

package main

import "fmt"

// --- Functions to test ---

// FizzBuzz returns "Fizz" for multiples of 3, "Buzz" for 5,
// "FizzBuzz" for both, and the number as a string otherwise.
// Python engineers know this one. Here it's our test subject.
func FizzBuzz(n int) string {
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	default:
		return fmt.Sprintf("%d", n)
	}
}

// Reverse returns the string with its bytes reversed.
// NOTE: this is byte-reversal, not rune-reversal. It breaks on multi-byte UTF-8.
// That is intentional -- one of the TODOs is to fix it.
func Reverse(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

// Contains reports whether needle is in haystack (case-sensitive).
func Contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// WordFrequency counts occurrences of each word in a slice.
func WordFrequency(words []string) map[string]int {
	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++
	}
	return freq
}

func main() {
	// Manual smoke test -- the real tests are in 01_tests_test.go
	fmt.Println("FizzBuzz(15):", FizzBuzz(15))
	fmt.Println("Reverse(\"golang\"):", Reverse("golang"))
	fmt.Println("Contains:", Contains([]string{"a", "b", "c"}, "b"))
	fmt.Println("WordFrequency:", WordFrequency([]string{"go", "python", "go"}))
}

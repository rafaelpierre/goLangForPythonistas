// Test file for 01_tests.go
//
// Run: go test .           -- run all tests
//      go test -v .        -- verbose output (shows each test name)
//      go test -run Fizz . -- run only tests matching "Fizz"
//      go test -bench . .  -- run benchmarks
//
// Python pytest equivalent:
//   pytest -v
//   pytest -k "fizz"
//   pytest --benchmark-only

package main

import (
	"testing"
)

// --- Table-driven test (the Go idiom) ---
// Python equivalent: @pytest.mark.parametrize("n,expected", [...])

func TestFizzBuzz(t *testing.T) {
	tests := []struct { // anonymous struct slice -- the standard table shape
		name     string
		input    int
		expected string
	}{
		{"multiple of 15", 15, "FizzBuzz"},
		{"multiple of 3", 9, "Fizz"},
		{"multiple of 5", 10, "Buzz"},
		{"neither", 7, "7"},
		{"one", 1, "1"},
		{"three", 3, "Fizz"},
		{"five", 5, "Buzz"},
	}

	for _, tc := range tests {
		tc := tc // capture range variable (needed pre-Go 1.22)
		t.Run(tc.name, func(t *testing.T) {
			got := FizzBuzz(tc.input)
			if got != tc.expected {
				t.Errorf("FizzBuzz(%d) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// --- Simple test with t.Fatal (stops on first failure) ---

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"golang", "gnalog"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"racecar", "racecar"},
	}

	for _, tc := range tests {
		got := Reverse(tc.input)
		if got != tc.expected {
			// t.Fatalf stops this test immediately (unlike t.Errorf which continues)
			t.Fatalf("Reverse(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestContains(t *testing.T) {
	haystack := []string{"apple", "banana", "cherry"}

	if !Contains(haystack, "banana") {
		t.Error("expected Contains to find 'banana'")
	}
	if Contains(haystack, "grape") {
		t.Error("expected Contains NOT to find 'grape'")
	}
	if Contains(nil, "anything") {
		t.Error("nil haystack should return false")
	}
}

func TestWordFrequency(t *testing.T) {
	input := []string{"go", "python", "go", "rust", "go", "python"}
	freq := WordFrequency(input)

	expected := map[string]int{"go": 3, "python": 2, "rust": 1}
	for word, count := range expected {
		if freq[word] != count {
			t.Errorf("freq[%q] = %d; want %d", word, freq[word], count)
		}
	}

	// Check no extra keys
	if len(freq) != len(expected) {
		t.Errorf("freq has %d keys; want %d", len(freq), len(expected))
	}
}

// --- Benchmark ---
// Run: go test -bench=BenchmarkReverse -benchmem .

func BenchmarkReverse(b *testing.B) {
	s := "the quick brown fox jumps over the lazy dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reverse(s)
	}
}

// --- TODO Tests ---
//
// TODO 1: Add a test TestFizzBuzzRange that calls FizzBuzz for n in [1..20]
// and spot-checks: n=1 -> "1", n=3 -> "Fizz", n=5 -> "Buzz", n=15 -> "FizzBuzz".
// Use t.Run subtests.
//
// TODO 2: The current Reverse() is byte-based and breaks on multi-byte UTF-8.
// Write a TestReverseUnicode test that passes "héllo" and shows the bug.
// Then fix Reverse() to operate on []rune instead of []byte and make the test pass.
// (Hint: convert to []rune, reverse, convert back to string.)
//
// TODO 3: Write a property-based test for Reverse: reversing a string twice
// should always return the original. Test it for 100 random strings generated
// with math/rand. This is the "round-trip" testing pattern.

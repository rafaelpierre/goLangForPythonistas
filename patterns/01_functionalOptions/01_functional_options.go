// Topic: Functional Options
//
// Problem: constructors with many optional parameters.
//
// Python solution: keyword arguments with defaults.
//   client = HTTPClient(timeout=30, retries=3, base_url="https://api.example.com")
//
// Go has no default arguments and no keyword arguments.
// The naive fix -- a giant Config struct passed to New() -- forces callers to
// name every field, and adding a new option is a breaking change if you embed
// the struct by value.
//
// The idiomatic Go solution: functional options.
// Each option is a function that mutates an internal config struct. Callers
// pass zero or more option functions; unknown options just aren't called.
// Adding a new WithXxx function is backwards-compatible.
//
// Real-world use: net/http.Server, grpc.Dial, zap.NewLogger all use this pattern.
//
// Run: go run 01_functional_options.go

//go:build ignore

package main

import (
	"fmt"
	"net/http"
	"time"
)

// --- The thing we're constructing ---

// HTTPClient wraps net/http.Client with retry logic and a base URL.
// Fields are unexported: callers configure via options, not direct field access.
type HTTPClient struct {
	baseURL    string
	timeout    time.Duration
	maxRetries int
	headers    map[string]string
	httpClient *http.Client
}

// --- The option type ---

// Option is a function that configures an HTTPClient.
// This is the functional-options pattern's key type.
type Option func(*HTTPClient)

// --- Option constructors ---
// Convention: prefix with "With". Each returns an Option (a closure).

func WithBaseURL(url string) Option {
	return func(c *HTTPClient) {
		c.baseURL = url
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *HTTPClient) {
		c.timeout = d
		c.httpClient.Timeout = d
	}
}

func WithMaxRetries(n int) Option {
	return func(c *HTTPClient) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

func WithHeader(key, value string) Option {
	return func(c *HTTPClient) {
		c.headers[key] = value
	}
}

// --- Constructor ---

// NewHTTPClient builds an HTTPClient with sensible defaults, then applies
// every provided option in order. Callers only specify what they need.
func NewHTTPClient(opts ...Option) *HTTPClient {
	// 1. Start with a fully-initialized default state.
	//    Zero values alone are not enough here because httpClient and headers
	//    need non-nil initialization -- the zero value of a map is nil, and
	//    writing to a nil map panics.
	c := &HTTPClient{
		baseURL:    "",
		timeout:    10 * time.Second,
		maxRetries: 3,
		headers:    make(map[string]string),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// 2. Apply each option in order. Options run as closures over 'c'.
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// --- Methods on HTTPClient ---

// Get performs a GET request with automatic retries on transient errors.
// This is a simplified version -- real production code would inspect status codes.
func (c *HTTPClient) Get(path string) (*http.Response, error) {
	url := c.baseURL + path

	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, reqErr := http.NewRequest(http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("HTTPClient.Get: build request: %w", reqErr)
		}
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}

		resp, err = c.httpClient.Do(req)
		if err == nil {
			return resp, nil // success
		}
		if attempt < c.maxRetries {
			fmt.Printf("  attempt %d failed (%v), retrying...\n", attempt+1, err)
		}
	}

	return nil, fmt.Errorf("HTTPClient.Get %q: all %d attempts failed: %w", url, c.maxRetries+1, err)
}

func (c *HTTPClient) String() string {
	return fmt.Sprintf("HTTPClient{baseURL=%q, timeout=%v, maxRetries=%d, headers=%v}",
		c.baseURL, c.timeout, c.maxRetries, c.headers)
}

func main() {
	// Default client -- all defaults apply
	defaultClient := NewHTTPClient()
	fmt.Println("default:", defaultClient)

	// Fully configured client -- pick only the options you need
	apiClient := NewHTTPClient(
		WithBaseURL("https://api.example.com"),
		WithTimeout(5*time.Second),
		WithMaxRetries(5),
		WithHeader("Authorization", "Bearer secret-token"),
		WithHeader("Accept", "application/json"),
	)
	fmt.Println("api client:", apiClient)

	// Minimal client -- only override what differs from defaults
	fastClient := NewHTTPClient(
		WithBaseURL("https://fast-service.internal"),
		WithTimeout(500*time.Millisecond),
		WithMaxRetries(0), // no retries for latency-sensitive paths
	)
	fmt.Println("fast client:", fastClient)

	// Note: the actual HTTP call below will fail (not a real server),
	// which demonstrates the retry + error wrapping working together.
	fmt.Println("\nAttempting request to a non-existent server:")
	_, err := fastClient.Get("/health")
	if err != nil {
		fmt.Println("error (expected):", err)
	}

	// TODO 1: Add a WithUserAgent(ua string) option that sets the
	// "User-Agent" header. Use it when constructing apiClient above.

	// TODO 2: Add a WithBasicAuth(user, pass string) option that stores
	// credentials on the struct and applies them in Get() via
	// req.SetBasicAuth(user, pass). Test it by printing req.Header.

	// TODO 3: The current options are applied in order. Add a
	// WithRetryDelay(d time.Duration) option and add an exponential-backoff
	// sleep between retries in Get(). Use time.Sleep(d * (1 << attempt))
	// so each attempt waits twice as long. Watch out: don't sleep after the
	// final failed attempt.

	// STRETCH: Add a WithRoundTripper(rt http.RoundTripper) option so that
	// in tests you can inject a fake transport that returns canned responses
	// without hitting a real network. This is the interface-DI pattern from
	// the next exercise applied to http.Client.
}

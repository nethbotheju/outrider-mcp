package search

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SearchResult holds a single search hit returned by any provider.
type SearchResult struct {
	Title       string
	URL         string
	Description string
}

// Provider is the interface that every search engine must implement.
type Provider interface {
	Name() string
	Search(ctx context.Context, query string, count int) ([]SearchResult, error)
}

// FormatResults renders search results as numbered text for LLM consumption.
func FormatResults(results []SearchResult) string {
	if len(results) == 0 {
		return "No search results found."
	}

	var buf strings.Builder
	for i, r := range results {
		fmt.Fprintf(&buf, "%d. %s\n   URL: %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description)
	}
	return buf.String()
}

// clamp restricts n to the inclusive range [min, max].
func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// newHTTPClient returns a shared *http.Client with sensible defaults.
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

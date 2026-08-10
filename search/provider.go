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
	// Optional rich metadata used mainly by SearXNG.
	PublishedDate string
	Source        string
	Journal       string
	DOI           string
	PDFURL        string
	Authors       []string
}

// Provider is the interface that every search engine must implement.
type Provider interface {
	Name() string
	DefaultCount() int
	MaxCount() int
	SupportsCategories() bool
	CategoryDescription() string
	Search(ctx context.Context, query string, count int, category string) ([]SearchResult, error)
}

// FormatResults renders search results as numbered text for LLM consumption.
func FormatResults(results []SearchResult) string {
	if len(results) == 0 {
		return "No search results found."
	}

	var buf strings.Builder
	for i, r := range results {
		fmt.Fprintf(&buf, "%d. %s\n   URL: %s", i+1, r.Title, r.URL)
		if r.Description != "" {
			fmt.Fprintf(&buf, "\n   %s", r.Description)
		}
		if r.PublishedDate != "" {
			date := r.PublishedDate
			if len(date) > 10 {
				date = date[:10]
			}
			fmt.Fprintf(&buf, "\n   Published: %s", date)
		}
		if r.Source != "" {
			fmt.Fprintf(&buf, "\n   Source: %s", r.Source)
		}
		if r.Journal != "" {
			fmt.Fprintf(&buf, "\n   Journal: %s", r.Journal)
		}
		if r.DOI != "" {
			fmt.Fprintf(&buf, "\n   DOI: %s", r.DOI)
		}
		if len(r.Authors) > 0 {
			fmt.Fprintf(&buf, "\n   Authors: %s", strings.Join(r.Authors, ", "))
		}
		if r.PDFURL != "" {
			fmt.Fprintf(&buf, "\n   PDF: %s", r.PDFURL)
		}
		buf.WriteString("\n\n")
	}
	return strings.TrimRight(buf.String(), "\n")
}

// resolveCount clamps n to a valid count for the provider.
func resolveCount(n, def, max int) int {
	if n <= 0 {
		return def
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

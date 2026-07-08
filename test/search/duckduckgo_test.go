package search_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/nethbotheju/outrider-mcp/search"
)

// TestDuckDuckGoSearchLive hits the real DuckDuckGo HTML endpoint.
// It verifies that a search returns results with populated fields.
func TestDuckDuckGoSearchLive(t *testing.T) {
	provider := search.NewDuckDuckGoProvider()

	if provider.Name() != "DuckDuckGo" {
		t.Fatalf("expected provider name 'DuckDuckGo', got %q", provider.Name())
	}

	results, err := provider.Search(context.Background(), "golang testing package", 5, "")
	if err != nil {
		t.Fatalf("Search() returned error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Search() returned 0 results, expected at least 1")
	}

	for i, r := range results {
		if r.Title == "" {
			t.Errorf("result[%d]: Title is empty", i)
		}
		if r.URL == "" {
			t.Errorf("result[%d]: URL is empty", i)
		}
		if !strings.HasPrefix(r.URL, "http://") && !strings.HasPrefix(r.URL, "https://") {
			t.Errorf("result[%d]: URL %q does not have http(s) scheme", i, r.URL)
		}
		if r.Description == "" {
			t.Errorf("result[%d]: Description is empty", i)
		}
	}

	// Print results in a readable format.
	t.Log("\n" + formatResultList(results))
}

// TestDuckDuckGoSearchCountRespected verifies that the count parameter is respected.
func TestDuckDuckGoSearchCountRespected(t *testing.T) {
	provider := search.NewDuckDuckGoProvider()

	results, err := provider.Search(context.Background(), "golang", 3, "")
	if err != nil {
		t.Fatalf("Search() returned error: %v", err)
	}

	if len(results) > 3 {
		t.Errorf("requested 3 results, got %d", len(results))
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	t.Logf("\nRequested 3 results, got %d — OK", len(results))
}

// TestFormatResultsLive verifies FormatResults with live data.
func TestFormatResultsLive(t *testing.T) {
	provider := search.NewDuckDuckGoProvider()

	results, err := provider.Search(context.Background(), "golang", 2, "")
	if err != nil {
		t.Fatalf("Search() returned error: %v", err)
	}

	formatted := search.FormatResults(results)
	if formatted == "" {
		t.Fatal("FormatResults returned empty string")
	}
	if !strings.Contains(formatted, "1.") {
		t.Error("FormatResults missing numbered list marker '1.'")
	}
	for _, r := range results {
		if !strings.Contains(formatted, r.Title) {
			t.Errorf("FormatResults output missing title %q", r.Title)
		}
	}

	t.Log("\n" + formatResultList(results))
	t.Logf("\nFormatted output:\n%s", formatted)
}

// formatResultList renders search results as a vertical list with no truncation.
func formatResultList(results []search.SearchResult) string {
	if len(results) == 0 {
		return "  (no results)"
	}

	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "  Result %d\n", i+1)
		fmt.Fprintf(&b, "  ├─ Title:       %s\n", r.Title)
		fmt.Fprintf(&b, "  ├─ URL:         %s\n", r.URL)
		fmt.Fprintf(&b, "  └─ Description: %s\n", r.Description)
		if i < len(results)-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	return b.String()
}

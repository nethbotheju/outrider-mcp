package fetcher_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nethbotheju/web-search-mcp/fetcher"
)

// TestFetcherHTMPLive fetches a real HTML page and verifies content extraction.
func TestFetcherHTMLLive(t *testing.T) {
	f := fetcher.NewFetcher()

	result, err := f.FetchURL(context.Background(), "https://go.dev/", 5000)
	if err != nil {
		t.Fatalf("FetchURL() returned error: %v", err)
	}

	if result.URL != "https://go.dev/" {
		t.Errorf("expected URL 'https://go.dev/', got %q", result.URL)
	}
	if result.Title == "" {
		t.Error("expected non-empty Title")
	}
	if result.Content == "" {
		t.Error("expected non-empty Content")
	}

	t.Logf("Title:   %s", result.Title)
	t.Logf("URL:     %s", result.URL)
	t.Logf("Content: %s", truncate(result.Content, 200))
}

// TestFetcherPlainTextLive fetches a plain-text endpoint (robots.txt).
func TestFetcherPlainTextLive(t *testing.T) {
	f := fetcher.NewFetcher()

	result, err := f.FetchURL(context.Background(), "https://go.dev/robots.txt", 5000)
	if err != nil {
		t.Fatalf("FetchURL() returned error: %v", err)
	}

	if result.Content == "" {
		t.Error("expected non-empty Content for robots.txt")
	}
	if !strings.Contains(result.Content, "User-agent") {
		t.Error("expected robots.txt to contain 'User-agent'")
	}

	t.Logf("Title:   %s", result.Title)
	t.Logf("Content: %s", truncate(result.Content, 200))
}

// TestFetcherTruncationLive verifies that maxLength truncation works on a real page.
func TestFetcherTruncationLive(t *testing.T) {
	f := fetcher.NewFetcher()

	result, err := f.FetchURL(context.Background(), "https://go.dev/", 500)
	if err != nil {
		t.Fatalf("FetchURL() returned error: %v", err)
	}

	if !strings.Contains(result.Content, "[Content truncated at 500 characters]") {
		t.Error("expected truncation marker in content")
	}

	t.Logf("Content length with truncation marker: %d chars", len(result.Content))
}

// TestFetcherInvalidURL verifies error handling for bad URLs.
func TestFetcherInvalidURL(t *testing.T) {
	f := fetcher.NewFetcher()

	_, err := f.FetchURL(context.Background(), "ftp://example.com/file", 5000)
	if err == nil {
		t.Fatal("expected error for unsupported URL scheme, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported URL scheme") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// truncate is a helper to limit log output.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

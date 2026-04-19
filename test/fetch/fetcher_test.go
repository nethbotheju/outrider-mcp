package fetcher_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nethbotheju/outrider-mcp/fetcher"
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

// TestFetcherJinaLive fetches a real HTML page via the Jina Reader API (Tier 1).
func TestFetcherJinaLive(t *testing.T) {
	f := fetcher.NewFetcher()

	result, err := f.FetchViaJina(context.Background(), "https://go.dev/")
	if err != nil {
		t.Fatalf("FetchViaJina() error: %v", err)
	}

	if result.URL != "https://go.dev/" {
		t.Errorf("URL = %q, want %q", result.URL, "https://go.dev/")
	}
	if result.Title == "" {
		t.Error("expected non-empty Title")
	}
	if result.Content == "" {
		t.Error("expected non-empty Content")
	}

	t.Logf("Title:   %s", result.Title)
	t.Logf("Content: %s", truncate(result.Content, 200))
}

// TestFetcherJinaInvalidURL verifies error handling when Jina receives a non-existent domain.
func TestFetcherJinaInvalidURL(t *testing.T) {
	f := fetcher.NewFetcher()

	_, err := f.FetchViaJina(context.Background(), "https://this-domain-does-not-exist-12345.com")
	if err == nil {
		t.Fatal("expected error for non-existent domain, got nil")
	}
	t.Logf("Jina error for bad domain (expected): %v", err)
}

// TestFetcherBrowserLive fetches a real HTML page via chromedp headless browser (Tier 2).
func TestFetcherBrowserLive(t *testing.T) {
	if !fetcher.ChromeAvailable() {
		t.Skip("Chrome not available, skipping browser test")
	}

	f := fetcher.NewFetcher()

	result, err := f.FetchViaBrowser(context.Background(), "https://go.dev/")
	if err != nil {
		t.Fatalf("FetchViaBrowser() error: %v", err)
	}

	if result.URL != "https://go.dev/" {
		t.Errorf("URL = %q, want %q", result.URL, "https://go.dev/")
	}
	if result.Title == "" {
		t.Error("expected non-empty Title")
	}
	if result.Content == "" {
		t.Error("expected non-empty Content")
	}

	t.Logf("Title:   %s", result.Title)
	t.Logf("Content: %s", truncate(result.Content, 200))
}

// TestFetcherBrowserInvalidURL verifies error handling when the browser fetches a non-existent domain.
func TestFetcherBrowserInvalidURL(t *testing.T) {
	if !fetcher.ChromeAvailable() {
		t.Skip("Chrome not available, skipping browser test")
	}

	f := fetcher.NewFetcher()

	_, err := f.FetchViaBrowser(context.Background(), "https://this-domain-does-not-exist-12345.com")
	if err == nil {
		t.Fatal("expected error for non-existent domain, got nil")
	}
	t.Logf("Browser error for bad domain (expected): %v", err)
}

// TestFetcherStaticLive fetches a real HTML page via static HTTP + readability (Tier 3).
func TestFetcherStaticLive(t *testing.T) {
	f := fetcher.NewFetcher()

	result, err := f.FetchViaStatic(context.Background(), "https://go.dev/")
	if err != nil {
		t.Fatalf("FetchViaStatic() error: %v", err)
	}

	if result.URL != "https://go.dev/" {
		t.Errorf("URL = %q, want %q", result.URL, "https://go.dev/")
	}
	if result.Title == "" {
		t.Error("expected non-empty Title")
	}
	if result.Content == "" {
		t.Error("expected non-empty Content")
	}

	t.Logf("Title:   %s", result.Title)
	t.Logf("Content: %s", truncate(result.Content, 200))
}

// TestFetcherStaticPlainText fetches a plain-text endpoint (robots.txt) via static HTTP.
func TestFetcherStaticPlainText(t *testing.T) {
	f := fetcher.NewFetcher()

	result, err := f.FetchViaStatic(context.Background(), "https://go.dev/robots.txt")
	if err != nil {
		t.Fatalf("FetchViaStatic() error: %v", err)
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

// TestFetcherStaticInvalidURL verifies error handling when static HTTP fetches a non-existent domain.
func TestFetcherStaticInvalidURL(t *testing.T) {
	f := fetcher.NewFetcher()

	_, err := f.FetchViaStatic(context.Background(), "https://this-domain-does-not-exist-12345.com")
	if err == nil {
		t.Fatal("expected error for non-existent domain, got nil")
	}
	t.Logf("Static error for bad domain (expected): %v", err)
}

// TestFetcherFallbackAllFail verifies that the fallback chain returns an aggregated error when all tiers fail.
func TestFetcherFallbackAllFail(t *testing.T) {
	f := fetcher.NewFetcher()

	_, err := f.FetchURL(context.Background(), "https://this-domain-does-not-exist-12345.com", 5000)
	if err == nil {
		t.Fatal("expected error when all tiers fail, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "all tiers failed") {
		t.Errorf("expected 'all tiers failed' in error, got: %v", err)
	}
	if !strings.Contains(errMsg, "jina:") {
		t.Errorf("expected 'jina:' in error, got: %v", err)
	}
	if !strings.Contains(errMsg, "static:") {
		t.Errorf("expected 'static:' in error, got: %v", err)
	}

	t.Logf("Fallback chain error (expected):\n%v", err)
}

// truncate is a helper to limit log output.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

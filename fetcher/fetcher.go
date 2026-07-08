package fetcher

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/nethbotheju/outrider-mcp/config"
)

// Format controls how fetched content is sanitized.
type Format string

const (
	// FormatLean strips link URLs and images for minimal token usage.
	FormatLean Format = "lean"
	// FormatMarkdown preserves full links and images.
	FormatMarkdown Format = "markdown"
)

// ParseFormat converts a string to a Format, defaulting to lean.
func ParseFormat(s string) Format {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "markdown":
		return FormatMarkdown
	default:
		return FormatLean
	}
}

// FetchResult holds the extracted content from a fetched URL.
type FetchResult struct {
	Content string
	URL     string
	Title   string
	Format  Format
}

// Fetcher downloads URLs and extracts readable Markdown content using a tiered strategy:
// Tier 1: Jina Reader API (best quality, handles JS)
// Tier 2: chromedp headless browser (good quality, handles JS)
// Tier 3: static HTTP + readability (no JS rendering)
type Fetcher struct {
	jinaEnabled    bool
	jinaAPIKey     string
	browserEnabled bool
	timeout        time.Duration
}

// NewFetcher creates a Fetcher from configuration.
func NewFetcher(cfg config.FetchConfig) *Fetcher {
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Fetcher{
		jinaEnabled:    cfg.Jina.Enabled,
		jinaAPIKey:     cfg.Jina.APIKey,
		browserEnabled: cfg.Browser.Enabled,
		timeout:        timeout,
	}
}

// maxBodyBytes caps the response body at 10 MB to prevent OOM on huge pages.
const maxBodyBytes = 10 << 20

// FetchURL fetches rawURL and returns its content as Markdown, truncated to maxLength.
func (f *Fetcher) FetchURL(ctx context.Context, rawURL string, maxLength int, format string) (*FetchResult, error) {
	fmtVal := ParseFormat(format)

	if maxLength <= 0 {
		maxLength = 10000
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (only http and https are supported)", parsedURL.Scheme)
	}

	result, err := f.fetchWithFallback(ctx, rawURL, fmtVal)
	if err != nil {
		return nil, err
	}

	if len(result.Content) > maxLength {
		result.Content = result.Content[:maxLength]
		result.Content += fmt.Sprintf("\n\n[Content truncated at %d characters]", maxLength)
	}

	return result, nil
}

// fetchWithFallback tries each tier in order, falling through on failure.
func (f *Fetcher) fetchWithFallback(ctx context.Context, rawURL string, format Format) (*FetchResult, error) {
	var errs []string

	if f.jinaEnabled {
		jinaCtx, jinaCancel := context.WithTimeout(ctx, f.timeout)
		defer jinaCancel()
		result, err := f.fetchViaJina(jinaCtx, rawURL, format)
		if err == nil {
			return result, nil
		}
		errs = append(errs, fmt.Sprintf("jina: %v", err))
	}

	if f.browserEnabled && chromeAvailable() {
		browserCtx, browserCancel := context.WithTimeout(ctx, 30*time.Second)
		defer browserCancel()
		result, err := f.fetchViaBrowser(browserCtx, rawURL, format)
		if err == nil {
			return result, nil
		}
		errs = append(errs, fmt.Sprintf("browser: %v", err))
	}

	staticCtx, staticCancel := context.WithTimeout(ctx, f.timeout)
	defer staticCancel()
	static := newHTTPStaticFetcher(f.timeout)
	result, err := static.fetch(staticCtx, rawURL, format)
	if err == nil {
		return result, nil
	}
	errs = append(errs, fmt.Sprintf("local: %v", err))

	return nil, fmt.Errorf("all tiers failed:\n  - %s", strings.Join(errs, "\n  - "))
}

// FetchViaJina exposes the Jina Reader tier for tests.
func (f *Fetcher) FetchViaJina(ctx context.Context, rawURL string) (*FetchResult, error) {
	return f.fetchViaJina(ctx, rawURL, FormatLean)
}

// FetchViaBrowser exposes the headless-browser tier for tests.
func (f *Fetcher) FetchViaBrowser(ctx context.Context, rawURL string) (*FetchResult, error) {
	return f.fetchViaBrowser(ctx, rawURL, FormatLean)
}

// FetchViaStatic exposes the local/static HTTP tier for tests.
func (f *Fetcher) FetchViaStatic(ctx context.Context, rawURL string) (*FetchResult, error) {
	return newHTTPStaticFetcher(f.timeout).fetch(ctx, rawURL, FormatLean)
}

// ChromeAvailable reports whether a Chrome/Chromium binary is available.
func (f *Fetcher) ChromeAvailable() bool {
	return chromeAvailable()
}

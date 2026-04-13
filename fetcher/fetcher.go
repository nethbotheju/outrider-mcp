package fetcher

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// FetchResult holds the extracted content from a fetched URL.
type FetchResult struct {
	Content string
	URL     string
	Title   string
}

// Fetcher downloads URLs and extracts readable Markdown content using a 3-tier strategy:
// Tier 1: Jina Reader API (best quality, handles JS)
// Tier 2: chromedp headless browser + readability (good quality, handles JS)
// Tier 3: static HTTP + readability (decent quality, no JS rendering)
type Fetcher struct {
	jinaAPIKey string
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		jinaAPIKey: os.Getenv("JINA_API_KEY"),
	}
}

// maxBodyBytes caps the response body at 10 MB to prevent OOM on huge pages.
const maxBodyBytes = 10 << 20

// FetchURL fetches rawURL and returns its content as Markdown, truncated to maxLength.
func (f *Fetcher) FetchURL(ctx context.Context, rawURL string, maxLength int) (*FetchResult, error) {
	if maxLength <= 0 {
		maxLength = 50000
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (only http and https are supported)", parsedURL.Scheme)
	}

	result, err := f.fetchWithFallback(ctx, rawURL)
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
func (f *Fetcher) fetchWithFallback(ctx context.Context, rawURL string) (*FetchResult, error) {
	var errs []string

	// Tier 1: Jina Reader
	jinaCtx, jinaCancel := context.WithTimeout(ctx, 15*time.Second)
	defer jinaCancel()

	result, err := f.fetchViaJina(jinaCtx, rawURL)
	if err == nil {
		return result, nil
	}
	errs = append(errs, fmt.Sprintf("jina: %v", err))

	// Tier 2: chromedp (only if Chrome is available on the system)
	if chromeAvailable() {
		browserCtx, browserCancel := context.WithTimeout(ctx, 30*time.Second)
		defer browserCancel()

		result, err = f.fetchViaBrowser(browserCtx, rawURL)
		if err == nil {
			return result, nil
		}
		errs = append(errs, fmt.Sprintf("browser: %v", err))
	}

	// Tier 3: static HTTP + readability
	staticCtx, staticCancel := context.WithTimeout(ctx, 15*time.Second)
	defer staticCancel()

	static := newHTTPStaticFetcher()
	result, err = static.fetch(staticCtx, rawURL)
	if err == nil {
		return result, nil
	}
	errs = append(errs, fmt.Sprintf("static: %v", err))

	return nil, fmt.Errorf("all tiers failed:\n  - %s", strings.Join(errs, "\n  - "))
}

func (f *Fetcher) FetchViaJina(ctx context.Context, rawURL string) (*FetchResult, error) {
	return f.fetchViaJina(ctx, rawURL)
}

func (f *Fetcher) FetchViaBrowser(ctx context.Context, rawURL string) (*FetchResult, error) {
	return f.fetchViaBrowser(ctx, rawURL)
}

func (f *Fetcher) FetchViaStatic(ctx context.Context, rawURL string) (*FetchResult, error) {
	return newHTTPStaticFetcher().fetch(ctx, rawURL)
}

func ChromeAvailable() bool {
	return chromeAvailable()
}

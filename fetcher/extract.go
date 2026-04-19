package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"
)

type httpStaticFetcher struct {
	client *http.Client
}

func newHTTPStaticFetcher() *httpStaticFetcher {
	return &httpStaticFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *httpStaticFetcher) fetch(ctx context.Context, rawURL string) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("static: creating request: %w", err)
	}

	req.Header.Set("User-Agent", "outrider/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("static: fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("static: HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("static: reading response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")

	// Non-HTML content: return as-is
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		parsedURL, _ := url.Parse(rawURL)
		return &FetchResult{
			Content: string(body),
			URL:     rawURL,
			Title:   parsedURL.Host,
		}, nil
	}

	return extractContent(body, rawURL)
}

// extractContent is the shared pipeline: readability extracts the article,
// then html-to-markdown converts it to Markdown.
func extractContent(htmlBody []byte, rawURL string) (*FetchResult, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	article, err := readability.FromReader(strings.NewReader(string(htmlBody)), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("readability: %w", err)
	}

	// Render the article's DOM node to HTML, then convert to Markdown
	var htmlBuf bytes.Buffer
	if err := article.RenderHTML(&htmlBuf); err != nil {
		return nil, fmt.Errorf("rendering article HTML: %w", err)
	}

	markdown, err := md.ConvertString(htmlBuf.String())
	if err != nil {
		return nil, fmt.Errorf("markdown conversion: %w", err)
	}

	return &FetchResult{
		Content: markdown,
		URL:     rawURL,
		Title:   article.Title(),
	}, nil
}

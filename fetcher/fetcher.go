package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// FetchResult holds the extracted content from a fetched URL.
type FetchResult struct {
	Content string
	URL     string
	Title   string
}

// Fetcher downloads URLs and extracts readable text from HTML.
type Fetcher struct {
	HTTPClient *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// maxBodyBytes caps the response body at 10 MB to prevent OOM on huge pages.
const maxBodyBytes = 10 << 20

// FetchURL fetches rawURL and returns its content as clean text, truncated to maxLength.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "web-search-mcp/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain,application/json")

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")

	var content, title string

	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml") {
		content, title, err = extractHTMLContent(body)
		if err != nil {
			return nil, fmt.Errorf("extracting HTML content: %w", err)
		}
	} else {
		content = string(body)
		title = parsedURL.Host
	}

	if len(content) > maxLength {
		content = content[:maxLength]
		content += fmt.Sprintf("\n\n[Content truncated at %d characters]", maxLength)
	}

	return &FetchResult{
		Content: content,
		URL:     rawURL,
		Title:   title,
	}, nil
}

// noiseSelector targets elements that add no value for LLM consumption.
const noiseSelector = "script, style, nav, footer, header, iframe, noscript, svg, img"

// extractHTMLContent parses HTML, strips noise, and returns the main readable text.
func extractHTMLContent(html []byte) (string, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return "", "", fmt.Errorf("parsing HTML: %w", err)
	}

	title := doc.Find("title").Text()

	doc.Find(noiseSelector).Remove()

	// Prefer semantic content containers; fall back to full body.
	selection := doc.Find("main, article, [role='main']").First()
	if selection.Length() == 0 {
		selection = doc.Find("body")
	}

	text := cleanWhitespace(selection.Text())

	return text, title, nil
}

// whitespaceRe collapses runs of 3+ newlines into 2.
var whitespaceRe = regexp.MustCompile(`\n{3,}`)

// cleanWhitespace normalises whitespace for cleaner LLM input.
func cleanWhitespace(text string) string {
	text = strings.ReplaceAll(text, "\t", " ")
	text = whitespaceRe.ReplaceAllString(text, "\n\n")

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

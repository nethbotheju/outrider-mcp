package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SearchResult holds a single search hit returned to the caller.
type SearchResult struct {
	Title       string
	URL         string
	Description string
}

// BraveClient wraps the Brave Search API.
type BraveClient struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewBraveClient(apiKey string) *BraveClient {
	return &BraveClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// braveAPIResponse maps the top-level JSON envelope from the Brave Search API.
type braveAPIResponse struct {
	Web struct {
		Results []braveWebResult `json:"results"`
	} `json:"web"`
}

type braveWebResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Search queries Brave and returns a slice of SearchResult. Count is clamped to [1, 20].
func (c *BraveClient) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("BRAVE_SEARCH_API_KEY is not set")
	}

	count = clamp(count, 1, 20)

	searchURL := fmt.Sprintf(
		"https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
		url.QueryEscape(query), count,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-Subscription-Token", c.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing search request: %w", err)
	}
	defer resp.Body.Close()

	if err := statusToError(resp.StatusCode); err != nil {
		return nil, err
	}

	var apiResp braveAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding search response: %w", err)
	}

	results := make([]SearchResult, 0, len(apiResp.Web.Results))
	for _, r := range apiResp.Web.Results {
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
		})
	}

	return results, nil
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

// statusToError converts non-200 HTTP status codes to descriptive errors.
func statusToError(code int) error {
	switch code {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid Brave Search API key (401 Unauthorized)")
	case http.StatusTooManyRequests:
		return fmt.Errorf("Brave Search API rate limit exceeded (429 Too Many Requests)")
	default:
		return fmt.Errorf("Brave Search API returned status %d", code)
	}
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

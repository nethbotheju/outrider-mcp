package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// SearXNGProvider scrapes a self-hosted SearXNG instance's JSON API.
type SearXNGProvider struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewSearXNGProvider creates a new SearXNG search provider.
func NewSearXNGProvider(baseURL string) *SearXNGProvider {
	return &SearXNGProvider{
		HTTPClient: newHTTPClient(),
		BaseURL:    baseURL,
	}
}

// Name returns the provider identifier.
func (p *SearXNGProvider) Name() string { return "SearXNG" }

// Search queries the SearXNG JSON API and returns parsed search results.
func (p *SearXNGProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	count = clamp(count, 1, 20)

	searchURL := p.BaseURL + "/search?q=" + url.QueryEscape(query) + "&format=json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; outrider/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SearXNG returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	results, err := parseSearXNGResults(body, count)
	if err != nil {
		return nil, err
	}
	log.Printf("[search] SearXNG returned %d results for query %q\n", len(results), query)
	return results, nil
}

// searxngResponse represents the JSON structure returned by SearXNG's /search endpoint.
type searxngResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func parseSearXNGResults(data []byte, count int) ([]SearchResult, error) {
	var resp searxngResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing SearXNG JSON response: %w", err)
	}

	results := make([]SearchResult, 0, count)
	for _, r := range resp.Results {
		if len(results) >= count {
			break
		}
		if r.Title == "" || r.URL == "" {
			continue
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
		})
	}

	return results, nil
}

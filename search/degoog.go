package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// DegoogProvider queries a self-hosted Degoog instance's JSON search API.
type DegoogProvider struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewDegoogProvider creates a new Degoog search provider.
func NewDegoogProvider(baseURL string) *DegoogProvider {
	return &DegoogProvider{
		HTTPClient: newHTTPClient(),
		BaseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
	}
}

// Name returns the provider identifier.
func (p *DegoogProvider) Name() string { return "Degoog" }

// DefaultCount is the number of results returned when count is not specified.
func (p *DegoogProvider) DefaultCount() int { return 10 }

// MaxCount is the hard upper bound for result count.
func (p *DegoogProvider) MaxCount() int { return 20 }

// SupportsCategories reports whether this provider supports category filtering.
func (p *DegoogProvider) SupportsCategories() bool { return false }

// CategoryDescription is empty for Degoog.
func (p *DegoogProvider) CategoryDescription() string { return "" }

// Search queries the Degoog JSON API and returns parsed search results.
func (p *DegoogProvider) Search(ctx context.Context, query string, count int, category string) ([]SearchResult, error) {
	count = resolveCount(count, p.DefaultCount(), p.MaxCount())

	searchURL := p.BaseURL + "/api/search?q=" + url.QueryEscape(query)

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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("Degoog returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	results, err := parseDegoogResults(body, count)
	if err != nil {
		return nil, err
	}
	log.Printf("[search] Degoog returned %d results for query %q", len(results), query)
	return results, nil
}

// degoogResponse represents the JSON structure returned by Degoog's /api/search endpoint.
type degoogResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
		Content string `json:"content"`
	} `json:"results"`
}

func parseDegoogResults(data []byte, count int) ([]SearchResult, error) {
	var resp degoogResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing Degoog JSON response: %w", err)
	}

	results := make([]SearchResult, 0, count)
	for _, r := range resp.Results {
		if len(results) >= count {
			break
		}
		title := strings.TrimSpace(r.Title)
		u := strings.TrimSpace(r.URL)
		if title == "" || u == "" {
			continue
		}
		description := strings.TrimSpace(r.Snippet)
		if description == "" {
			description = strings.TrimSpace(r.Content)
		}
		results = append(results, SearchResult{
			Title:       title,
			URL:         u,
			Description: description,
		})
	}

	return results, nil
}

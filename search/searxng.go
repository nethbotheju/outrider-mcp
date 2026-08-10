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

// SearchCategory is a SearXNG category filter.
type SearchCategory string

// Supported SearXNG search categories.
const (
	CategoryGeneral SearchCategory = "general"
	CategoryScience SearchCategory = "science"
	CategoryNews    SearchCategory = "news"
	CategoryIT      SearchCategory = "it"
)

// CategoryChoices returns all supported category values.
func CategoryChoices() []SearchCategory {
	return []SearchCategory{CategoryGeneral, CategoryScience, CategoryNews, CategoryIT}
}

// SearXNGProvider queries a self-hosted SearXNG instance's JSON API.
type SearXNGProvider struct {
	HTTPClient *http.Client
	BaseURL    string
	APIKey     string
}

// NewSearXNGProvider creates a new SearXNG search provider.
func NewSearXNGProvider(baseURL, apiKey string) *SearXNGProvider {
	return &SearXNGProvider{
		HTTPClient: newHTTPClient(),
		BaseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		APIKey:     apiKey,
	}
}

// Name returns the provider identifier.
func (p *SearXNGProvider) Name() string { return "SearXNG" }

// DefaultCount is the number of results returned when count is not specified.
func (p *SearXNGProvider) DefaultCount() int { return 5 }

// MaxCount is the hard upper bound for result count.
func (p *SearXNGProvider) MaxCount() int { return 10 }

// SupportsCategories reports whether this provider supports category filtering.
func (p *SearXNGProvider) SupportsCategories() bool { return true }

// CategoryDescription explains the supported categories for tool schemas.
func (p *SearXNGProvider) CategoryDescription() string {
	return "Search category. Determines which sources SearXNG queries and what metadata each result carries. " +
		"'general' (default): broad web search (Google, Bing, DuckDuckGo, Wikipedia) — best for most queries. " +
		"'science': academic and scientific sources (arXiv, PubMed, PDBe, Semantic Scholar, CrossRef) — use for research papers; results include authors, journal, DOI, and PDF link when available. " +
		"'news': current events and reporting (Google/Bing/DuckDuckGo News) — use for recent events; results include publish date and source/publisher. " +
		"'it': developer and technical Q&A (Stack Overflow, SuperUser, Ask Ubuntu, MDN) — use for code, tooling, errors, and sysadmin questions."
}

// Search queries the SearXNG JSON API and returns parsed search results.
func (p *SearXNGProvider) Search(ctx context.Context, query string, count int, category string) ([]SearchResult, error) {
	count = resolveCount(count, p.DefaultCount(), p.MaxCount())

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("pageno", "1")
	if category != "" {
		params.Set("categories", category)
	}
	searchURL := p.BaseURL + "/search?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; outrider/1.0)")
	req.Header.Set("Accept", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("SearXNG returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	results, err := parseSearXNGResults(body, count)
	if err != nil {
		return nil, err
	}
	log.Printf("[search] SearXNG returned %d results for query %q", len(results), query)
	return results, nil
}

// searxngResponse represents the JSON structure returned by SearXNG's /search endpoint.
type searxngResponse struct {
	Results []struct {
		Title         string   `json:"title"`
		URL           string   `json:"url"`
		Content       string   `json:"content"`
		PublishedDate string   `json:"publishedDate"`
		Source        string   `json:"source"`
		Journal       string   `json:"journal"`
		DOI           string   `json:"doi"`
		PDFURL        string   `json:"pdf_url"`
		Authors       []string `json:"authors"`
		Author        any      `json:"author"`
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
		title := strings.TrimSpace(r.Title)
		u := strings.TrimSpace(r.URL)
		if title == "" || u == "" {
			continue
		}
		results = append(results, SearchResult{
			Title:         title,
			URL:           u,
			Description:   strings.TrimSpace(r.Content),
			PublishedDate: normStr(r.PublishedDate),
			Source:        normStr(r.Source),
			Journal:       normStr(r.Journal),
			DOI:           normStr(r.DOI),
			PDFURL:        normStr(r.PDFURL),
			Authors:       normalizeAuthors(r.Authors, r.Author),
		})
	}

	return results, nil
}

func normStr(v string) string {
	s := strings.TrimSpace(v)
	if s == "" {
		return ""
	}
	return s
}

func normalizeAuthors(authors []string, author any) []string {
	if len(authors) > 0 {
		out := make([]string, 0, len(authors))
		for _, a := range authors {
			if a = strings.TrimSpace(a); a != "" {
				out = append(out, a)
			}
		}
		if len(out) > 0 {
			return out
		}
	}

	if author == nil {
		return nil
	}

	switch v := author.(type) {
	case string:
		if s := strings.TrimSpace(v); s != "" {
			return []string{s}
		}
	case []any:
		out := make([]string, 0, len(v))
		for _, a := range v {
			out = append(out, authorName(a))
		}
		if len(out) > 0 {
			return out
		}
	case map[string]any:
		if s := authorName(v); s != "" {
			return []string{s}
		}
	}

	return nil
}

func authorName(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case map[string]any:
		for _, key := range []string{"name", "Name", "author"} {
			if s, ok := x[key].(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

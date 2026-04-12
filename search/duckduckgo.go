package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DuckDuckGoProvider scrapes the DuckDuckGo HTML search endpoint.
type DuckDuckGoProvider struct {
	HTTPClient *http.Client
}

// NewDuckDuckGoProvider creates a new DuckDuckGo search provider.
func NewDuckDuckGoProvider() *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		HTTPClient: newHTTPClient(),
	}
}

// Name returns the provider identifier.
func (p *DuckDuckGoProvider) Name() string { return "DuckDuckGo" }

// Search queries DuckDuckGo HTML and parses the result page.
func (p *DuckDuckGoProvider) Search(ctx context.Context, query string, count int) ([]SearchResult, error) {
	count = clamp(count, 1, 20)

	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; web-search-mcp/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DuckDuckGo returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return parseDuckDuckGoResults(body, count)
}

// parseDuckDuckGoResults extracts search results from the DuckDuckGo HTML page.
func parseDuckDuckGoResults(html []byte, count int) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html)))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var results []SearchResult

	doc.Find("div.result").Each(func(i int, s *goquery.Selection) {
		if len(results) >= count {
			return
		}

		titleEl := s.Find("a.result__a")
		title := strings.TrimSpace(titleEl.Text())

		href, _ := titleEl.Attr("href")
		realURL := extractDDGURL(href)

		snippetEl := s.Find("a.result__snippet")
		snippet := strings.TrimSpace(snippetEl.Text())

		if title == "" || realURL == "" {
			return
		}

		results = append(results, SearchResult{
			Title:       title,
			URL:          realURL,
			Description: snippet,
		})
	})

	return results, nil
}

// extractDDGURL resolves the actual URL from a DuckDuckGo redirect link.
func extractDDGURL(href string) string {
	if href == "" {
		return ""
	}

	// Normalise scheme-relative links.
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}

	u, err := url.Parse(href)
	if err != nil {
		return href
	}

	// If this is a DDG redirect link, extract the real URL from "uddg".
	if strings.Contains(u.Host, "duckduckgo.com") && u.Path == "/l/" {
		if real := u.Query().Get("uddg"); real != "" {
			return real
		}
	}

	return href
}

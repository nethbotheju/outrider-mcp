package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (f *Fetcher) fetchViaJina(ctx context.Context, rawURL string, format Format) (*FetchResult, error) {
	jinaURL := "https://r.jina.ai/" + rawURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jinaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("jina: creating request: %w", err)
	}

	req.Header.Set("Accept", "text/plain")
	if f.jinaAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+f.jinaAPIKey)
	}
	if format == FormatLean {
		req.Header.Set("X-Retain-Links", "text")
		req.Header.Set("X-Retain-Images", "none")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jina: fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("jina: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("jina: reading response: %w", err)
	}

	title, content := parseJinaResponse(string(body))
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("jina: returned empty content")
	}

	return &FetchResult{
		Content: content,
		URL:     rawURL,
		Title:   title,
		Format:  format,
	}, nil
}

// parseJinaResponse extracts the Title and strips Jina's metadata header
// from the response. Jina returns:
//
//	Title: <title>
//	URL Source: <url>
//	Markdown Content:
//	<actual content>
func parseJinaResponse(raw string) (title string, content string) {
	title, content = "", raw

	lines := strings.Split(raw, "\n")

	var contentStartIdx int
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Title:") {
			title = strings.TrimSpace(strings.TrimPrefix(trimmed, "Title:"))
		}
		if trimmed == "Markdown Content:" {
			contentStartIdx = i + 1
			break
		}
	}

	if contentStartIdx > 0 && contentStartIdx < len(lines) {
		content = strings.TrimSpace(strings.Join(lines[contentStartIdx:], "\n"))
	}

	if title == "" {
		if u, err := url.Parse(extractSourceURL(raw)); err == nil {
			title = u.Host
		}
	}

	return title, content
}

func extractSourceURL(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "URL Source:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "URL Source:"))
		}
	}
	return ""
}

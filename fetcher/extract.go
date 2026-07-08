package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type httpStaticFetcher struct {
	client *http.Client
}

func newHTTPStaticFetcher(timeout time.Duration) *httpStaticFetcher {
	return &httpStaticFetcher{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (h *httpStaticFetcher) fetch(ctx context.Context, rawURL string, format Format) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("local: creating request: %w", err)
	}

	req.Header.Set("User-Agent", "outrider/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("local: fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local: HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("local: reading response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")

	// Non-HTML content: return as-is.
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		parsedURL, _ := url.Parse(rawURL)
		title := ""
		if parsedURL != nil {
			title = parsedURL.Host
		}
		return &FetchResult{
			Content: string(body),
			URL:     rawURL,
			Title:   title,
			Format:  format,
		}, nil
	}

	return extractContent(body, rawURL, format)
}

// extractContent is the shared pipeline: readability extracts the article,
// then html-to-markdown converts it to Markdown.
func extractContent(htmlBody []byte, rawURL string, format Format) (*FetchResult, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	prepared, err := preprocessHTML(htmlBody, format)
	if err != nil {
		return nil, fmt.Errorf("preprocessing HTML: %w", err)
	}

	article, err := readability.FromReader(strings.NewReader(string(prepared)), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("readability: %w", err)
	}

	// Render the article's DOM node to HTML, then convert to Markdown.
	var htmlBuf bytes.Buffer
	if err := article.RenderHTML(&htmlBuf); err != nil {
		return nil, fmt.Errorf("rendering article HTML: %w", err)
	}

	markdown, err := md.ConvertString(htmlBuf.String())
	if err != nil {
		return nil, fmt.Errorf("markdown conversion: %w", err)
	}

	if format == FormatLean {
		markdown = collapseWhitespace(markdown)
	}

	return &FetchResult{
		Content: strings.TrimSpace(markdown),
		URL:     rawURL,
		Title:   article.Title(),
		Format:  format,
	}, nil
}

// preprocessHTML removes noise and, for lean mode, strips links and images.
func preprocessHTML(htmlBody []byte, format Format) ([]byte, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	// Always drop non-content elements.
	doc.Find("script, style, noscript, iframe, form").Remove()

	if format == FormatLean {
		doc.Find("nav, header, footer, aside, [role='navigation'], [role='banner'], [role='contentinfo'], [role='complementary'], .sidebar, .nav, .menu, .breadcrumb, .related, #related").Remove()

		// Replace links with their text content.
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			s.ReplaceWithNodes(&html.Node{Type: html.TextNode, Data: text})
		})

		// Replace images with a compact alt marker.
		doc.Find("img, picture").Each(func(_ int, s *goquery.Selection) {
			alt, _ := s.Attr("alt")
			alt = strings.TrimSpace(alt)
			replacement := ""
			if alt != "" {
				replacement = fmt.Sprintf("[image: %s]", alt)
			}
			s.ReplaceWithNodes(&html.Node{Type: html.TextNode, Data: replacement})
		})
	}

	html, err := doc.Html()
	if err != nil {
		return nil, err
	}
	return []byte(html), nil
}

var multipleBlankLines = regexp.MustCompile(`\n{3,}`)
var trailingSpaces = regexp.MustCompile(`[^\S\r\n]+$`)

func collapseWhitespace(s string) string {
	s = trailingSpaces.ReplaceAllString(s, "")
	s = multipleBlankLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/fetcher"
)

// WebFetchInput defines the parameters for the web_fetch tool.
type WebFetchInput struct {
	URL       string `json:"url" jsonschema:"The URL to fetch content from"`
	Format    string `json:"format,omitempty" jsonschema:"Output format. 'lean' (default): strips link URLs and images for minimal tokens. 'markdown': preserves full links and images."`
	MaxLength int    `json:"maxLength,omitempty" jsonschema:"Max content length in characters (default 10000, max 50000)"`
}

// WebFetchOutput defines the output for the web_fetch tool.
type WebFetchOutput struct {
	Content string `json:"content" jsonschema:"The fetched page content as clean Markdown"`
}

type webFetchHandler struct {
	fetcher *fetcher.Fetcher
}

func (h *webFetchHandler) handle(ctx context.Context, req *mcp.CallToolRequest, input WebFetchInput) (*mcp.CallToolResult, WebFetchOutput, error) {
	if input.URL == "" {
		result := &mcp.CallToolResult{}
		result.SetError(fmt.Errorf("URL cannot be empty"))
		return result, WebFetchOutput{}, nil
	}

	format := fetcher.ParseFormat(input.Format)

	maxLength := input.MaxLength
	if maxLength <= 0 {
		maxLength = 10000
	}
	if maxLength > 50000 {
		maxLength = 50000
	}

	fetchResult, err := h.fetcher.FetchURL(ctx, input.URL, maxLength, string(format))
	if err != nil {
		result := &mcp.CallToolResult{}
		result.SetError(fmt.Errorf("fetching URL: %w", err))
		return result, WebFetchOutput{}, nil
	}

	content := formatContent(fetchResult)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}, WebFetchOutput{Content: content}, nil
}

func formatContent(result *fetcher.FetchResult) string {
	content := strings.TrimSpace(result.Content)
	if content == "" {
		content = result.Content
	}
	if strings.HasPrefix(content, "# ") {
		return fmt.Sprintf("URL: %s\n\n%s", result.URL, content)
	}
	return fmt.Sprintf("# %s\n\nURL: %s\n\n%s", result.Title, result.URL, content)
}

// RegisterWebFetch adds the web_fetch tool to an MCP server.
func RegisterWebFetch(server *mcp.Server, f *fetcher.Fetcher) {
	h := &webFetchHandler{fetcher: f}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_fetch",
		Description: "Fetch a web page URL and return its content as clean Markdown. Tries the Jina Reader API first, then a headless browser (if enabled), then local readability extraction. Use format 'lean' (default) to strip link URLs and images for minimal tokens, or 'markdown' to preserve them.",
	}, h.handle)
}

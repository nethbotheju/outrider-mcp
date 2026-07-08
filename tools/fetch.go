package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/fetcher"
)

// FetchInput defines the parameters for the fetch tool.
type FetchInput struct {
	URL       string `json:"url" jsonschema:"The URL to fetch content from"`
	MaxLength int    `json:"maxLength,omitempty" jsonschema:"Maximum content length in characters (default 50000)"`
}

// FetchOutput defines the output for the fetch tool.
type FetchOutput struct {
	Content string `json:"content" jsonschema:"The fetched page content as clean Markdown"`
}

// fetchHandler holds the dependencies for the fetch tool.
type fetchHandler struct {
	fetcher *fetcher.Fetcher
}

// handle executes the fetch tool logic.
func (h *fetchHandler) handle(ctx context.Context, req *mcp.CallToolRequest, input FetchInput) (*mcp.CallToolResult, FetchOutput, error) {
	if input.URL == "" {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: URL cannot be empty"},
			},
		}
		result.SetError(fmt.Errorf("URL cannot be empty"))
		return result, FetchOutput{}, nil
	}

	if input.MaxLength <= 0 {
		input.MaxLength = 50000
	}

	fetchResult, err := h.fetcher.FetchURL(ctx, input.URL, input.MaxLength)
	if err != nil {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error fetching URL: %v", err)},
			},
		}
		result.SetError(err)
		return result, FetchOutput{}, nil
	}

	content := fmt.Sprintf("Title: %s\nURL: %s\n\n%s", fetchResult.Title, fetchResult.URL, fetchResult.Content)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}, FetchOutput{Content: content}, nil
}

// RegisterFetch adds the fetch tool to an MCP server.
func RegisterFetch(server *mcp.Server, f *fetcher.Fetcher) {
	h := &fetchHandler{fetcher: f}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fetch",
		Description: "Fetch the content of a web page URL and return it as clean Markdown. Uses Jina Reader API (primary), headless Chrome (fallback), or static HTTP + readability (last resort). Useful for reading the full content of a page found via web_search.",
	}, h.handle)
}

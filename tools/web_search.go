package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/search"
)

// WebSearchInput defines the parameters for the web_search tool.
type WebSearchInput struct {
	Query string `json:"query" jsonschema:"required,The search query string"`
	Count int    `json:"count,omitempty" jsonschema:"Number of results to return (default 10, max 20)"`
}

// WebSearchOutput defines the output for the web_search tool.
type WebSearchOutput struct {
	Results string `json:"results" jsonschema:"Formatted search results with titles, URLs, and descriptions"`
}

// webSearchHandler holds the dependencies for the web_search tool.
type webSearchHandler struct {
	provider search.Provider
}

// handle executes the web_search tool logic.
func (h *webSearchHandler) handle(ctx context.Context, req *mcp.CallToolRequest, input WebSearchInput) (*mcp.CallToolResult, WebSearchOutput, error) {
	if input.Query == "" {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: search query cannot be empty"},
			},
		}
		result.SetError(fmt.Errorf("search query cannot be empty"))
		return result, WebSearchOutput{}, nil
	}

	results, err := h.provider.Search(ctx, input.Query, input.Count)
	if err != nil {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error searching the web: %v", err)},
			},
		}
		result.SetError(err)
		return result, WebSearchOutput{}, nil
	}

	formatted := search.FormatResults(results)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatted},
		},
	}, WebSearchOutput{Results: formatted}, nil
}

// RegisterWebSearch adds the web_search tool to an MCP server.
func RegisterWebSearch(server *mcp.Server, provider search.Provider) {
	h := &webSearchHandler{provider: provider}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Search the web using " + provider.Name() + ". Returns a list of sources with titles, URLs, and descriptions.",
	}, h.handle)
}

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/search"
)

// WebSearchInput is the parameter schema for providers without category support.
type WebSearchInput struct {
	Query string `json:"query" jsonschema:"The search query"`
	Count int    `json:"count,omitempty" jsonschema:"Number of results to return (default 10, max 20)"`
}

func (in WebSearchInput) query() string    { return in.Query }
func (in WebSearchInput) count() int       { return in.Count }
func (in WebSearchInput) category() string { return "" }

// SearXNGWebSearchInput adds the category parameter used by SearXNG.
type SearXNGWebSearchInput struct {
	Query    string `json:"query" jsonschema:"The search query"`
	Count    int    `json:"count,omitempty" jsonschema:"Number of results to return (default 5, max 10)"`
	Category string `json:"category,omitempty" jsonschema:"Search category. 'general' (default): broad web search. 'science': academic/scientific sources. 'news': current events. 'it': developer and technical Q&A."`
}

func (in SearXNGWebSearchInput) query() string    { return in.Query }
func (in SearXNGWebSearchInput) count() int       { return in.Count }
func (in SearXNGWebSearchInput) category() string { return in.Category }

type webSearchInput interface {
	query() string
	count() int
	category() string
}

type webSearchHandler struct {
	provider search.Provider
}

func (h *webSearchHandler) handle(ctx context.Context, req *mcp.CallToolRequest, input webSearchInput) (*mcp.CallToolResult, WebSearchOutput, error) {
	if input.query() == "" {
		result := &mcp.CallToolResult{}
		result.SetError(fmt.Errorf("search query cannot be empty"))
		return result, WebSearchOutput{}, nil
	}

	count := input.count()
	if count <= 0 {
		count = h.provider.DefaultCount()
	}
	if count > h.provider.MaxCount() {
		count = h.provider.MaxCount()
	}

	results, err := h.provider.Search(ctx, input.query(), count, input.category())
	if err != nil {
		result := &mcp.CallToolResult{}
		result.SetError(fmt.Errorf("searching the web: %w", err))
		return result, WebSearchOutput{}, nil
	}

	formatted := search.FormatResults(results)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: formatted},
		},
	}, WebSearchOutput{Results: formatted}, nil
}

// WebSearchOutput defines the output for the web_search tool.
type WebSearchOutput struct {
	Results string `json:"results" jsonschema:"Formatted search results with titles, URLs, and descriptions"`
}

// RegisterWebSearch adds the web_search tool to an MCP server.
func RegisterWebSearch(server *mcp.Server, provider search.Provider) {
	h := &webSearchHandler{provider: provider}

	if provider.SupportsCategories() {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "web_search",
			Description: "Search the web using " + provider.Name() + ". Returns a numbered list of sources with titles, URLs, descriptions, and optional metadata (published date, source, journal, DOI, PDF link, authors).",
		}, func(ctx context.Context, req *mcp.CallToolRequest, input SearXNGWebSearchInput) (*mcp.CallToolResult, WebSearchOutput, error) {
			return h.handle(ctx, req, input)
		})
		return
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Search the web using " + provider.Name() + ". Returns a numbered list of sources with titles, URLs, and descriptions.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input WebSearchInput) (*mcp.CallToolResult, WebSearchOutput, error) {
		return h.handle(ctx, req, input)
	})
}

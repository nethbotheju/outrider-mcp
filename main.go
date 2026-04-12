package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/web-search-mcp/fetcher"
	"github.com/nethbotheju/web-search-mcp/search"
)

type WebSearchInput struct {
	Query string `json:"query" jsonschema:"required,The search query string"`
	Count int    `json:"count,omitempty" jsonschema:"Number of results to return (default 10, max 20)"`
}

type WebSearchOutput struct {
	Results string `json:"results" jsonschema:"Formatted search results with titles, URLs, and descriptions"`
}

type FetchInput struct {
	URL       string `json:"url" jsonschema:"required,The URL to fetch content from"`
	MaxLength int    `json:"maxLength,omitempty" jsonschema:"Maximum content length in characters (default 50000)"`
}

type FetchOutput struct {
	Content string `json:"content" jsonschema:"The fetched page content as clean text"`
}

var (
	braveClient *search.BraveClient
	httpFetcher *fetcher.Fetcher
)

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("BRAVE_SEARCH_API_KEY")
	braveClient = search.NewBraveClient(apiKey)
	httpFetcher = fetcher.NewFetcher()

	log.Println("Starting web-search-mcp server...")
	if apiKey == "" {
		log.Println("WARNING: BRAVE_SEARCH_API_KEY is not set. web_search tool will not work.")
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "web-search-mcp",
			Version: "1.0.0",
		},
		nil,
	)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Search the web using Brave Search. Returns a list of sources with titles, URLs, and descriptions.",
	}, handleWebSearch)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fetch",
		Description: "Fetch the content of a web page URL and return it as clean text. Useful for reading the full content of a page found via web_search.",
	}, handleFetch)

	// stdout is reserved for JSON-RPC on stdio transport -- all logging must go to stderr.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleWebSearch(ctx context.Context, req *mcp.CallToolRequest, input WebSearchInput) (*mcp.CallToolResult, WebSearchOutput, error) {
	if input.Query == "" {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: search query cannot be empty"},
			},
		}
		result.SetError(fmt.Errorf("search query cannot be empty"))
		return result, WebSearchOutput{}, nil
	}

	if input.Count <= 0 {
		input.Count = 10
	}

	results, err := braveClient.Search(ctx, input.Query, input.Count)
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

func handleFetch(ctx context.Context, req *mcp.CallToolRequest, input FetchInput) (*mcp.CallToolResult, FetchOutput, error) {
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

	fetchResult, err := httpFetcher.FetchURL(ctx, input.URL, input.MaxLength)
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

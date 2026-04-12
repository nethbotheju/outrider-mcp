package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/web-search-mcp/fetcher"
	"github.com/nethbotheju/web-search-mcp/search"
	"github.com/nethbotheju/web-search-mcp/tools"
)

func main() {
	provider := search.NewDuckDuckGoProvider()
	fetcher := fetcher.NewFetcher()

	log.Println("Starting web-search-mcp server...")
	log.Printf("Search provider: %s", provider.Name())

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "web-search-mcp",
			Version: "1.0.0",
		},
		nil,
	)

	tools.RegisterWebSearch(server, provider)
	tools.RegisterFetch(server, fetcher)

	// stdout is reserved for JSON-RPC on stdio transport -- all logging must go to stderr.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

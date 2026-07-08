package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/answer"
	"github.com/nethbotheju/outrider-mcp/fetcher"
	"github.com/nethbotheju/outrider-mcp/search"
	"github.com/nethbotheju/outrider-mcp/tools"
)

func main() {
	provider := search.NewProviderFromEnv()
	fetcher := fetcher.NewFetcher()

	log.Println("Starting outrider server...")
	log.Printf("Search provider: %s", provider.Name())

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "outrider",
			Version: "1.0.0",
		},
		nil,
	)

	tools.RegisterWebSearch(server, provider)
	tools.RegisterFetch(server, fetcher)

	if apiKey := os.Getenv("ANSWER_LLM_API_KEY"); apiKey != "" {
		agent := answer.NewAgent(provider, fetcher)
		tools.RegisterAnswer(server, agent)
		log.Println("Answer tool: enabled")
	} else {
		log.Println("Answer tool: disabled (set ANSWER_LLM_API_KEY to enable)")
	}

	// stdout is reserved for JSON-RPC on stdio transport -- all logging must go to stderr.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

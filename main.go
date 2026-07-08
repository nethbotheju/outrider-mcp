package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/answer"
	"github.com/nethbotheju/outrider-mcp/config"
	"github.com/nethbotheju/outrider-mcp/fetcher"
	"github.com/nethbotheju/outrider-mcp/search"
	"github.com/nethbotheju/outrider-mcp/tools"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	provider := search.NewProviderFromConfig(cfg)
	f := fetcher.NewFetcher(cfg.Fetch)

	log.Println("Starting outrider server...")
	log.Printf("Search provider: %s", provider.Name())

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "outrider",
			Version: "1.0.0",
		},
		nil,
	)

	registered := registerTools(server, cfg, provider, f)
	if !registered {
		log.Println("Warning: all tools were disabled; enabling web_search as a fallback")
		cfg.Tools.WebSearch.Enabled = true
		registerTools(server, cfg, provider, f)
	}

	// stdout is reserved for JSON-RPC on stdio transport -- all logging must go to stderr.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func registerTools(server *mcp.Server, cfg *config.Config, provider search.Provider, f *fetcher.Fetcher) bool {
	registered := false

	if cfg.Tools.WebSearch.Enabled {
		tools.RegisterWebSearch(server, provider)
		log.Println("Tool registered: web_search")
		registered = true
	}

	if cfg.Tools.WebFetch.Enabled {
		tools.RegisterWebFetch(server, f)
		log.Println("Tool registered: web_fetch")
		registered = true
	}

	if cfg.Tools.WebAnswer.Enabled && cfg.Answer.Enabled && cfg.Answer.Model != "" && cfg.Answer.BaseURL != "" {
		agent := answer.NewAgent(provider, f, cfg.Answer)
		tools.RegisterWebAnswer(server, agent)
		log.Println("Tool registered: web_answer")
		registered = true
	} else if cfg.Tools.WebAnswer.Enabled {
		log.Println("Tool web_answer is enabled but not configured (missing model or baseUrl); skipping")
	}

	return registered
}

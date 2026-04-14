package answer_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nethbotheju/web-search-mcp/answer"
	"github.com/nethbotheju/web-search-mcp/fetcher"
	"github.com/nethbotheju/web-search-mcp/search"
)

func TestAnswerLive(t *testing.T) {
	apiKey := os.Getenv("ANSWER_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("ANSWER_LLM_API_KEY not set, skipping live answer test")
	}

	provider := search.NewDuckDuckGoProvider()
	f := fetcher.NewFetcher()
	agent := answer.NewAgent(provider, f)

	result, err := agent.Run(context.Background(), "What is the current version of Go programming language?")
	if err != nil {
		t.Fatalf("Agent.Run() returned error: %v", err)
	}

	if result == "" {
		t.Fatal("Agent.Run() returned empty result")
	}

	t.Logf("Answer:\n%s", result)
}

func TestAnswerFactualQuery(t *testing.T) {
	apiKey := os.Getenv("ANSWER_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("ANSWER_LLM_API_KEY not set, skipping live answer test")
	}

	provider := search.NewDuckDuckGoProvider()
	f := fetcher.NewFetcher()
	agent := answer.NewAgent(provider, f)

	result, err := agent.Run(context.Background(), "Who created the Go programming language?")
	if err != nil {
		t.Fatalf("Agent.Run() returned error: %v", err)
	}

	if result == "" {
		t.Fatal("Agent.Run() returned empty result")
	}

	lower := strings.ToLower(result)
	if !strings.Contains(lower, "google") {
		t.Errorf("Expected answer to mention Google, got: %s", result)
	}

	t.Logf("Answer:\n%s", result)
}

func TestAnswerEmptyQuestion(t *testing.T) {
	apiKey := os.Getenv("ANSWER_LLM_API_KEY")
	if apiKey == "" {
		t.Skip("ANSWER_LLM_API_KEY not set, skipping live answer test")
	}

	provider := search.NewDuckDuckGoProvider()
	f := fetcher.NewFetcher()
	agent := answer.NewAgent(provider, f)

	_, err := agent.Run(context.Background(), "")
	if err == nil {
		t.Fatal("Expected error for empty question, got nil")
	}

	t.Logf("Error for empty question (expected): %v", err)
}

func TestAnswerInvalidAPIKey(t *testing.T) {
	originalKey := os.Getenv("ANSWER_LLM_API_KEY")
	os.Setenv("ANSWER_LLM_API_KEY", "invalid-key-12345")
	defer os.Setenv("ANSWER_LLM_API_KEY", originalKey)

	provider := search.NewDuckDuckGoProvider()
	f := fetcher.NewFetcher()
	agent := answer.NewAgent(provider, f)

	_, err := agent.Run(context.Background(), "test query")
	if err == nil {
		t.Fatal("Expected error for invalid API key, got nil")
	}

	t.Logf("Error for invalid API key (expected): %v", err)
}

package answer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nethbotheju/outrider-mcp/config"
	"github.com/nethbotheju/outrider-mcp/fetcher"
	"github.com/nethbotheju/outrider-mcp/search"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

const maxFetchChars = 10000

type execCounters struct {
	searches int
	fetches  int
}

// Agent answers questions using a small side LLM and the search/fetch tools.
type Agent struct {
	client   openai.Client
	model    string
	provider search.Provider
	fetcher  *fetcher.Fetcher
	cfg      config.AnswerConfig
}

// NewAgent creates an Agent from configuration.
func NewAgent(provider search.Provider, f *fetcher.Fetcher, cfg config.AnswerConfig) *Agent {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	}

	return &Agent{
		client:   openai.NewClient(opts...),
		model:    cfg.Model,
		provider: provider,
		fetcher:  f,
		cfg:      cfg,
	}
}

// Run answers a question. If url is non-empty, it answers from that page directly;
// otherwise it runs a bounded search-and-fetch loop.
func (a *Agent) Run(ctx context.Context, question, url string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "", fmt.Errorf("question cannot be empty")
	}

	if strings.TrimSpace(url) != "" {
		return a.answerFromURL(ctx, question, url)
	}
	return a.answerViaSearch(ctx, question)
}

func (a *Agent) answerFromURL(ctx context.Context, question, url string) (string, error) {
	log.Printf("[answer] url mode: fetching %s", url)
	result, err := a.fetcher.FetchURL(ctx, url, maxFetchChars, string(fetcher.FormatLean))
	if err != nil {
		return "", fmt.Errorf("fetching URL: %w", err)
	}

	context := fmt.Sprintf("Title: %s\nURL: %s\n\n%s\n\nQuestion: %s", result.Title, result.URL, result.Content, question)
	resp, err := a.chatCompletion(ctx, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(urlSystemPrompt),
		openai.UserMessage(context),
	}, false)
	if err != nil {
		return "", err
	}

	return stripThoughtTags(resp.Content), nil
}

func (a *Agent) answerViaSearch(ctx context.Context, question string) (string, error) {
	log.Printf("[answer] search mode: starting bounded loop (maxTurns=%d)", a.cfg.MaxTurns)
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(searchSystemPrompt),
		openai.UserMessage(question),
	}

	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "web_search",
			Description: openai.String("Search the web. Returns numbered titles, URLs, and short descriptions. Use this to discover relevant pages."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search keywords, optimized for relevance",
					},
					"count": map[string]any{
						"type":        "integer",
						"description": "Number of results (default 5, max 10)",
					},
				},
				"required": []string{"query"},
			},
		}),
		openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "web_fetch",
			Description: openai.String("Fetch the full content of a specific URL as clean markdown. Use this to read a page found via web_search that is likely to contain the answer."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "The URL to read",
					},
				},
				"required": []string{"url"},
			},
		}),
	}

	counters := &execCounters{}
	turns := 0
	for turns < a.cfg.MaxTurns {
		turns++
		log.Printf("[answer] turn %d: sending request to LLM", turns)

		resp, err := a.chatCompletion(ctx, messages, true, tools...)
		if err != nil {
			return "", fmt.Errorf("LLM API error (turn %d): %w", turns, err)
		}

		if len(resp.ToolCalls) == 0 {
			if strings.TrimSpace(resp.Content) != "" {
				log.Printf("[answer] turn %d: final answer received (%d chars)", turns, len(resp.Content))
				return stripThoughtTags(resp.Content), nil
			}
			break
		}

		log.Printf("[answer] turn %d: LLM requested %d tool call(s)", turns, len(resp.ToolCalls))
		assistantMsg := resp.ToAssistantMessageParam()
		assistantMsg.ToolCalls = make([]openai.ChatCompletionMessageToolCallUnionParam, len(resp.ToolCalls))
		for i, tc := range resp.ToolCalls {
			assistantMsg.ToolCalls[i] = tc.ToParam()
		}
		messages = append(messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantMsg})

		for _, tc := range resp.ToolCalls {
			log.Printf("[answer]   tool_call: %s(%s)", tc.Function.Name, tc.Function.Arguments)
			result, err := a.executeToolCall(ctx, tc, counters)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %v", err)
				log.Printf("[answer]   tool_error: %v", err)
			}
			messages = append(messages, openai.ToolMessage(result, tc.ID))
		}
	}

	closer := forceAnswerPrompt
	if turns >= a.cfg.MaxTurns {
		closer += " (turn limit reached)"
	}
	log.Printf("[answer] forcing final answer after %d turns", turns)
	resp, err := a.chatCompletion(ctx, append(messages, openai.UserMessage(closer)), false)
	if err != nil {
		return "", fmt.Errorf("LLM API error (final): %w", err)
	}
	return stripThoughtTags(resp.Content), nil
}

func (a *Agent) chatCompletion(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	allowTools bool,
	tools ...openai.ChatCompletionToolUnionParam,
) (openai.ChatCompletionMessage, error) {
	params := openai.ChatCompletionNewParams{
		Model:       a.model,
		Messages:    messages,
		Temperature: openai.Float(a.cfg.Temperature),
	}

	if allowTools && len(tools) > 0 {
		params.Tools = tools
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.String("auto"),
		}
	} else if a.cfg.MaxTokens > 0 {
		params.MaxCompletionTokens = openai.Int(int64(a.cfg.MaxTokens))
	}

	resp, err := a.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	if len(resp.Choices) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("LLM returned no choices")
	}
	return resp.Choices[0].Message, nil
}

func (a *Agent) executeToolCall(ctx context.Context, tc openai.ChatCompletionMessageToolCallUnion, counters *execCounters) (string, error) {
	fn := tc.Function

	switch fn.Name {
	case "web_search":
		if counters.searches >= a.cfg.MaxSearches {
			return "Search limit reached. Answer using the results you already have.", nil
		}
		counters.searches++

		var args struct {
			Query string `json:"query"`
			Count int    `json:"count"`
		}
		if err := json.Unmarshal([]byte(fn.Arguments), &args); err != nil {
			return "", fmt.Errorf("invalid web_search arguments: %w", err)
		}
		if strings.TrimSpace(args.Query) == "" {
			return "Error: empty search query.", nil
		}

		category := ""
		if a.provider.SupportsCategories() {
			category = "general"
		}
		results, err := a.provider.Search(ctx, args.Query, args.Count, category)
		if err != nil {
			return "", fmt.Errorf("search failed: %w", err)
		}
		if len(results) == 0 {
			return "No results found. Try a different query, or answer that the information was not found.", nil
		}
		return search.FormatResults(results), nil

	case "web_fetch":
		if counters.fetches >= a.cfg.MaxFetches {
			return "Fetch limit reached. Answer using the information you already have.", nil
		}
		counters.fetches++

		var args struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(fn.Arguments), &args); err != nil {
			return "", fmt.Errorf("invalid web_fetch arguments: %w", err)
		}
		if strings.TrimSpace(args.URL) == "" {
			return "Error: empty url.", nil
		}

		result, err := a.fetcher.FetchURL(ctx, args.URL, maxFetchChars, string(fetcher.FormatLean))
		if err != nil {
			return "", fmt.Errorf("fetch failed: %w", err)
		}
		return fmt.Sprintf("Title: %s\nURL: %s\n\n%s", result.Title, result.URL, result.Content), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", fn.Name)
	}
}

// stripThoughtTags removes <thought>...</thought> blocks that some models emit.
func stripThoughtTags(s string) string {
	for {
		start := strings.Index(s, "<thought>")
		if start == -1 {
			break
		}
		end := strings.Index(s, "</thought>")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("</thought>"):]
	}
	return strings.TrimSpace(s)
}

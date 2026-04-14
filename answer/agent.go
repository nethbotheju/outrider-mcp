package answer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nethbotheju/web-search-mcp/fetcher"
	"github.com/nethbotheju/web-search-mcp/search"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"
	defaultModel   = "gemini-2.5-flash-lite"
	maxIterations  = 5
	maxFetchChars  = 10000
)

type Agent struct {
	client   openai.Client
	model    string
	provider search.Provider
	fetcher  *fetcher.Fetcher
}

func NewAgent(provider search.Provider, f *fetcher.Fetcher) *Agent {
	apiKey := os.Getenv("ANSWER_LLM_API_KEY")
	baseURL := os.Getenv("ANSWER_LLM_BASE_URL")
	model := os.Getenv("ANSWER_LLM_MODEL")

	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if model == "" {
		model = defaultModel
	}

	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	}

	return &Agent{
		client:   openai.NewClient(opts...),
		model:    model,
		provider: provider,
		fetcher:  f,
	}
}

func (a *Agent) Run(ctx context.Context, question string) (string, error) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(question),
	}

	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "web_search",
			Description: openai.String("Search the web for information. Returns a list of results with titles, URLs, and descriptions."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query string",
					},
				},
				"required": []string{"query"},
			},
		}),
		openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
			Name:        "fetch",
			Description: openai.String("Fetch the full content of a web page by URL. Returns the page content as clean text."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "The URL to fetch content from",
					},
				},
				"required": []string{"url"},
			},
		}),
	}

	for i := range maxIterations {
		resp, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return "", fmt.Errorf("LLM API error (iteration %d): %w", i+1, err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("LLM returned no choices (iteration %d)", i+1)
		}

		choice := resp.Choices[0]

		if choice.FinishReason == "tool_calls" || len(choice.Message.ToolCalls) > 0 {
			assistantMsg := choice.Message.ToAssistantMessageParam()
			assistantMsg.ToolCalls = make([]openai.ChatCompletionMessageToolCallUnionParam, len(choice.Message.ToolCalls))
			for j, tc := range choice.Message.ToolCalls {
				assistantMsg.ToolCalls[j] = tc.ToParam()
			}
			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &assistantMsg,
			})

			for _, tc := range choice.Message.ToolCalls {
				result, err := a.executeTool(ctx, tc)
				if err != nil {
					result = fmt.Sprintf("Error executing tool: %v", err)
				}
				messages = append(messages, openai.ToolMessage(result, tc.ID))
			}
			continue
		}

		return choice.Message.Content, nil
	}

	return "", fmt.Errorf("agent loop exceeded %d iterations without producing a final answer", maxIterations)
}

func (a *Agent) executeTool(ctx context.Context, tc openai.ChatCompletionMessageToolCallUnion) (string, error) {
	fn := tc.Function

	switch fn.Name {
	case "web_search":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(fn.Arguments), &args); err != nil {
			return "", fmt.Errorf("invalid web_search arguments: %w", err)
		}
		if args.Query == "" {
			return "Error: search query cannot be empty", nil
		}

		results, err := a.provider.Search(ctx, args.Query, 5)
		if err != nil {
			return "", fmt.Errorf("search failed: %w", err)
		}
		return search.FormatResults(results), nil

	case "fetch":
		var args struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(fn.Arguments), &args); err != nil {
			return "", fmt.Errorf("invalid fetch arguments: %w", err)
		}
		if args.URL == "" {
			return "Error: URL cannot be empty", nil
		}

		result, err := a.fetcher.FetchURL(ctx, args.URL, maxFetchChars)
		if err != nil {
			return "", fmt.Errorf("fetch failed: %w", err)
		}
		return fmt.Sprintf("Title: %s\nURL: %s\n\n%s", result.Title, result.URL, result.Content), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", fn.Name)
	}
}

const systemPrompt = `You are a research assistant with access to web search and page fetching tools. Your job is to find accurate, up-to-date information and return concise, factual answers.

Guidelines:
- Use web_search to find relevant information. Formulate concise, effective search queries.
- Use fetch to read the full content of relevant pages when search snippets don't provide enough detail.
- If search results already contain the answer, you may respond directly without fetching additional pages.
- Provide concise answers. Include source URLs when possible.
- Only use information from the tool results. Do not hallucinate or make up information.`

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/outrider-mcp/answer"
)

// WebAnswerInput defines the parameters for the web_answer tool.
type WebAnswerInput struct {
	Question string `json:"question" jsonschema:"The question to answer by searching the web"`
	URL      string `json:"url,omitempty" jsonschema:"Optional. A specific URL to read and answer from. Omit to search the web."`
}

// WebAnswerOutput defines the output for the web_answer tool.
type WebAnswerOutput struct {
	Answer string `json:"answer" jsonschema:"A concise answer to the question with source URLs"`
}

type webAnswerHandler struct {
	agent *answer.Agent
}

func (h *webAnswerHandler) handle(ctx context.Context, req *mcp.CallToolRequest, input WebAnswerInput) (*mcp.CallToolResult, WebAnswerOutput, error) {
	if input.Question == "" {
		result := &mcp.CallToolResult{}
		result.SetError(fmt.Errorf("question cannot be empty"))
		return result, WebAnswerOutput{}, nil
	}

	ans, err := h.agent.Run(ctx, input.Question, input.URL)
	if err != nil {
		result := &mcp.CallToolResult{}
		result.SetError(fmt.Errorf("getting answer: %w", err))
		return result, WebAnswerOutput{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: ans},
		},
	}, WebAnswerOutput{Answer: ans}, nil
}

// RegisterWebAnswer adds the web_answer tool to an MCP server.
func RegisterWebAnswer(server *mcp.Server, agent *answer.Agent) {
	h := &webAnswerHandler{agent: agent}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_answer",
		Description: "Answer a question from the web without flooding your context with full page content. Pass a 'question' and optionally a 'url'. If a URL is given, reads that page and answers. If no URL, searches the web and reads only the most relevant pages, then answers. Returns a concise answer plus a sources list.",
	}, h.handle)
}

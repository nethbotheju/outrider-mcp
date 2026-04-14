package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nethbotheju/web-search-mcp/answer"
)

type AnswerInput struct {
	Question string `json:"question" jsonschema:"required,The question to answer by searching the web"`
}

type AnswerOutput struct {
	Answer string `json:"answer" jsonschema:"A concise answer to the question with source URLs"`
}

type answerHandler struct {
	agent *answer.Agent
}

func (h *answerHandler) handle(ctx context.Context, req *mcp.CallToolRequest, input AnswerInput) (*mcp.CallToolResult, AnswerOutput, error) {
	if input.Question == "" {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: question cannot be empty"},
			},
		}
		result.SetError(fmt.Errorf("question cannot be empty"))
		return result, AnswerOutput{}, nil
	}

	answer, err := h.agent.Run(ctx, input.Question)
	if err != nil {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error getting answer: %v", err)},
			},
		}
		result.SetError(err)
		return result, AnswerOutput{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: answer},
		},
	}, AnswerOutput{Answer: answer}, nil
}

func RegisterAnswer(server *mcp.Server, agent *answer.Agent) {
	h := &answerHandler{agent: agent}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "answer",
		Description: "Answer a question by searching the web and reading relevant pages. Returns a concise answer with sources. Use this instead of web_search + fetch when you need a direct answer to a factual question.",
	}, h.handle)
}

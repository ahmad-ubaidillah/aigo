package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type GeneralHandler struct {
	client llm.LLMClient
	model  string
}

func NewGeneralHandler(client llm.LLMClient, model string) *GeneralHandler {
	return &GeneralHandler{
		client: client,
		model:  model,
	}
}

func (h *GeneralHandler) CanHandle(intent string) bool {
	return intent == types.IntentGeneral
}

func (h *GeneralHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	// If no LLM client configured, return the task description
	if h.client == nil {
		return &types.ToolResult{
			Success: true,
			Output:  fmt.Sprintf("I received your message: %s\n\n(No LLM configured - please set up GLM or OpenAI)", task.Description),
		}, nil
	}

	// Call LLM to answer the question
	prompt := fmt.Sprintf("You are Aigo, a helpful AI assistant. Answer the following question concisely and accurately:\n\n%s", task.Description)
	
	response, err := h.client.Complete(ctx, prompt)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("LLM error: %v", err),
		}, nil
	}

	// Validate response
	if response == nil {
		return &types.ToolResult{
			Success: false,
			Error:   "LLM returned nil response",
		}, nil
	}

	// Check for empty content
	content := strings.TrimSpace(response.Content)
	if content == "" {
		return &types.ToolResult{
			Success: false,
			Error:   "LLM returned empty response",
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  content,
	}, nil
}

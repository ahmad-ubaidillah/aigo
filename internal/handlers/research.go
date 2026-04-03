package handlers

import (
	"context"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/research"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type ResearchHandler struct {
	client *research.SearchClient
}

func NewResearchHandler(client *research.SearchClient) *ResearchHandler {
	return &ResearchHandler{client: client}
}

func (h *ResearchHandler) CanHandle(intent string) bool {
	return intent == types.IntentResearch
}

func (h *ResearchHandler) Execute(ctx context.Context, task *types.Task, _ string) (*types.ToolResult, error) {
	desc := strings.TrimSpace(task.Description)

	if strings.HasPrefix(desc, "web ") {
		query := strings.TrimPrefix(desc, "web ")
		return h.webSearch(ctx, query)
	}

	if strings.HasPrefix(desc, "code ") {
		spec := strings.TrimPrefix(desc, "code ")
		parts := strings.SplitN(spec, " ", 2)
		pattern := parts[0]
		language := ""
		if len(parts) > 1 {
			language = parts[1]
		}
		return h.codeSearch(ctx, pattern, language)
	}

	if strings.HasPrefix(desc, "docs ") {
		spec := strings.TrimPrefix(desc, "docs ")
		parts := strings.SplitN(spec, " ", 2)
		query := parts[0]
		docType := ""
		if len(parts) > 1 {
			docType = parts[1]
		}
		return h.docsSearch(ctx, query, docType)
	}

	return &types.ToolResult{
		Success: false,
		Error:   "unknown research command, use 'web <query>', 'code <pattern> [language]', or 'docs <query> [type]'",
	}, nil
}

func (h *ResearchHandler) webSearch(ctx context.Context, query string) (*types.ToolResult, error) {
	result, err := h.client.WebSearch(ctx, query, 5)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  result.Result,
	}, nil
}

func (h *ResearchHandler) codeSearch(ctx context.Context, pattern, language string) (*types.ToolResult, error) {
	result, err := h.client.CodeSearch(ctx, pattern, language)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  result.Result,
	}, nil
}

func (h *ResearchHandler) docsSearch(ctx context.Context, query, docType string) (*types.ToolResult, error) {
	result, err := h.client.DocsSearch(ctx, query, docType)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  result.Result,
	}, nil
}

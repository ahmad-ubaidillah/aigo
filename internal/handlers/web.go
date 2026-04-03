package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type WebHandler struct{}

func (h *WebHandler) CanHandle(intent string) bool {
	return intent == types.IntentWeb
}

func (h *WebHandler) Search(query string) (*types.ToolResult, error) {
	escaped := strings.ReplaceAll(query, " ", "+")
	cmd := exec.Command("curl", "-s", "https://html.duckduckgo.com/html/?q="+escaped)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("search query %q: %w", query, err)
	}

	text := stripHTML(string(output))
	if len(text) > 1000 {
		text = text[:1000]
	}

	return &types.ToolResult{
		Success: true,
		Output:  text,
	}, nil
}

func (h *WebHandler) Extract(url string) (*types.ToolResult, error) {
	cmd := exec.Command("curl", "-s", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("extract url %q: %w", url, err)
	}

	text := stripHTML(string(output))
	if len(text) > 2000 {
		text = text[:2000]
	}

	return &types.ToolResult{
		Success: true,
		Output:  text,
	}, nil
}

func (h *WebHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	desc := strings.TrimSpace(task.Description)

	if strings.HasPrefix(desc, "search ") {
		query := strings.TrimPrefix(desc, "search ")
		return h.Search(query)
	}

	if strings.HasPrefix(desc, "extract ") {
		url := strings.TrimPrefix(desc, "extract ")
		return h.Extract(url)
	}

	return &types.ToolResult{
		Success: false,
		Error:   "unknown web command, use 'search <query>' or 'extract <url>'",
	}, nil
}

func stripHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(input, " ")

	reSpace := regexp.MustCompile(`\s+`)
	text = reSpace.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

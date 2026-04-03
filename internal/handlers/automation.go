package handlers

import (
	"context"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type AutomationHandler struct{}

func (h *AutomationHandler) CanHandle(intent string) bool {
	return intent == types.IntentAutomation
}

func (h *AutomationHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	return &types.ToolResult{
		Success: true,
		Output:  "Automation handler — cron scheduling not yet implemented",
	}, nil
}

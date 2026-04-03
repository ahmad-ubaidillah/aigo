package handlers

import (
	"context"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type GeneralHandler struct{}

func (h *GeneralHandler) CanHandle(intent string) bool {
	return intent == types.IntentGeneral
}

func (h *GeneralHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	return &types.ToolResult{
		Success: true,
		Output:  task.Description,
	}, nil
}

package handlers

import (
	"context"
	"fmt"

	"github.com/ahmad-ubaidillah/aigo/internal/opencode"
	"github.com/ahmad-ubaidillah/aigo/internal/tools"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Installer interface {
	CheckOpenCode() (bool, string, error)
	InstallOpenCode(ctx context.Context, path string) error
}

type CodingHandler struct {
	ocClient  *opencode.Client
	registry  *tools.ToolRegistry
	installer Installer
}

func NewCodingHandler(registry *tools.ToolRegistry, installer Installer) *CodingHandler {
	return &CodingHandler{
		registry:  registry,
		installer: installer,
	}
}

func (h *CodingHandler) CanHandle(intent string) bool {
	return intent == types.IntentCoding
}

func (h *CodingHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	available, path, err := h.installer.CheckOpenCode()
	if err != nil {
		return h.nativeFallback(ctx, task, "OpenCode check failed: "+err.Error())
	}

	if !available {
		installErr := h.installer.InstallOpenCode(ctx, path)
		if installErr != nil {
			return h.nativeFallback(ctx, task, "OpenCode not available and install failed: "+installErr.Error())
		}
		h.ocClient, _ = opencode.NewClient(path, 300, workspace)
	}

	if h.ocClient != nil {
		health, _ := h.ocClient.HealthCheck()
		if !health.Success {
			return h.nativeFallback(ctx, task, "OpenCode health check failed: "+health.Error)
		}
	}

	if h.ocClient != nil {
		return h.ocClient.Run(ctx, task.Description, task.SessionID)
	}

	return h.nativeFallback(ctx, task, "OpenCode not available")
}

func (h *CodingHandler) nativeFallback(ctx context.Context, task *types.Task, reason string) (*types.ToolResult, error) {
	bashTool := h.registry.Get("bash")
	if bashTool == nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("coding task failed: %s. No fallback available.", reason),
		}, nil
	}

	result, err := bashTool.Execute(ctx, map[string]any{"command": task.Description})
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("coding task failed: %s. Native fallback also failed: %v", reason, err),
		}, nil
	}

	result.Output = fmt.Sprintf("[Native Fallback] %s\n\n%s", reason, result.Output)
	return result, nil
}

package tools

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type BashTool struct{}

func (t *BashTool) Name() string        { return "bash" }
func (t *BashTool) Description() string { return "Execute a shell command" }
func (t *BashTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Shell command to execute",
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Timeout in seconds (default 30)",
			},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	cmdVal, ok := params["command"]
	if !ok {
		return &types.ToolResult{Success: false, Error: "missing required parameter: command"}, nil
	}
	command, ok := cmdVal.(string)
	if !ok {
		return &types.ToolResult{Success: false, Error: "parameter command must be a string"}, nil
	}

	timeoutSec := 30
	if tVal, ok := params["timeout"]; ok {
		switch v := tVal.(type) {
		case int:
			timeoutSec = v
		case float64:
			timeoutSec = int(v)
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "sh", "-c", command)
	out, err := cmd.CombinedOutput()

	result := &types.ToolResult{
		Success: err == nil,
		Output:  strings.TrimSpace(string(out)),
	}
	if err != nil {
		result.Error = err.Error()
	}

	OutputTruncate(result, 102400)
	return result, nil
}

type TaskTool struct{}

func (t *TaskTool) Name() string        { return "task" }
func (t *TaskTool) Description() string { return "Spawn a subagent task" }
func (t *TaskTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"description": map[string]any{
				"type":        "string",
				"description": "Task description",
			},
			"category": map[string]any{
				"type":        "string",
				"description": "quick|unspecified-low|unspecified-high|writing",
			},
			"session_id": map[string]any{
				"type":        "string",
				"description": "Parent session ID",
			},
		},
		"required": []string{"description"},
	}
}

func (t *TaskTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	descVal, ok := params["description"]
	if !ok {
		return &types.ToolResult{Success: false, Error: "missing required parameter: description"}, nil
	}
	description, ok := descVal.(string)
	if !ok {
		return &types.ToolResult{Success: false, Error: "parameter description must be a string"}, nil
	}

	category := ""
	if c, ok := params["category"]; ok {
		if s, ok := c.(string); ok {
			category = s
		}
	}

	sessionID := ""
	if s, ok := params["session_id"]; ok {
		if str, ok := s.(string); ok {
			sessionID = str
		}
	}

	return &types.ToolResult{
		Success: true,
		Output:  "Task queued: " + description,
		Metadata: map[string]string{
			"category":   category,
			"session_id": sessionID,
		},
	}, nil
}

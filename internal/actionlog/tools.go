package actionlog

import (
	"context"
	"fmt"

	"github.com/hermes-v2/aigo/internal/tools"
)

type ActionLogTool struct {
	log *ActionLog
}

func NewActionLogTool(basePath string) (*ActionLogTool, error) {
	log, err := New(basePath, 1000)
	if err != nil {
		return nil, err
	}
	return &ActionLogTool{log: log}, nil
}

func (t *ActionLogTool) Name() string   { return "actionlog_list" }
func (t *ActionLogTool) Description() string {
	return "List recent actions"
}

func (t *ActionLogTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *ActionLogTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "actionlog_list",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{"type": "string", "description": "Project path"},
					"limit":   map[string]interface{}{"type": "number", "description": "Max actions"},
				},
			},
		},
	}
}

func (t *ActionLogTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	actions := t.log.List(project, limit)
	if len(actions) == 0 {
		return "No actions", nil
	}
	var result string
	for _, a := range actions {
		result += fmt.Sprintf("%s | %s | %s\n", a.ID[:8], a.Type, a.Description)
	}
	return result, nil
}

type ActionLogUndoTool struct {
	log *ActionLog
}

func NewActionLogUndoTool(basePath string) (*ActionLogUndoTool, error) {
	log, err := New(basePath, 1000)
	if err != nil {
		return nil, err
	}
	return &ActionLogUndoTool{log: log}, nil
}

func (t *ActionLogUndoTool) Name() string   { return "actionlog_undo" }
func (t *ActionLogUndoTool) Description() string {
	return "Undo the last action"
}

func (t *ActionLogUndoTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *ActionLogUndoTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "actionlog_undo",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{"type": "string", "description": "Project path"},
				},
			},
		},
	}
}

func (t *ActionLogUndoTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	action, err := t.log.UndoLast(project)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Undone: %s", action.Description), nil
}

type ActionLogDiffTool struct {
	log *ActionLog
}

func NewActionLogDiffTool(basePath string) (*ActionLogDiffTool, error) {
	log, err := New(basePath, 1000)
	if err != nil {
		return nil, err
	}
	return &ActionLogDiffTool{log: log}, nil
}

func (t *ActionLogDiffTool) Name() string   { return "actionlog_diff" }
func (t *ActionLogDiffTool) Description() string {
	return "Get diff for an action"
}

func (t *ActionLogDiffTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *ActionLogDiffTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "actionlog_diff",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action_id": map[string]interface{}{"type": "string", "description": "Action ID"},
				},
				"required": []string{"action_id"},
			},
		},
	}
}

func (t *ActionLogDiffTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	actionID, _ := args["action_id"].(string)
	if actionID == "" {
		return "", fmt.Errorf("action_id is required")
	}
	diff, err := t.log.GetDiff(actionID)
	if err != nil {
		return "", err
	}
	return diff, nil
}

func RegisterActionLogTools(reg *tools.Registry, basePath string) error {
	listTool, err := NewActionLogTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(listTool)

	undoTool, err := NewActionLogUndoTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(undoTool)

	diffTool, err := NewActionLogDiffTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(diffTool)

	return nil
}
package diffsandbox

import (
	"context"
	"fmt"

	"github.com/hermes-v2/aigo/internal/tools"
)

type SandboxAddTool struct {
	sandbox *Sandbox
}

func NewSandboxAddTool(basePath string) (*SandboxAddTool, error) {
	sandbox, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &SandboxAddTool{sandbox: sandbox}, nil
}

func (t *SandboxAddTool) Name() string   { return "sandbox_add" }
func (t *SandboxAddTool) Description() string {
	return "Add a change to diff sandbox for review"
}

func (t *SandboxAddTool) Annotations() tools.Annotations {
	return tools.Annotations{}
}

func (t *SandboxAddTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "sandbox_add",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project":     map[string]interface{}{"type": "string", "description": "Project path"},
					"file":        map[string]interface{}{"type": "string", "description": "File path"},
					"old_content":  map[string]interface{}{"type": "string", "description": "Original content"},
					"new_content": map[string]interface{}{"type": "string", "description": "New content"},
				},
				"required": []string{"project", "file", "new_content"},
			},
		},
	}
}

func (t *SandboxAddTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	file, _ := args["file"].(string)
	oldContent, _ := args["old_content"].(string)
	newContent, _ := args["new_content"].(string)

	if file == "" || newContent == "" {
		return "", fmt.Errorf("file and new_content are required")
	}
	id, err := t.sandbox.Add(project, file, oldContent, newContent)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Added: %s", id[:8]), nil
}

type SandboxListTool struct {
	sandbox *Sandbox
}

func NewSandboxListTool(basePath string) (*SandboxListTool, error) {
	sandbox, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &SandboxListTool{sandbox: sandbox}, nil
}

func (t *SandboxListTool) Name() string   { return "sandbox_list" }
func (t *SandboxListTool) Description() string {
	return "List pending changes in sandbox"
}

func (t *SandboxListTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *SandboxListTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "sandbox_list",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{"type": "string", "description": "Project path"},
				},
				"required": []string{"project"},
			},
		},
	}
}

func (t *SandboxListTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	if project == "" {
		return "", fmt.Errorf("project is required")
	}
	changes := t.sandbox.Pending(project)
	if len(changes) == 0 {
		return "No pending changes", nil
	}
	var result string
	for _, c := range changes {
		result += fmt.Sprintf("%s | %s\n", c.ID[:8], c.File)
	}
	return result, nil
}

type SandboxApplyTool struct {
	sandbox *Sandbox
}

func NewSandboxApplyTool(basePath string) (*SandboxApplyTool, error) {
	sandbox, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &SandboxApplyTool{sandbox: sandbox}, nil
}

func (t *SandboxApplyTool) Name() string   { return "sandbox_apply" }
func (t *SandboxApplyTool) Description() string {
	return "Apply a change from sandbox"
}

func (t *SandboxApplyTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *SandboxApplyTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "sandbox_apply",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{"type": "string", "description": "Project path"},
					"change_id": map[string]interface{}{"type": "string", "description": "Change ID (omit for all)"},
				},
				"required": []string{"project"},
			},
		},
	}
}

func (t *SandboxApplyTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	changeID, _ := args["change_id"].(string)

	if project == "" {
		return "", fmt.Errorf("project is required")
	}
	if changeID != "" {
		if err := t.sandbox.Apply(project, changeID); err != nil {
			return "", err
		}
		return "Applied", nil
	}
	if err := t.sandbox.ApplyAll(project); err != nil {
		return "", err
	}
	return "All applied", nil
}

type SandboxRejectTool struct {
	sandbox *Sandbox
}

func NewSandboxRejectTool(basePath string) (*SandboxRejectTool, error) {
	sandbox, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &SandboxRejectTool{sandbox: sandbox}, nil
}

func (t *SandboxRejectTool) Name() string   { return "sandbox_reject" }
func (t *SandboxRejectTool) Description() string {
	return "Reject a change from sandbox"
}

func (t *SandboxRejectTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true}
}

func (t *SandboxRejectTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "sandbox_reject",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{"type": "string", "description": "Project path"},
					"change_id": map[string]interface{}{"type": "string", "description": "Change ID (omit for all)"},
				},
				"required": []string{"project"},
			},
		},
	}
}

func (t *SandboxRejectTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	changeID, _ := args["change_id"].(string)

	if project == "" {
		return "", fmt.Errorf("project is required")
	}
	if changeID != "" {
		if err := t.sandbox.Reject(project, changeID); err != nil {
			return "", err
		}
		return "Rejected", nil
	}
	t.sandbox.RejectAll(project)
	return "All rejected", nil
}

type SandboxShowTool struct {
	sandbox *Sandbox
}

func NewSandboxShowTool(basePath string) (*SandboxShowTool, error) {
	sandbox, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &SandboxShowTool{sandbox: sandbox}, nil
}

func (t *SandboxShowTool) Name() string   { return "sandbox_show" }
func (t *SandboxShowTool) Description() string {
	return "Show diff for a sandbox change"
}

func (t *SandboxShowTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *SandboxShowTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "sandbox_show",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project":  map[string]interface{}{"type": "string", "description": "Project path"},
					"change_id": map[string]interface{}{"type": "string", "description": "Change ID"},
				},
				"required": []string{"project", "change_id"},
			},
		},
	}
}

func (t *SandboxShowTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	project, _ := args["project"].(string)
	changeID, _ := args["change_id"].(string)

	if project == "" || changeID == "" {
		return "", fmt.Errorf("project and change_id are required")
	}
	diff, err := t.sandbox.Diff(project, changeID)
	if err != nil {
		return "", err
	}
	return diff, nil
}

func RegisterSandboxTools(reg *tools.Registry, basePath string) error {
	addTool, err := NewSandboxAddTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(addTool)

	listTool, err := NewSandboxListTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(listTool)

	applyTool, err := NewSandboxApplyTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(applyTool)

	rejectTool, err := NewSandboxRejectTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(rejectTool)

	showTool, err := NewSandboxShowTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(showTool)

	return nil
}
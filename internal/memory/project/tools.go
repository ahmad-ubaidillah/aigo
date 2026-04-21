package project

import (
	"context"
	"fmt"

	"github.com/hermes-v2/aigo/internal/tools"
)

type ProjectContextTool struct {
	pm *ProjectStore
}

func NewProjectContextTool(basePath string) (*ProjectContextTool, error) {
	pm, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &ProjectContextTool{pm: pm}, nil
}

func (t *ProjectContextTool) Name() string   { return "project_context" }
func (t *ProjectContextTool) Description() string {
	return "Get project context (type, structure, last edits)"
}

func (t *ProjectContextTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *ProjectContextTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "project_context",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *ProjectContextTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	p := t.pm.Get(".")
	if p == nil {
		return "No project context", nil
	}
	return p.Context, nil
}

type ProjectAddFactTool struct {
	pm *ProjectStore
}

func NewProjectAddFactTool(basePath string) (*ProjectAddFactTool, error) {
	pm, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &ProjectAddFactTool{pm: pm}, nil
}

func (t *ProjectAddFactTool) Name() string { return "project_add_fact" }
func (t *ProjectAddFactTool) Description() string {
	return "Add a fact to project memory"
}

func (t *ProjectAddFactTool) Annotations() tools.Annotations {
	return tools.Annotations{}
}

func (t *ProjectAddFactTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "project_add_fact",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fact": map[string]interface{}{"type": "string", "description": "Fact to add"},
				},
				"required": []string{"fact"},
			},
		},
	}
}

func (t *ProjectAddFactTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	fact, _ := args["fact"].(string)
	if fact == "" {
		return "", fmt.Errorf("fact is required")
	}
	if err := t.pm.Remember(".", fact); err != nil {
		return "", err
	}
	return "Added", nil
}

func RegisterProjectMemoryTools(reg *tools.Registry, basePath string) error {
	ctxTool, err := NewProjectContextTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(ctxTool)

	factTool, err := NewProjectAddFactTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(factTool)

	return nil
}
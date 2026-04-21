package codex

import (
	"context"
	"fmt"

	"github.com/hermes-v2/aigo/internal/tools"
)

type CodexIndexTool struct {
	idx *CodeIndex
}

func NewCodexIndexTool(basePath string) (*CodexIndexTool, error) {
	idx, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &CodexIndexTool{idx: idx}, nil
}

func (t *CodexIndexTool) Name() string { return "codex_index" }
func (t *CodexIndexTool) Description() string {
	return "Index project symbols and errors"
}

func (t *CodexIndexTool) Annotations() tools.Annotations {
	return tools.Annotations{}
}

func (t *CodexIndexTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "codex_index",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_dir": map[string]interface{}{"type": "string", "description": "Project directory to index"},
				},
			},
		},
	}
}

func (t *CodexIndexTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	dir, _ := args["project_dir"].(string)
	if dir == "" {
		return "", fmt.Errorf("project_dir is required")
	}
	if err := t.idx.Index(); err != nil {
		return "", err
	}
	return "Indexed", nil
}

type CodexFindSymbolTool struct {
	idx *CodeIndex
}

func NewCodexFindSymbolTool(basePath string) (*CodexFindSymbolTool, error) {
	idx, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &CodexFindSymbolTool{idx: idx}, nil
}

func (t *CodexFindSymbolTool) Name() string { return "codex_find_symbol" }
func (t *CodexFindSymbolTool) Description() string {
	return "Find a symbol definition in the project"
}

func (t *CodexFindSymbolTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *CodexFindSymbolTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "codex_find_symbol",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{"type": "string", "description": "Symbol name to find"},
				},
				"required": []string{"symbol"},
			},
		},
	}
}

func (t *CodexFindSymbolTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}
	locations := t.idx.Find(symbol)
	if len(locations) == 0 {
		return "Symbol not found", nil
	}
	var result string
	for _, loc := range locations {
		result += fmt.Sprintf("%s:%d - %s\n", loc.File, loc.Line, loc.Kind)
	}
	return result, nil
}

type CodexMapErrorTool struct {
	idx *CodeIndex
}

func NewCodexMapErrorTool(basePath string) (*CodexMapErrorTool, error) {
	idx, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &CodexMapErrorTool{idx: idx}, nil
}

func (t *CodexMapErrorTool) Name() string { return "codex_map_error" }
func (t *CodexMapErrorTool) Description() string {
	return "Map an error to its source location"
}

func (t *CodexMapErrorTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *CodexMapErrorTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "codex_map_error",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{"type": "string", "description": "Error message"},
				},
				"required": []string{"error"},
			},
		},
	}
}

func (t *CodexMapErrorTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	errMsg, _ := args["error"].(string)
	if errMsg == "" {
		return "", fmt.Errorf("error is required")
	}
	locations := t.idx.MapError(errMsg)
	if len(locations) == 0 {
		return "Unknown error", nil
	}
	var result string
	for _, loc := range locations {
		result += fmt.Sprintf("%s:%d - %s\n", loc.File, loc.Line, loc.Kind)
	}
	return result, nil
}

func RegisterCodexTools(reg *tools.Registry, basePath string) error {
	indexTool, err := NewCodexIndexTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(indexTool)

	findSymbolTool, err := NewCodexFindSymbolTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(findSymbolTool)

	mapErrorTool, err := NewCodexMapErrorTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(mapErrorTool)

	return nil
}
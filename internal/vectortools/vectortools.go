// Package vectortools registers vector memory tools in Aigo's tool registry.
// These tools expose semantic vector memory (save, search, stats) to the agent.
package vectortools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hermes-v2/aigo/internal/memory/vector"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterVectorTools registers all vector memory tools in the registry.
func RegisterVectorTools(reg *tools.Registry, vs *vector.VectorStore) {
	reg.Register(&VecSaveTool{vs: vs})
	reg.Register(&VecSearchTool{vs: vs})
	reg.Register(&VecSemanticSearchTool{vs: vs})
	reg.Register(&VecStatsTool{vs: vs})
}

// --- vec_save ---

type VecSaveTool struct {
	vs *vector.VectorStore
}

func (t *VecSaveTool) Name() string { return "vec_save" }
func (t *VecSaveTool) Description() string {
	return "Save text to semantic vector memory with auto-embedding. Use for storing knowledge, facts, decisions, and context that need semantic recall."
}
func (t *VecSaveTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"database"}}
}
func (t *VecSaveTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "vec_save",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "The text content to save",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Category for organization (default: general)",
						"default":     "general",
					},
					"tags": map[string]interface{}{
						"type":        "string",
						"description": "Comma-separated tags for filtering",
					},
				},
				"required": []string{"text"},
			},
		},
	}
}

func (t *VecSaveTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	text, _ := args["text"].(string)
	if text == "" {
		return "", fmt.Errorf("vec_save: text is required")
	}
	category, _ := args["category"].(string)
	if category == "" {
		category = "general"
	}
	tags, _ := args["tags"].(string)

	id, err := t.vs.Save(text, category, tags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("✅ Saved to vector memory (id=%d, category=%s)", id, category), nil
}

// --- vec_search ---

type VecSearchTool struct {
	vs *vector.VectorStore
}

func (t *VecSearchTool) Name() string { return "vec_search" }
func (t *VecSearchTool) Description() string {
	return "Hybrid semantic + keyword search in vector memory. Combines meaning-based and keyword-based retrieval for best results."
}
func (t *VecSearchTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *VecSearchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "vec_search",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max results (default: 5)",
						"default":     5,
					},
					"alpha": map[string]interface{}{
						"type":        "number",
						"description": "Semantic vs keyword weight: 1.0=pure semantic, 0.0=pure keyword (default: 0.7)",
						"default":     0.7,
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *VecSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("vec_search: query is required")
	}
	limit := 5
	if l, ok := args["limit"]; ok {
		switch v := l.(type) {
		case float64:
			limit = int(v)
		case string:
			limit, _ = strconv.Atoi(v)
		}
	}
	alpha := 0.7
	if a, ok := args["alpha"]; ok {
		if f, ok := a.(float64); ok {
			alpha = f
		}
	}

	results, err := t.vs.Search(query, limit, alpha)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "No results found in vector memory.", nil
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return string(out), nil
}

// --- vec_semantic_search ---

type VecSemanticSearchTool struct {
	vs *vector.VectorStore
}

func (t *VecSemanticSearchTool) Name() string { return "vec_semantic_search" }
func (t *VecSemanticSearchTool) Description() string {
	return "Pure semantic similarity search (cosine). Best for finding conceptually similar content regardless of exact keywords."
}
func (t *VecSemanticSearchTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *VecSemanticSearchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "vec_semantic_search",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Semantic query (meaning-based, not keyword)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max results (default: 5)",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *VecSemanticSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("vec_semantic_search: query is required")
	}
	limit := 5
	if l, ok := args["limit"]; ok {
		switch v := l.(type) {
		case float64:
			limit = int(v)
		case string:
			limit, _ = strconv.Atoi(v)
		}
	}

	results, err := t.vs.SemanticSearch(query, limit)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "No semantically similar content found.", nil
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return string(out), nil
}

// --- vec_stats ---

type VecStatsTool struct {
	vs *vector.VectorStore
}

func (t *VecStatsTool) Name() string { return "vec_stats" }
func (t *VecStatsTool) Description() string {
	return "Show vector memory statistics: total entries, categories, embedding dimensions."
}
func (t *VecStatsTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *VecStatsTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "vec_stats",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *VecStatsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	count, err := t.vs.Count()
	if err != nil {
		return "", err
	}
	cats, err := t.vs.Categories()
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"total_entries":      count,
		"embedding_dim":      256,
		"embedding_method":   "SimHash",
		"categories":         cats,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

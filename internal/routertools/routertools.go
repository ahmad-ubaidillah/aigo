// Package routertools registers semantic router tools in Aigo's tool registry.
package routertools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hermes-v2/aigo/internal/router"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterRouterTools registers router tools.
func RegisterRouterTools(reg *tools.Registry, r *router.Router) {
	reg.Register(&RouteQueryTool{router: r})
	reg.Register(&RouterStatsTool{router: r})
	reg.Register(&RouterConfigTool{router: r})
}

// --- route_query ---

type RouteQueryTool struct {
	router *router.Router
}

func (t *RouteQueryTool) Name() string { return "route_query" }
func (t *RouteQueryTool) Description() string {
	return "Analyze a query and determine the best model/category to handle it. Uses semantic routing to match task type to model strengths."
}
func (t *RouteQueryTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *RouteQueryTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "route_query",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The query or task to route",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *RouteQueryTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("route_query: query is required")
	}

	result := t.router.RouteForQuery(ctx, query, router.SimpleClassify)
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// --- router_stats ---

type RouterStatsTool struct {
	router *router.Router
}

func (t *RouterStatsTool) Name() string { return "router_stats" }
func (t *RouterStatsTool) Description() string {
	return "Show semantic router statistics: which categories have been routed, total decisions made."
}
func (t *RouterStatsTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *RouterStatsTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "router_stats",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *RouterStatsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	stats := t.router.GetStats()
	cfg := t.router.GetConfig()

	result := map[string]interface{}{
		"routing_stats":   stats,
		"default_model":   cfg.DefaultModel,
		"cheap_model":     cfg.CheapModel,
		"auto_classify":   cfg.AutoClassify,
		"routes_count":    len(cfg.Routes),
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// --- router_config ---

type RouterConfigTool struct {
	router *router.Router
}

func (t *RouterConfigTool) Name() string { return "router_config" }
func (t *RouterConfigTool) Description() string {
	return "Get or update semantic router configuration at runtime."
}
func (t *RouterConfigTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"config"}}
}
func (t *RouterConfigTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "router_config",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"get", "update"},
						"description": "Get current config or update it",
						"default":     "get",
					},
					"default_model": map[string]interface{}{
						"type":        "string",
						"description": "New default model (for update action)",
					},
					"cheap_model": map[string]interface{}{
						"type":        "string",
						"description": "New cheap/fast model (for update action)",
					},
					"auto_classify": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable/disable auto classification",
					},
				},
			},
		},
	}
}

func (t *RouterConfigTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	if action == "" {
		action = "get"
	}

	if action == "get" {
		cfg := t.router.GetConfig()
		out, _ := json.MarshalIndent(cfg, "", "  ")
		return string(out), nil
	}

	// Update
	cfg := t.router.GetConfig()
	if dm, ok := args["default_model"].(string); ok && dm != "" {
		cfg.DefaultModel = dm
	}
	if cm, ok := args["cheap_model"].(string); ok {
		cfg.CheapModel = cm
	}
	if ac, ok := args["auto_classify"].(bool); ok {
		cfg.AutoClassify = ac
	}
	t.router.UpdateConfig(cfg)

	return "✅ Router config updated.", nil
}

// Package browsertools registers browser automation tools for Aigo.
package browsertools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hermes-v2/aigo/internal/browser"
	"github.com/hermes-v2/aigo/internal/tools"
)

const lightpandaWS = "ws://localhost:9222"

// RegisterBrowserTools adds browser_inspect, browser_workflow_run,
// and browser_workflow_validate to the tool registry.
func RegisterBrowserTools(reg *tools.Registry) {
	reg.Register(&BrowserInspectTool{})
	reg.Register(&BrowserWorkflowRunTool{})
	reg.Register(&BrowserWorkflowValidateTool{})
}

// --- browser_inspect ---

type BrowserInspectTool struct{}

func (t *BrowserInspectTool) Name() string { return "browser_inspect" }
func (t *BrowserInspectTool) Description() string {
	return "Inspect a web page to discover interactive elements (inputs, buttons, links, selects). Returns JSON with selectors and metadata for each element."
}
func (t *BrowserInspectTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *BrowserInspectTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "browser_inspect",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "The URL to inspect for interactive elements",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}
func (t *BrowserInspectTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	result, err := browser.Inspect(lightpandaWS, url)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// --- browser_workflow_run ---

type BrowserWorkflowRunTool struct{}

func (t *BrowserWorkflowRunTool) Name() string { return "browser_workflow_run" }
func (t *BrowserWorkflowRunTool) Description() string {
	return "Execute a YAML browser workflow. Provide the YAML content and the action name to run."
}
func (t *BrowserWorkflowRunTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false, SideEffects: []string{"network"}}
}
func (t *BrowserWorkflowRunTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "browser_workflow_run",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"yaml_content": map[string]interface{}{
						"type":        "string",
						"description": "The YAML workflow content",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "The action name to execute",
					},
				},
				"required": []string{"yaml_content", "action"},
			},
		},
	}
}
func (t *BrowserWorkflowRunTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	yamlContent, _ := args["yaml_content"].(string)
	actionName, _ := args["action"].(string)
	if yamlContent == "" || actionName == "" {
		return "", fmt.Errorf("yaml_content and action are required")
	}

	wf, err := browser.LoadFromBytes([]byte(yamlContent))
	if err != nil {
		return "", err
	}

	exec := browser.NewExecutor(lightpandaWS)
	result, err := exec.Run(wf, actionName)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// --- browser_workflow_validate ---

type BrowserWorkflowValidateTool struct{}

func (t *BrowserWorkflowValidateTool) Name() string { return "browser_workflow_validate" }
func (t *BrowserWorkflowValidateTool) Description() string {
	return "Validate YAML browser workflow syntax without executing it. Returns parsed structure or error."
}
func (t *BrowserWorkflowValidateTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *BrowserWorkflowValidateTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "browser_workflow_validate",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"yaml_content": map[string]interface{}{
						"type":        "string",
						"description": "The YAML workflow content to validate",
					},
				},
				"required": []string{"yaml_content"},
			},
		},
	}
}
func (t *BrowserWorkflowValidateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	yamlContent, _ := args["yaml_content"].(string)
	if yamlContent == "" {
		return "", fmt.Errorf("yaml_content is required")
	}

	wf, err := browser.LoadFromBytes([]byte(yamlContent))
	if err != nil {
		return fmt.Sprintf("INVALID: %v", err), nil
	}

	// Summarize the workflow
	summary := map[string]interface{}{
		"valid":   true,
		"name":    wf.Name,
		"actions": len(wf.Actions),
	}
	actionNames := make([]string, 0, len(wf.Actions))
	for name := range wf.Actions {
		actionNames = append(actionNames, name)
	}
	summary["action_names"] = actionNames

	b, _ := json.MarshalIndent(summary, "", "  ")
	return string(b), nil
}

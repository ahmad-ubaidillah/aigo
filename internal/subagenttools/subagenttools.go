// Package subagenttools registers sub-agent delegation tools.
// Inspired by Oh My OpenAgent's Discipline Agents:
//   - delegate: Spawn a specialized sub-agent for a task
//   - ultrawork: Decompose + parallel execute (OMO's ultrawork)
//   - intent_gate: Analyze true user intent before acting
//   - subagent_status: Check sub-agent progress
//   - subagent_results: Get results from completed sub-agents
package subagenttools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hermes-v2/aigo/internal/subagent"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterSubAgentTools registers all sub-agent tools.
func RegisterSubAgentTools(reg *tools.Registry, orch *subagent.Orchestrator, brainFunc subagent.BrainFunc) {
	reg.Register(&DelegateTool{orch: orch})
	reg.Register(&UltraworkTool{orch: orch, brainFunc: brainFunc})
	reg.Register(&IntentGateTool{brainFunc: brainFunc})
	reg.Register(&SubAgentStatusTool{orch: orch})
	reg.Register(&SubAgentResultsTool{orch: orch})
}

// --- delegate ---

type DelegateTool struct {
	orch *subagent.Orchestrator
}

func (t *DelegateTool) Name() string { return "delegate" }
func (t *DelegateTool) Description() string {
	return "Delegate a task to a specialized sub-agent. Choose the right role for the job: builder (code), reasoner (architecture), explorer (research), memory (knowledge recall)."
}
func (t *DelegateTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"subagent"}}
}
func (t *DelegateTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "delegate",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"goal": map[string]interface{}{
						"type":        "string",
						"description": "What the sub-agent should accomplish",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"builder", "reasoner", "explorer", "memory"},
						"description": "Specialist role: builder (code), reasoner (analysis), explorer (research), memory (recall)",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"deep", "quick", "ultrabrain", "visual", "general"},
						"description": "Task category for model routing: deep (autonomous), quick (fast fix), ultrabrain (hard logic)",
						"default":     "general",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Lean context to pass to the sub-agent (file paths, relevant info)",
					},
				},
				"required": []string{"goal", "role"},
			},
		},
	}
}

func (t *DelegateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	goal, _ := args["goal"].(string)
	if goal == "" {
		return "", fmt.Errorf("delegate: goal is required")
	}

	roleStr, _ := args["role"].(string)
	role := subagent.Role(roleStr)
	if role == "" {
		role = subagent.RoleBuilder
	}

	categoryStr, _ := args["category"].(string)
	category := subagent.Category(categoryStr)
	if category == "" {
		category = subagent.CategoryGeneral
	}

	contextInfo, _ := args["context"].(string)

	task := subagent.Task{
		ID:       "manual-1",
		Goal:     goal,
		Role:     role,
		Category: category,
		Context:  contextInfo,
		Priority: 1,
	}

	results, err := t.orch.Execute(ctx, []subagent.Task{task})
	if err != nil {
		return "", fmt.Errorf("delegate: %w", err)
	}

	if result, ok := results["manual-1"]; ok {
		return result, nil
	}
	return "Sub-agent completed but no result returned.", nil
}

// --- ultrawork ---

type UltraworkTool struct {
	orch      *subagent.Orchestrator
	brainFunc subagent.BrainFunc
}

func (t *UltraworkTool) Name() string { return "ultrawork" }
func (t *UltraworkTool) Description() string {
	return "One command to rule them all. Decomposes a complex goal into parallel subtasks and fires multiple specialized sub-agents simultaneously. Like OMO's ultrawork — doesn't stop until done."
}
func (t *UltraworkTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"subagent", "filesystem", "network"}}
}
func (t *UltraworkTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "ultrawork",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"goal": map[string]interface{}{
						"type":        "string",
						"description": "The complex goal to accomplish. Be specific.",
					},
					"max_agents": map[string]interface{}{
						"type":        "integer",
						"description": "Max parallel sub-agents (default: 5)",
						"default":     5,
					},
				},
				"required": []string{"goal"},
			},
		},
	}
}

// truncateStr truncates a string to maxLen with ellipsis.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (t *UltraworkTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	goal, _ := args["goal"].(string)
	if goal == "" {
		return "", fmt.Errorf("ultrawork: goal is required")
	}

	// Phase 1: Intent analysis
	intent, err := subagent.IntentAnalysis(ctx, t.brainFunc, goal)
	if err != nil {
		// Fallback: treat as single task
		intent = &subagent.IntentResult{
			TrueIntent:  goal,
			Category:    subagent.CategoryGeneral,
			Complexity:  5,
			NeedsDecomp: true,
		}
	}

	// Phase 2: Task decomposition
	tasks, err := t.orch.Decompose(ctx, intent.TrueIntent)
	if err != nil {
		return "", fmt.Errorf("ultrawork: decompose: %w", err)
	}

	// Phase 3: Parallel execution
	results, err := t.orch.Execute(ctx, tasks)
	if err != nil {
		return "", fmt.Errorf("ultrawork: execute: %w", err)
	}

	// Phase 4: Synthesize results
	var sb strings.Builder
	sb.WriteString("## 🏋️ Ultrawork Results\n\n")
	sb.WriteString(fmt.Sprintf("**Intent:** %s\n", intent.TrueIntent))
	sb.WriteString(fmt.Sprintf("**Complexity:** %d/10\n", intent.Complexity))
	sb.WriteString(fmt.Sprintf("**Tasks completed:** %d\n\n", len(results)))

	for taskID, result := range results {
		sb.WriteString(fmt.Sprintf("### Task: %s\n", taskID))
		sb.WriteString(truncateStr(result, 500))
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}

// --- intent_gate ---

type IntentGateTool struct {
	brainFunc subagent.BrainFunc
}

func (t *IntentGateTool) Name() string { return "intent_gate" }
func (t *IntentGateTool) Description() string {
	return "Analyze the true intent behind a user message before acting. Like OMO's IntentGate — prevents literal misinterpretation. Use this when the user's request is ambiguous."
}
func (t *IntentGateTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *IntentGateTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "intent_gate",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The user message to analyze",
					},
				},
				"required": []string{"message"},
			},
		},
	}
}

func (t *IntentGateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	message, _ := args["message"].(string)
	if message == "" {
		return "", fmt.Errorf("intent_gate: message is required")
	}

	intent, err := subagent.IntentAnalysis(ctx, t.brainFunc, message)
	if err != nil {
		return "", err
	}

	out, _ := json.MarshalIndent(intent, "", "  ")
	return string(out), nil
}

// --- subagent_status ---

type SubAgentStatusTool struct {
	orch *subagent.Orchestrator
}

func (t *SubAgentStatusTool) Name() string { return "subagent_status" }
func (t *SubAgentStatusTool) Description() string {
	return "Check the status of all sub-agents: running, completed, failed."
}
func (t *SubAgentStatusTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SubAgentStatusTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "subagent_status",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *SubAgentStatusTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	agents := t.orch.ListAgents()
	if len(agents) == 0 {
		return "No sub-agents have been spawned yet.", nil
	}

	out, _ := json.MarshalIndent(agents, "", "  ")
	return string(out), nil
}

// --- subagent_results ---

type SubAgentResultsTool struct {
	orch *subagent.Orchestrator
}

func (t *SubAgentResultsTool) Name() string { return "subagent_results" }
func (t *SubAgentResultsTool) Description() string {
	return "Get all results from completed sub-agents."
}
func (t *SubAgentResultsTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SubAgentResultsTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "subagent_results",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *SubAgentResultsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	results := t.orch.GetResults()
	if len(results) == 0 {
		return "No sub-agent results yet.", nil
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return string(out), nil
}

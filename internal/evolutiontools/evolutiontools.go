// Package evolutiontools implements tool functions for the self-evolution system.
package evolutiontools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hermes-v2/aigo/internal/evolution"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterEvolutionTools registers all evolution tools with the given registry.
func RegisterEvolutionTools(reg *tools.Registry, mgr *evolution.Manager) {
	reg.Register(&EvolveProposeTool{mgr: mgr})
	reg.Register(&EvolveApplyTool{mgr: mgr})
	reg.Register(&EvolveRevertTool{mgr: mgr})
	reg.Register(&EvolveHistoryTool{mgr: mgr})
	reg.Register(&EvolveManifestTool{mgr: mgr})
	reg.Register(&EvolveContractTool{mgr: mgr})
}

// --- evolve_propose ---
type EvolveProposeTool struct{ mgr *evolution.Manager }

func (t *EvolveProposeTool) Name() string       { return "evolve_propose" }
func (t *EvolveProposeTool) Description() string { return "Propose a code change to improve the agent. Provide either find+replace for targeted edits or new_content for full rewrite." }
func (t *EvolveProposeTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true, ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *EvolveProposeTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "evolve_propose",
			Description: "Propose a code change. Provide 'file' path (relative to project root), and either 'find'+'replace' for a targeted edit, or 'new_content' for a full file rewrite. Must include 'reason'.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file":         map[string]string{"type": "string", "description": "Path to .go file (relative to project dir)"},
					"find":         map[string]string{"type": "string", "description": "Text to find (for find/replace mode)"},
					"replace":      map[string]string{"type": "string", "description": "Replacement text (for find/replace mode)"},
					"new_content":  map[string]string{"type": "string", "description": "Full new file content (for full rewrite mode)"},
					"reason":       map[string]string{"type": "string", "description": "Reason for this change"},
				},
				"required": []string{"file", "reason"},
			},
		},
	}
}
func (t *EvolveProposeTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	file, _ := args["file"].(string)
	find, _ := args["find"].(string)
	replace, _ := args["replace"].(string)
	newContent, _ := args["new_content"].(string)
	reason, _ := args["reason"].(string)

	if file == "" || reason == "" {
		return "", fmt.Errorf("file and reason are required")
	}

	proposal, err := t.mgr.Propose(file, find, replace, newContent, reason)
	if err != nil {
		return "", err
	}

	data, _ := json.MarshalIndent(proposal, "", "  ")
	return string(data), nil
}

// --- evolve_apply ---
type EvolveApplyTool struct{ mgr *evolution.Manager }

func (t *EvolveApplyTool) Name() string       { return "evolve_apply" }
func (t *EvolveApplyTool) Description() string { return "Apply a proposed evolution change by ID. Runs go build to verify, auto-reverts on failure." }
func (t *EvolveApplyTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true, ReadOnly: false, SideEffects: []string{"filesystem", "system"}}
}
func (t *EvolveApplyTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "evolve_apply",
			Description: "Apply a proposed change by its ID. Creates a backup (.bak), applies the change, and runs go build. Auto-reverts if build fails.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]string{"type": "string", "description": "Proposal ID (from evolve_propose)"},
				},
				"required": []string{"id"},
			},
		},
	}
}
func (t *EvolveApplyTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}
	return t.mgr.Apply(id)
}

// --- evolve_revert ---
type EvolveRevertTool struct{ mgr *evolution.Manager }

func (t *EvolveRevertTool) Name() string       { return "evolve_revert" }
func (t *EvolveRevertTool) Description() string { return "Revert an applied evolution change by ID, restoring from backup." }
func (t *EvolveRevertTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *EvolveRevertTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "evolve_revert",
			Description: "Revert an applied change by restoring the .bak backup file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]string{"type": "string", "description": "Proposal ID to revert"},
				},
				"required": []string{"id"},
			},
		},
	}
}
func (t *EvolveRevertTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return "", fmt.Errorf("id is required")
	}
	return t.mgr.Revert(id)
}

// --- evolve_history ---
type EvolveHistoryTool struct{ mgr *evolution.Manager }

func (t *EvolveHistoryTool) Name() string       { return "evolve_history" }
func (t *EvolveHistoryTool) Description() string { return "Show all evolution proposals and their status." }
func (t *EvolveHistoryTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *EvolveHistoryTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "evolve_history",
			Description: "List all evolution proposals with their status (pending, applied, reverted, failed).",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *EvolveHistoryTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	history := t.mgr.History()
	if len(history) == 0 {
		return "No evolution proposals yet.", nil
	}
	data, _ := json.MarshalIndent(history, "", "  ")
	return string(data), nil
}

// --- evolve_manifest ---
type EvolveManifestTool struct{ mgr *evolution.Manager }

func (t *EvolveManifestTool) Name() string       { return "evolve_manifest" }
func (t *EvolveManifestTool) Description() string { return "Generate a YAML manifest of Aigo's codebase. Lists packages, exported types, functions, and dependencies." }
func (t *EvolveManifestTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true, SideEffects: []string{}}
}
func (t *EvolveManifestTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "evolve_manifest",
			Description: "Scan all Go files in internal/ and cmd/ directories. Produces a YAML manifest with package info, exported types/functions, and inter-package dependencies.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"output_path": map[string]string{
						"type":        "string",
						"description": "Optional: save manifest to this file path instead of returning it",
					},
				},
			},
		},
	}
}
func (t *EvolveManifestTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	yamlStr, err := evolution.GenerateManifest(t.mgr.ProjectDir())
	if err != nil {
		return "", fmt.Errorf("generating manifest: %w", err)
	}

	// Optionally save to file
	if outputPath, ok := args["output_path"].(string); ok && outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return "", fmt.Errorf("creating output directory: %w", err)
		}
		if err := os.WriteFile(outputPath, []byte(yamlStr), 0644); err != nil {
			return "", fmt.Errorf("writing manifest: %w", err)
		}
		return fmt.Sprintf("Manifest saved to %s\n\n%s", outputPath, yamlStr), nil
	}

	return yamlStr, nil
}

// --- evolve_contract ---
type EvolveContractTool struct{ mgr *evolution.Manager }

func (t *EvolveContractTool) Name() string       { return "evolve_contract" }
func (t *EvolveContractTool) Description() string { return "Auto-validate after code changes. Runs go build, go vet, and tool registration checks." }
func (t *EvolveContractTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true, SideEffects: []string{"system"}}
}
func (t *EvolveContractTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "evolve_contract",
			Description: "Run contract validation after code changes. Steps: 1) go build (compilation), 2) go vet (code quality), 3) Register*Tools function check. Pass a file path or 'all' for full project.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"target": map[string]string{
						"type":        "string",
						"description": "File path to validate, or 'all' for full project validation",
					},
				},
				"required": []string{"target"},
			},
		},
	}
}
func (t *EvolveContractTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	target, _ := args["target"].(string)
	if target == "" {
		return "", fmt.Errorf("target is required")
	}

	result, err := evolution.RunContract(t.mgr.ProjectDir(), target)
	if err != nil {
		return "", fmt.Errorf("contract validation: %w", err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

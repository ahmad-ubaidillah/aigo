// Package skillhubtools provides agent tool interfaces to the OnlineHub skill marketplace.
// Connects Aigo's tools to the full 1700+ skill index (SQLite FTS5, Smithery, GitHub, LobeHub).
package skillhubtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hermes-v2/aigo/internal/skillhub"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterSkillHubTools registers all skill hub tools using OnlineHub as backend.
func RegisterSkillHubTools(reg *tools.Registry, hub *skillhub.OnlineHub) {
	reg.Register(&SkillSearchTool{hub: hub})
	reg.Register(&SkillInstallTool{hub: hub})
	reg.Register(&SkillListTool{hub: hub})
	reg.Register(&SkillPopularTool{hub: hub})
	reg.Register(&SkillSyncTool{hub: hub})
	reg.Register(&SkillInfoTool{hub: hub})
	reg.Register(&SkillRemoveTool{hub: hub})
	reg.Register(&SkillStatsTool{hub: hub})
}

// --- skill_search ---

type SkillSearchTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillSearchTool) Name() string { return "skill_search" }
func (t *SkillSearchTool) Description() string {
	return "Search for skills in the marketplace by name, description, or tags. Searches across 1700+ indexed skills from Smithery, Hermes, Anthropic, and LobeHub."
}
func (t *SkillSearchTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SkillSearchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_search",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (name, description, or tags)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max results (default: 10)",
						"default":     10,
					},
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Filter by source: smithery, hermes, github, lobehub",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *SkillSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("skill_search: query is required")
	}
	limit := 10
	if l, ok := args["limit"]; ok {
		if f, ok := l.(float64); ok {
			limit = int(f)
		}
	}
	source, _ := args["source"].(string)

	var results []skillhub.Skill
	var err error

	if source != "" {
		results, err = t.hub.BrowseOnline(source, limit)
	} else {
		results, err = t.hub.Search(query, limit)
	}

	if err != nil {
		return fmt.Sprintf("Search error: %v", err), nil
	}
	if len(results) == 0 {
		return fmt.Sprintf("No skills found for '%s'", query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d skills:\n\n", len(results)))
	for i, s := range results {
		installs := ""
		if s.Installs > 0 {
			installs = fmt.Sprintf(" (%d installs)", s.Installs)
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] %s%s\n   %s\n   ID: %s\n\n",
			i+1, s.Source, s.Name, installs,
			truncate(s.Description, 100), s.Identifier))
	}
	return sb.String(), nil
}

// --- skill_install ---

type SkillInstallTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillInstallTool) Name() string { return "skill_install" }
func (t *SkillInstallTool) Description() string {
	return "Install a skill from the marketplace by its identifier."
}
func (t *SkillInstallTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"filesystem", "network"}}
}
func (t *SkillInstallTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_install",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"identifier": map[string]interface{}{
						"type":        "string",
						"description": "Skill identifier (e.g., 'smithery/io.github.exa-labs/exa-mcp-server')",
					},
				},
				"required": []string{"identifier"},
			},
		},
	}
}

func (t *SkillInstallTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	identifier, _ := args["identifier"].(string)
	if identifier == "" {
		return "", fmt.Errorf("skill_install: identifier is required")
	}
	if err := t.hub.Install(identifier); err != nil {
		return fmt.Sprintf("Install failed: %v", err), nil
	}
	return fmt.Sprintf("✅ Skill '%s' installed successfully.", identifier), nil
}

// --- skill_list ---

type SkillListTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillListTool) Name() string { return "skill_list" }
func (t *SkillListTool) Description() string {
	return "List all installed skills."
}
func (t *SkillListTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SkillListTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_list",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *SkillListTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	skills, err := t.hub.ListInstalled()
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}
	if len(skills) == 0 {
		stats := t.hub.Stats()
		return fmt.Sprintf("No skills installed. Total indexed: %v", stats["total_indexed"]), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Installed skills (%d):\n\n", len(skills)))
	for _, s := range skills {
		sb.WriteString(fmt.Sprintf("- %s (source: %s)\n", s.Name, s.Source))
	}
	return sb.String(), nil
}

// --- skill_popular ---

type SkillPopularTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillPopularTool) Name() string { return "skill_popular" }
func (t *SkillPopularTool) Description() string {
	return "Show the most popular skills by install count across all sources."
}
func (t *SkillPopularTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SkillPopularTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_popular",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results (default: 10)",
						"default":     10,
					},
				},
			},
		},
	}
}

func (t *SkillPopularTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := 10
	if l, ok := args["limit"]; ok {
		if f, ok := l.(float64); ok {
			limit = int(f)
		}
	}
	skills, err := t.hub.PopularSkills(limit)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}
	if len(skills) == 0 {
		return "No popular skills data. Run skill_sync first.", nil
	}

	var sb strings.Builder
	sb.WriteString("🔥 Popular Skills:\n\n")
	for i, s := range skills {
		installs := ""
		if s.Installs > 0 {
			installs = fmt.Sprintf(" (%d installs)", s.Installs)
		}
		sb.WriteString(fmt.Sprintf("%d. %s%s\n   %s\n   Source: %s\n\n",
			i+1, s.Name, installs, truncate(s.Description, 80), s.Source))
	}
	return sb.String(), nil
}

// --- skill_sync ---

type SkillSyncTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillSyncTool) Name() string { return "skill_sync" }
func (t *SkillSyncTool) Description() string {
	return "Sync the skill index from online sources (Smithery, GitHub, LobeHub). Updates the local database with the latest skills."
}
func (t *SkillSyncTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"network", "database"}}
}
func (t *SkillSyncTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_sync",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *SkillSyncTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	result, err := t.hub.SyncIndex()
	if err != nil {
		return fmt.Sprintf("Sync error: %v", err), nil
	}
	return result.String(), nil
}

// --- skill_info ---

type SkillInfoTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillInfoTool) Name() string { return "skill_info" }
func (t *SkillInfoTool) Description() string {
	return "Get detailed information about a skill by its identifier or name."
}
func (t *SkillInfoTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SkillInfoTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_info",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"identifier": map[string]interface{}{
						"type":        "string",
						"description": "Skill identifier or name",
					},
				},
				"required": []string{"identifier"},
			},
		},
	}
}

func (t *SkillInfoTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	identifier, _ := args["identifier"].(string)
	if identifier == "" {
		return "", fmt.Errorf("skill_info: identifier is required")
	}

	skill, err := t.hub.FindByIdentifier(identifier)
	if err != nil {
		// Try as name
		skill, err = t.hub.FindByName(identifier)
	}
	if err != nil {
		return fmt.Sprintf("Skill not found: %s", identifier), nil
	}

	out, _ := json.MarshalIndent(skill, "", "  ")
	return string(out), nil
}

// --- skill_remove ---

type SkillRemoveTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillRemoveTool) Name() string { return "skill_remove" }
func (t *SkillRemoveTool) Description() string {
	return "Remove an installed skill."
}
func (t *SkillRemoveTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *SkillRemoveTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_remove",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Skill name to remove",
					},
				},
				"required": []string{"name"},
			},
		},
	}
}

func (t *SkillRemoveTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	if name == "" {
		return "", fmt.Errorf("skill_remove: name is required")
	}
	if err := t.hub.Remove(name); err != nil {
		return fmt.Sprintf("Remove failed: %v", err), nil
	}
	return fmt.Sprintf("✅ Skill '%s' removed.", name), nil
}

// --- skill_stats ---

type SkillStatsTool struct {
	hub *skillhub.OnlineHub
}

func (t *SkillStatsTool) Name() string { return "skill_stats" }
func (t *SkillStatsTool) Description() string {
	return "Show skill hub statistics: total indexed, installed, categories, and sources."
}
func (t *SkillStatsTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}
func (t *SkillStatsTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "skill_stats",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *SkillStatsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	stats := t.hub.Stats()
	out, _ := json.MarshalIndent(stats, "", "  ")
	return string(out), nil
}

// --- helpers ---

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

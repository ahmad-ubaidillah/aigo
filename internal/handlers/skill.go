package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/skills"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type SkillHandler struct {
	registry    *skills.Registry
	marketplace *skills.Marketplace
}

func NewSkillHandler(reg *skills.Registry) *SkillHandler {
	return &SkillHandler{
		registry:    reg,
		marketplace: skills.NewMarketplace(reg),
	}
}

func (h *SkillHandler) CanHandle(intent string) bool {
	return intent == types.IntentSkill
}

func (h *SkillHandler) Execute(ctx context.Context, task *types.Task, _ string) (*types.ToolResult, error) {
	desc := strings.TrimSpace(task.Description)

	if strings.HasPrefix(desc, "list") {
		category := strings.TrimPrefix(desc, "list")
		category = strings.TrimSpace(category)
		return h.listSkills(category)
	}

	if strings.HasPrefix(desc, "search ") {
		query := strings.TrimPrefix(desc, "search ")
		return h.searchSkills(query)
	}

	if strings.HasPrefix(desc, "run ") {
		args := strings.TrimPrefix(desc, "run ")
		parts := strings.SplitN(args, " ", 2)
		if len(parts) < 2 {
			return &types.ToolResult{
				Success: false,
				Error:   "usage: skill run <name> <args>",
			}, nil
		}
		return h.runSkill(parts[0], parts[1])
	}

	if strings.HasPrefix(desc, "add ") {
		spec := strings.TrimPrefix(desc, "add ")
		return h.addSkill(spec)
	}

	if strings.HasPrefix(desc, "install ") {
		name := strings.TrimPrefix(desc, "install ")
		return h.installSkill(name)
	}

	if strings.HasPrefix(desc, "market ") {
		query := strings.TrimPrefix(desc, "market ")
		return h.searchMarketplace(query)
	}

	return &types.ToolResult{
		Success: false,
		Error:   "unknown skill command. Use: list, search <query>, run <name> <args>, add <spec>, install <name>, market <query>",
	}, nil
}

func (h *SkillHandler) listSkills(category string) (*types.ToolResult, error) {
	list, err := h.registry.List(category)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}

	var lines []string
	lines = append(lines, "=== Built-in Skills ===")
	for _, s := range h.marketplace.ListBuiltIn() {
		if category == "" || s.Category == category {
			lines = append(lines, fmt.Sprintf("  %s: %s", s.Name, s.Description))
		}
	}

	lines = append(lines, "\n=== Local Skills ===")
	if len(list) == 0 {
		lines = append(lines, "  (none)")
	} else {
		for _, s := range list {
			lines = append(lines, fmt.Sprintf("  %s: %s", s.Name, s.Description))
		}
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(lines, "\n"),
	}, nil
}

func (h *SkillHandler) searchSkills(query string) (*types.ToolResult, error) {
	list, err := h.registry.Search(query)
	if err != nil {
		return nil, fmt.Errorf("search skills: %w", err)
	}

	if len(list) == 0 {
		return &types.ToolResult{
			Success: true,
			Output:  "No local skills matching '" + query + "'. Try 'market <query>' to search marketplace.",
		}, nil
	}

	var lines []string
	for _, s := range list {
		lines = append(lines, fmt.Sprintf("- %s: %s", s.Name, s.Description))
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(lines, "\n"),
	}, nil
}

func (h *SkillHandler) searchMarketplace(query string) (*types.ToolResult, error) {
	results, err := h.marketplace.Search(query, "built-in", "local", "skillsmp")
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Marketplace search failed: %v", err),
		}, nil
	}

	if len(results) == 0 {
		return &types.ToolResult{
			Success: true,
			Output:  "No skills found in marketplace for '" + query + "'",
		}, nil
	}

	var lines []string
	lines = append(lines, "=== Marketplace Results ===")

	bySource := make(map[string][]skills.MarketplaceSkill)
	for _, s := range results {
		bySource[s.Source] = append(bySource[s.Source], s)
	}

	for source, skills := range bySource {
		lines = append(lines, "\n["+source+"]")
		for _, s := range skills {
			lines = append(lines, fmt.Sprintf("  %s: %s", s.Name, s.Description))
		}
	}

	return &types.ToolResult{
		Success: true,
		Output:  strings.Join(lines, "\n"),
	}, nil
}

func (h *SkillHandler) runSkill(name, args string) (*types.ToolResult, error) {
	result, err := h.registry.Execute(name, args)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.ToolResult{
		Success:  true,
		Output:   result.Output,
		Metadata: result.Metadata,
	}, nil
}

func (h *SkillHandler) installSkill(name string) (*types.ToolResult, error) {
	results, err := h.marketplace.Search(name, "skillsmp")
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Search failed: %v", err),
		}, nil
	}

	if len(results) == 0 {
		return &types.ToolResult{
			Success: false,
			Error:   "Skill not found in marketplace",
		}, nil
	}

	for _, s := range results {
		if strings.ToLower(s.Name) == strings.ToLower(name) {
			err := h.marketplace.Install(s)
			if err != nil {
				return &types.ToolResult{
					Success: false,
					Error:   err.Error(),
				}, nil
			}
			return &types.ToolResult{
				Success: true,
				Output:  fmt.Sprintf("Installed skill: %s", s.Name),
			}, nil
		}
	}

	return &types.ToolResult{
		Success: false,
		Error:   "Skill '" + name + "' not found in marketplace",
	}, nil
}

func (h *SkillHandler) addSkill(spec string) (*types.ToolResult, error) {
	parts := strings.Split(spec, "|")
	if len(parts) < 3 {
		return &types.ToolResult{
			Success: false,
			Error:   "usage: skill add <name>|<description>|<command>|<category>|<tags>",
		}, nil
	}

	name := strings.TrimSpace(parts[0])
	description := strings.TrimSpace(parts[1])
	command := strings.TrimSpace(parts[2])
	category := ""
	tags := ""

	if len(parts) > 3 {
		category = strings.TrimSpace(parts[3])
	}
	if len(parts) > 4 {
		tags = strings.TrimSpace(parts[4])
	}

	err := h.registry.Register(name, description, command, category, tags)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  "Skill '" + name + "' added successfully",
	}, nil
}

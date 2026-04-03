// Package tools provides the core tool system for Aigo's autonomous agent loop.
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// SkillTool provides access to skills stored in ~/.hermes/skills/
type SkillTool struct {
	skillsDir string
}

// NewSkillTool creates a new SkillTool instance.
func NewSkillTool() *SkillTool {
	homeDir, _ := os.UserHomeDir()
	return &SkillTool{
		skillsDir: filepath.Join(homeDir, ".hermes", "skills"),
	}
}

// SkillInfo represents basic information about a skill.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// Name returns the tool name.
func (t *SkillTool) Name() string {
	return "skill"
}

// Description returns the tool description.
func (t *SkillTool) Description() string {
	return "List and load skills from ~/.hermes/skills/. Use 'list' to see available skills, 'load' to get skill content."
}

// Schema returns the JSON schema for the tool's parameters.
func (t *SkillTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"list", "load"},
				"description": "Action to perform: 'list' to list available skills, 'load' to load a specific skill",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the skill to load (required for 'load' action)",
			},
		},
		"required": []string{"action"},
	}
}

// Execute runs the skill tool with the given parameters.
func (t *SkillTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	action, _ := params["action"].(string)
	if action == "" {
		return &types.ToolResult{Success: false, Error: "missing required parameter: action"}, nil
	}

	switch action {
	case "list":
		output, err := t.listSkills()
		if err != nil {
			return &types.ToolResult{Success: false, Error: err.Error()}, nil
		}
		return &types.ToolResult{Success: true, Output: output}, nil
	case "load":
		name, _ := params["name"].(string)
		if name == "" {
			return &types.ToolResult{Success: false, Error: "'name' is required for load action"}, nil
		}
		output, err := t.loadSkill(name)
		if err != nil {
			return &types.ToolResult{Success: false, Error: err.Error()}, nil
		}
		return &types.ToolResult{Success: true, Output: output}, nil
	default:
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("unknown action: %s", action)}, nil
	}
}

// listSkills scans the skills directory and returns a list of available skills.
func (t *SkillTool) listSkills() (string, error) {
	// Check if skills directory exists
	if _, err := os.Stat(t.skillsDir); os.IsNotExist(err) {
		return "No skills directory found at " + t.skillsDir, nil
	}

	var skills []SkillInfo

	// Walk the skills directory
	err := filepath.Walk(t.skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Look for SKILL.md files and other markdown files
		if !info.IsDir() && (info.Name() == "SKILL.md" || strings.HasSuffix(info.Name(), ".md")) {
			skill, err := t.parseSkillFile(path)
			if err != nil {
				return nil // Skip files we can't parse
			}
			if skill != nil {
				skills = append(skills, *skill)
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error scanning skills directory: %w", err)
	}

	if len(skills) == 0 {
		return "No skills found in " + t.skillsDir, nil
	}

	// Format output
	var sb strings.Builder
	sb.WriteString("Available Skills:\n\n")
	for _, skill := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", skill.Name, skill.Description))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d skills\n", len(skills)))

	return sb.String(), nil
}

// loadSkill loads and returns the content of a specific skill.
func (t *SkillTool) loadSkill(name string) (string, error) {
	// Try to find the skill by name
	skillPath, err := t.findSkillPath(name)
	if err != nil {
		return "", err
	}

	// Read the skill file
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return "", fmt.Errorf("error reading skill file: %w", err)
	}

	return string(content), nil
}

// findSkillPath finds the path to a skill by its name.
func (t *SkillTool) findSkillPath(name string) (string, error) {
	// Common skill file names to check
	skillFiles := []string{"SKILL.md", "DESCRIPTION.md"}

	// First, try direct path match
	directPath := filepath.Join(t.skillsDir, name)
	if info, err := os.Stat(directPath); err == nil && !info.IsDir() {
		return directPath, nil
	}

	// Walk the directory to find the skill
	var foundPath string
	err := filepath.Walk(t.skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Check if this is a skill file
		for _, sf := range skillFiles {
			if info.Name() == sf {
				// Parse the file to check the name
				skill, err := t.parseSkillFile(path)
				if err == nil && skill != nil && skill.Name == name {
					foundPath = path
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("error searching for skill: %w", err)
	}

	if foundPath == "" {
		return "", fmt.Errorf("skill not found: %s", name)
	}

	return foundPath, nil
}

// parseSkillFile parses a skill markdown file and extracts metadata.
func (t *SkillTool) parseSkillFile(path string) (*SkillInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse YAML frontmatter
	name, description := t.parseFrontmatter(string(content))
	if name == "" {
		// Use filename as name if no frontmatter
		name = strings.TrimSuffix(filepath.Base(path), ".md")
	}

	// Get relative path for display
	relPath, _ := filepath.Rel(t.skillsDir, path)

	return &SkillInfo{
		Name:        name,
		Description: description,
		Path:        relPath,
	}, nil
}

// parseFrontmatter extracts name and description from YAML frontmatter.
func (t *SkillTool) parseFrontmatter(content string) (name, description string) {
	// Check for frontmatter markers
	if !strings.HasPrefix(content, "---") {
		return "", ""
	}

	// Find the end of frontmatter
	endIndex := strings.Index(content[3:], "\n---")
	if endIndex == -1 {
		return "", ""
	}

	frontmatter := content[3 : endIndex+3]
	lines := strings.Split(frontmatter, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			// Remove quotes if present
			name = strings.Trim(name, "\"'")
		}
		if strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			// Remove quotes if present
			description = strings.Trim(description, "\"'")
		}
	}

	return name, description
}

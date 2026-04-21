// Package personatools implements tool functions for persona management.
package personatools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hermes-v2/aigo/internal/persona"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterPersonaTools registers persona tools in the registry.
func RegisterPersonaTools(reg *tools.Registry, pm *persona.Manager) {
	reg.Register(&PersonaListTool{})
	reg.Register(&PersonaGetTool{manager: pm})
	reg.Register(&PersonaSetTool{manager: pm})
}

// --- persona_list ---
type PersonaListTool struct{}

func (t *PersonaListTool) Name() string        { return "persona_list" }
func (t *PersonaListTool) Description() string  { return "List available preset persona profiles." }
func (t *PersonaListTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *PersonaListTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "persona_list",
			Description: "List all available preset persona profiles with their descriptions.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *PersonaListTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	type presetInfo struct {
		Name        string `json:"name"`
		Role        string `json:"role"`
		Tone        string `json:"tone"`
		Language    string `json:"language"`
		Description string `json:"description"`
	}

	var presets []presetInfo
	for _, name := range persona.ListPresets() {
		p := persona.Presets[name]
		presets = append(presets, presetInfo{
			Name:        name,
			Role:        p.Role,
			Tone:        p.Tone,
			Language:    p.Language,
			Description: p.Description,
		})
	}

	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- persona_get ---
type PersonaGetTool struct {
	manager *persona.Manager
}

func (t *PersonaGetTool) Name() string        { return "persona_get" }
func (t *PersonaGetTool) Description() string  { return "Get the currently active persona profile details." }
func (t *PersonaGetTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *PersonaGetTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "persona_get",
			Description: "Get the current active persona profile including AI name, user name, role, tone, skills, and language.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *PersonaGetTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	profile := t.manager.GetActive()
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- persona_set ---
type PersonaSetTool struct {
	manager *persona.Manager
}

func (t *PersonaSetTool) Name() string { return "persona_set" }
func (t *PersonaSetTool) Description() string {
	return "Switch to a preset persona or update specific profile fields."
}
func (t *PersonaSetTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *PersonaSetTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "persona_set",
			Description: "Switch to a preset persona profile by name, or update specific fields. Provide 'preset' to switch presets, or individual fields to update.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"preset": map[string]string{
						"type":        "string",
						"description": "Preset name to switch to (standard, expert, creative, coach, analyst)",
					},
					"ai_name": map[string]string{
						"type":        "string",
						"description": "Set the AI's name",
					},
					"user_name": map[string]string{
						"type":        "string",
						"description": "Set the user's name",
					},
					"role": map[string]string{
						"type":        "string",
						"description": "Set the AI's role description",
					},
					"tone": map[string]string{
						"type":        "string",
						"description": "Set the communication tone",
					},
					"language": map[string]string{
						"type":        "string",
						"description": "Set the primary language (e.g. 'id', 'en')",
					},
				},
			},
		},
	}
}
func (t *PersonaSetTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	preset, _ := args["preset"].(string)
	aiName, _ := args["ai_name"].(string)
	userName, _ := args["user_name"].(string)
	role, _ := args["role"].(string)
	tone, _ := args["tone"].(string)
	language, _ := args["language"].(string)

	// If preset is specified, apply it first
	if preset != "" {
		if err := t.manager.ApplyPreset(preset); err != nil {
			return "", err
		}
		var msgs []string
		msgs = append(msgs, fmt.Sprintf("👤 Switched to '%s' preset", preset))

		// Apply any additional field overrides
		if aiName != "" {
			t.manager.SetName("ai", aiName)
			msgs = append(msgs, fmt.Sprintf("AI name set to: %s", aiName))
		}
		if userName != "" {
			t.manager.SetName("user", userName)
			msgs = append(msgs, fmt.Sprintf("User name set to: %s", userName))
		}
		if role != "" {
			t.manager.SetRole(role)
			msgs = append(msgs, fmt.Sprintf("Role updated: %s", role))
		}
		if tone != "" {
			t.manager.SetTone(tone)
			msgs = append(msgs, fmt.Sprintf("Tone updated: %s", tone))
		}
		if language != "" {
			profile := t.manager.GetActive()
			profile.Language = language
			t.manager.SetActive(profile)
			msgs = append(msgs, fmt.Sprintf("Language set to: %s", language))
		}

		return strings.Join(msgs, "\n"), nil
	}

	// No preset — update individual fields on current profile
	profile := t.manager.GetActive()
	var msgs []string

	if aiName != "" {
		t.manager.SetName("ai", aiName)
		msgs = append(msgs, fmt.Sprintf("AI name set to: %s", aiName))
	}
	if userName != "" {
		t.manager.SetName("user", userName)
		msgs = append(msgs, fmt.Sprintf("User name set to: %s", userName))
	}
	if role != "" {
		t.manager.SetRole(role)
		msgs = append(msgs, fmt.Sprintf("Role updated: %s", role))
	}
	if tone != "" {
		t.manager.SetTone(tone)
		msgs = append(msgs, fmt.Sprintf("Tone updated: %s", tone))
	}
	if language != "" {
		profile.Language = language
		t.manager.SetActive(profile)
		msgs = append(msgs, fmt.Sprintf("Language set to: %s", language))
	}

	if len(msgs) == 0 {
		return "No changes specified. Use 'preset' to switch presets, or set individual fields (ai_name, user_name, role, tone, language).", nil
	}

	return "👤 " + strings.Join(msgs, "\n"), nil
}

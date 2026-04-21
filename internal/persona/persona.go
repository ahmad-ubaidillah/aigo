// Package persona implements agent personality/profile management.
// Each persona defines the agent's name, tone, role, and skills.
package persona

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Profile defines agent personality.
type Profile struct {
	AIName      string   `json:"ai_name"`
	UserName    string   `json:"user_name"`
	Role        string   `json:"role"`
	Tone        string   `json:"tone"`
	Skills      []string `json:"skills,omitempty"`
	Language    string   `json:"language"`
	IsNew       bool     `json:"is_new"`
	Description string   `json:"description,omitempty"`
}

// Manager handles persona profiles.
type Manager struct {
	profilesDir string
	activeFile  string
}

// Preset profiles
var Presets = map[string]Profile{
	"standard": {
		AIName:      "Aigo",
		Role:        "AI assistant with long-term memory",
		Tone:        "Natural and friendly",
		Language:    "id",
		Description: "A well-rounded AI assistant for everyday tasks",
	},
	"expert": {
		AIName:      "Aigo",
		Role:        "Technical expert and advisor",
		Tone:        "Direct, precise, professional",
		Language:    "id",
		Description: "Specialized in technical topics with deep knowledge",
	},
	"creative": {
		AIName:      "Aigo",
		Role:        "Creative companion",
		Tone:        "Playful, imaginative, encouraging",
		Language:    "id",
		Description: "A creative partner for brainstorming and ideation",
	},
	"coach": {
		AIName:      "Aigo",
		Role:        "Personal coach and motivator",
		Tone:        "Supportive, challenging, growth-oriented",
		Language:    "id",
		Description: "Helps you grow and achieve your goals",
	},
	"analyst": {
		AIName:      "Aigo",
		Role:        "Data analyst and researcher",
		Tone:        "Analytical, thorough, evidence-based",
		Language:    "id",
		Description: "Focuses on data analysis and research",
	},
}

// New creates a new Persona Manager.
func New(baseDir string) *Manager {
	os.MkdirAll(baseDir, 0755)
	return &Manager{
		profilesDir: baseDir,
		activeFile:  filepath.Join(baseDir, "active.json"),
	}
}

// GetActive loads the currently active profile.
func (m *Manager) GetActive() Profile {
	data, err := os.ReadFile(m.activeFile)
	if err != nil {
		// Return default
		profile := Presets["standard"]
		profile.IsNew = true
		return profile
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		profile := Presets["standard"]
		profile.IsNew = true
		return profile
	}
	return profile
}

// SetActive saves the active profile.
func (m *Manager) SetActive(profile Profile) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.activeFile, data, 0644)
}

// ApplyPreset loads a preset profile.
func (m *Manager) ApplyPreset(name string) error {
	preset, ok := Presets[name]
	if !ok {
		available := make([]string, 0, len(Presets))
		for k := range Presets {
			available = append(available, k)
		}
		return fmt.Errorf("unknown preset '%s'. Available: %v", name, available)
	}
	return m.SetActive(preset)
}

// SetName sets the AI or user name.
func (m *Manager) SetName(target string, name string) error {
	profile := m.GetActive()
	profile.IsNew = false

	switch target {
	case "ai":
		profile.AIName = name
	case "user":
		profile.UserName = name
	default:
		return fmt.Errorf("invalid target '%s'. Use 'ai' or 'user'", target)
	}

	return m.SetActive(profile)
}

// SetRole updates the agent's role description.
func (m *Manager) SetRole(role string) error {
	profile := m.GetActive()
	profile.Role = role
	return m.SetActive(profile)
}

// SetTone updates the communication tone.
func (m *Manager) SetTone(tone string) error {
	profile := m.GetActive()
	profile.Tone = tone
	return m.SetActive(profile)
}

// BuildSystemPrompt generates a system prompt from the active profile.
func (m *Manager) BuildSystemPrompt() string {
	profile := m.GetActive()

	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf(`You are %s, %s.
Communication style: %s.
`, profile.AIName, profile.Role, profile.Tone))

	if profile.UserName != "" {
		sb.WriteString(fmt.Sprintf("The user's name is %s.\n", profile.UserName))
	}

	if profile.Language != "" {
		sb.WriteString(fmt.Sprintf("Respond primarily in %s.\n", profile.Language))
	}

	if len(profile.Skills) > 0 {
		sb.WriteString(fmt.Sprintf("Active skills: %s\n", strings.Join(profile.Skills, ", ")))
	}

	return sb.String()
}

// ListPresets returns available preset names.
func ListPresets() []string {
	names := make([]string, 0, len(Presets))
	for k := range Presets {
		names = append(names, k)
	}
	return names
}

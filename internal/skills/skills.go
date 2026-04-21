// Package skills implements OpenClaw-compatible skill loading for Aigo.
// Skills are SKILL.md files with YAML frontmatter + markdown body.
// They inject domain knowledge and workflows into the agent's context.
package skills

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Skill represents a loaded skill.
type Skill struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Category    string            `json:"category"`
	Body        string            `json:"body"`     // The full markdown content
	Path        string            `json:"path"`     // File path
	Metadata    map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Manager loads and manages skills.
type Manager struct {
	skillsDir string
	skills    map[string]*Skill
	mu        sync.RWMutex
}

// NewManager creates a new skill manager.
func NewManager(skillsDir string) *Manager {
	return &Manager{
		skillsDir: skillsDir,
		skills:    make(map[string]*Skill),
	}
}

// LoadAll scans and loads all SKILL.md files from the skills directory.
func (m *Manager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(m.skillsDir); os.IsNotExist(err) {
		os.MkdirAll(m.skillsDir, 0755)
		log.Printf("Skills directory created: %s", m.skillsDir)
		return nil
	}

	count := 0
	err := filepath.Walk(m.skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() != "SKILL.md" {
			return nil
		}

		skill, err := m.loadSkillFile(path)
		if err != nil {
			log.Printf("Warning: failed to load skill %s: %v", path, err)
			return nil
		}

		m.skills[skill.Name] = skill
		count++
		return nil
	})

	log.Printf("📚 Loaded %d skills from %s", count, m.skillsDir)
	return err
}

// loadSkillFile parses a SKILL.md file with YAML frontmatter.
func (m *Manager) loadSkillFile(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Parse YAML frontmatter
	var frontmatter map[string]interface{}
	var body string

	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Parse YAML
			if err := yaml.Unmarshal([]byte(parts[1]), &frontmatter); err != nil {
				return nil, fmt.Errorf("parse frontmatter: %w", err)
			}
			body = strings.TrimSpace(parts[2])
		}
	}

	if frontmatter == nil {
		// No frontmatter — use filename as name
		frontmatter = map[string]interface{}{
			"name":        filepath.Base(filepath.Dir(path)),
			"description": "No description",
		}
		body = content
	}

	name, _ := frontmatter["name"].(string)
	if name == "" {
		name = filepath.Base(filepath.Dir(path))
	}

	desc, _ := frontmatter["description"].(string)
	if len(desc) > 200 {
		desc = desc[:200] + "..."
	}

	// Determine category from parent directory
	category := filepath.Base(filepath.Dir(path))

	return &Skill{
		Name:        name,
		Description: desc,
		Category:    category,
		Body:        body,
		Path:        path,
		Metadata:    frontmatter,
	}, nil
}

// Get returns a skill by name.
func (m *Manager) Get(name string) *Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skills[name]
}

// List returns all skills sorted by name.
func (m *Manager) List() []Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skills := make([]Skill, 0, len(m.skills))
	for _, s := range m.skills {
		skills = append(skills, *s)
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills
}

// Search finds skills matching a query (name or description).
func (m *Manager) Search(query string) []Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queryLower := strings.ToLower(query)
	var results []Skill

	for _, s := range m.skills {
		if strings.Contains(strings.ToLower(s.Name), queryLower) ||
			strings.Contains(strings.ToLower(s.Description), queryLower) ||
			strings.Contains(strings.ToLower(s.Category), queryLower) {
			results = append(results, *s)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// InjectContext returns relevant skills as context for the system prompt.
// Only skills matching the query are injected (to save tokens).
func (m *Manager) InjectContext(query string, maxSkills int) string {
	matched := m.Search(query)
	if len(matched) == 0 {
		return ""
	}

	if maxSkills <= 0 {
		maxSkills = 3
	}
	if len(matched) > maxSkills {
		matched = matched[:maxSkills]
	}

	var parts []string
	parts = append(parts, "## Loaded Skills")

	for _, s := range matched {
		// Truncate body to save tokens
		body := s.Body
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		parts = append(parts, fmt.Sprintf("### %s\n%s\n%s", s.Name, s.Description, body))
	}

	return strings.Join(parts, "\n\n")
}

// Count returns the number of loaded skills.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.skills)
}

// InstallFromText creates a new skill from text content.
func (m *Manager) InstallFromText(name, content string) error {
	skillDir := filepath.Join(m.skillsDir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		return err
	}

	// Load the skill
	skill, err := m.loadSkillFile(skillPath)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.skills[skill.Name] = skill
	m.mu.Unlock()

	log.Printf("📥 Installed skill: %s", name)
	return nil
}

// parseTriggerPatterns extracts trigger phrases from skill description.
func parseTriggerPatterns(description string) []string {
	// Common trigger patterns in OpenClaw skills
	re := regexp.MustCompile(`[Uu]se (?:this skill )?when[:\.]?\s*(.+?)(?:\.|$)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		return strings.Split(matches[1], ", ")
	}
	return nil
}

package project

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Project struct {
	ID        string                 `json:"id"`
	Path      string                 `json:"path"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
	Context   string                 `json:"context"`
}

type ProjectStore struct {
	projects map[string]*Project
	basePath string
}

func New(basePath string) (*ProjectStore, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "projects")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	s := &ProjectStore{
		projects: make(map[string]*Project),
		basePath: basePath,
	}

	s.loadAll()
	return s, nil
}

func (s *ProjectStore) DetectProject(cwd string) (*Project, error) {
	cwd, err := filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}

	// Check if already known
	id := s.projectID(cwd)
	if p, ok := s.projects[id]; ok {
		p.UpdatedAt = time.Now()
		return p, nil
	}

	// Detect project type and create new
	projType, name := detectProjectType(cwd)

	project := &Project{
		ID:        id,
		Path:      cwd,
		Name:      name,
		Type:      projType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
		Context:   "",
	}

	s.projects[id] = project
	s.save(project)

	return project, nil
}

func detectProjectType(cwd string) (string, string) {
	name := filepath.Base(cwd)

	// Check for common project types
	indicators := []struct {
		file     string
		projType string
	}{
		{"go.mod", "go"},
		{"package.json", "javascript"},
		{"Cargo.toml", "rust"},
		{"pyproject.toml", "python"},
		{"pom.xml", "java"},
		{"build.gradle", "kotlin"},
		{"composer.json", "php"},
		{"Gemfile", "ruby"},
		{"Cargo.toml", "rust"},
		{"main.rs", "rust"},
		{"src/main.go", "go"},
		{"index.js", "javascript"},
		{"tsconfig.json", "typescript"},
		{"next.config.js", "nextjs"},
		{"vite.config.ts", "vite"},
		{"webpack.config.js", "webpack"},
		{"Makefile", "make"},
		{"CMakeLists.txt", "cmake"},
	}

	for _, ind := range indicators {
		if _, err := os.Stat(filepath.Join(cwd, ind.file)); err == nil {
			return ind.projType, name
		}
	}

	// Check for git
	if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
		return "git", name
	}

	return "unknown", name
}

func (s *ProjectStore) projectID(path string) string {
	hash := md5.Sum([]byte(path))
	return hex.EncodeToString(hash[:])
}

func (s *ProjectStore) UpdateContext(projectPath, context string) error {
	id := s.projectID(projectPath)
	if p, ok := s.projects[id]; ok {
		p.Context = context
		p.UpdatedAt = time.Now()
		return s.save(p)
	}
	return fmt.Errorf("project not found: %s", projectPath)
}

func (s *ProjectStore) GetContext(projectPath string) string {
	id := s.projectID(projectPath)
	if p, ok := s.projects[id]; ok {
		return p.Context
	}
	return ""
}

func (s *ProjectStore) AddMetadata(projectPath string, key string, value interface{}) error {
	id := s.projectID(projectPath)
	if p, ok := s.projects[id]; ok {
		if p.Metadata == nil {
			p.Metadata = make(map[string]interface{})
		}
		p.Metadata[key] = value
		p.UpdatedAt = time.Now()
		return s.save(p)
	}
	return fmt.Errorf("project not found: %s", projectPath)
}

func (s *ProjectStore) GetMetadata(projectPath, key string) interface{} {
	id := s.projectID(projectPath)
	if p, ok := s.projects[id]; ok {
		return p.Metadata[key]
	}
	return nil
}

func (s *ProjectStore) List() []*Project {
	var list []*Project
	for _, p := range s.projects {
		list = append(list, p)
	}
	return list
}

func (s *ProjectStore) Get(projectPath string) *Project {
	id := s.projectID(projectPath)
	return s.projects[id]
}

func (s *ProjectStore) Remember(projectPath, content string) error {
	id := s.projectID(projectPath)
	if p, ok := s.projects[id]; ok {
		if p.Context != "" {
			p.Context = p.Context + "\n" + content
		} else {
			p.Context = content
		}
		p.UpdatedAt = time.Now()
		return s.save(p)
	}
	return fmt.Errorf("project not found: %s", projectPath)
}

func (s *ProjectStore) Search(projectPath, query string) []string {
	p := s.Get(projectPath)
	if p == nil || p.Context == "" {
		return nil
	}

	var results []string
	lines := strings.Split(p.Context, "\n")
	queryLower := strings.ToLower(query)

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), queryLower) {
			results = append(results, line)
		}
	}

	return results
}

func (s *ProjectStore) save(p *Project) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	projDir := filepath.Join(s.basePath, p.ID)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projDir, "project.json"), data, 0644)
}

func (s *ProjectStore) loadAll() {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.basePath, e.Name(), "project.json"))
		if err != nil {
			continue
		}

		var p Project
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}

		s.projects[p.ID] = &p
	}
}

func (s *ProjectStore) Remove(projectPath string) error {
	id := s.projectID(projectPath)
	if _, ok := s.projects[id]; !ok {
		return fmt.Errorf("project not found: %s", projectPath)
	}

	delete(s.projects, id)
	return os.RemoveAll(filepath.Join(s.basePath, id))
}

func (s *ProjectStore) Stats() map[string]int {
	stats := make(map[string]int)
	stats["total"] = len(s.projects)

	typeCounts := make(map[string]int)
	for _, p := range s.projects {
		typeCounts[p.Type]++
	}

	for t, c := range typeCounts {
		stats["type_"+t] = c
	}

	return stats
}
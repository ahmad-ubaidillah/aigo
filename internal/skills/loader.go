package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type SkillInfo struct {
	Name        string
	Path        string
	Description string
	Category    string
	Version     string
	Enabled     bool
}

type FileSkillLoader struct {
	basePath string
	skills   map[string]*SkillInfo
	mu       sync.RWMutex
}

func NewFileSkillLoader(basePath string) *FileSkillLoader {
	return &FileSkillLoader{
		basePath: basePath,
		skills:   make(map[string]*SkillInfo),
	}
}

func (l *FileSkillLoader) Discover() error {
	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read skills dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(l.basePath, entry.Name())
		info, err := l.loadSkillDir(skillPath, entry.Name())
		if err != nil {
			continue
		}
		l.mu.Lock()
		l.skills[info.Name] = info
		l.mu.Unlock()
	}
	return nil
}

func (l *FileSkillLoader) loadSkillDir(dir, name string) (*SkillInfo, error) {
	skillFile := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		return nil, fmt.Errorf("no SKILL.md in %s", dir)
	}

	content, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, fmt.Errorf("read SKILL.md: %w", err)
	}

	info := &SkillInfo{
		Name:    name,
		Path:    dir,
		Enabled: true,
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			info.Description = strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "category:") {
			info.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
		}
		if strings.HasPrefix(line, "version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
		}
	}

	return info, nil
}

func (l *FileSkillLoader) Get(name string) (*SkillInfo, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	s, ok := l.skills[name]
	if !ok {
		return nil, fmt.Errorf("skill %s not found", name)
	}
	return s, nil
}

func (l *FileSkillLoader) List() []SkillInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]SkillInfo, 0, len(l.skills))
	for _, s := range l.skills {
		result = append(result, *s)
	}
	return result
}

func (l *FileSkillLoader) ReadContent(name string) (string, error) {
	info, err := l.Get(name)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(filepath.Join(info.Path, "SKILL.md"))
	if err != nil {
		return "", fmt.Errorf("read skill content: %w", err)
	}
	return string(content), nil
}

// Package memory implements the memory subsystem.
// 3-tier: Session → Daily → Long-term (inspired by ClawHive + Zeph).
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Store manages persistent memory.
type Store struct {
	basePath string
}

// New creates a new memory store.
func New(basePath string) (*Store, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create memory dir: %w", err)
	}
	// Create subdirectories
	for _, dir := range []string{"daily", "longterm", "sessions"} {
		os.MkdirAll(filepath.Join(basePath, dir), 0755)
	}
	return &Store{basePath: basePath}, nil
}

// SaveDaily saves a daily memory note.
func (s *Store) SaveDaily(content string) error {
	today := time.Now().Format("2006-01-02")
	path := filepath.Join(s.basePath, "daily", today+".md")

	// Append to existing file
	existing, _ := os.ReadFile(path)
	timestamp := time.Now().Format("15:04")
	entry := fmt.Sprintf("\n## %s\n%s\n", timestamp, content)

	return os.WriteFile(path, append(existing, []byte(entry)...), 0644)
}

// SaveLongTerm saves a long-term memory note.
func (s *Store) SaveLongTerm(content string) error {
	path := filepath.Join(s.basePath, "longterm", "MEMORY.md")
	existing, _ := os.ReadFile(path)
	timestamp := time.Now().Format("2006-01-02 15:04")
	entry := fmt.Sprintf("\n### %s\n%s\n", timestamp, content)
	return os.WriteFile(path, append(existing, []byte(entry)...), 0644)
}

// Search searches across all memory files.
func (s *Store) Search(query string, limit int) ([]string, error) {
	var results []string
	queryLower := strings.ToLower(query)

	dirs := []string{"daily", "longterm"}
	for _, dir := range dirs {
		dirPath := filepath.Join(s.basePath, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".md") && !strings.HasSuffix(entry.Name(), ".jsonl")) {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				continue
			}
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), queryLower) {
					results = append(results, fmt.Sprintf("[%s/%s] %s", dir, entry.Name(), strings.TrimSpace(line)))
					if len(results) >= limit {
						return results, nil
					}
				}
			}
		}
	}
	return results, nil
}

// GetDaily returns today's daily memory.
func (s *Store) GetDaily() string {
	today := time.Now().Format("2006-01-02")
	path := filepath.Join(s.basePath, "daily", today+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// GetLongTerm returns all long-term memory.
func (s *Store) GetLongTerm() string {
	path := filepath.Join(s.basePath, "longterm", "MEMORY.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// SaveSession saves a conversation session.
func (s *Store) SaveSession(sessionID string, messages []map[string]interface{}) error {
	path := filepath.Join(s.basePath, "sessions", sessionID+".jsonl")
	var lines []string
	for _, msg := range messages {
		b, _ := json.Marshal(msg)
		lines = append(lines, string(b))
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// Context returns memory context to inject into the agent prompt.
func (s *Store) Context() string {
	var parts []string

	longTerm := s.GetLongTerm()
	if longTerm != "" {
		parts = append(parts, "## Long-term Memory\n"+longTerm)
	}

	daily := s.GetDaily()
	if daily != "" {
		parts = append(parts, "## Today's Notes\n"+daily)
	}

	if len(parts) == 0 {
		return ""
	}
	return "<memory>\n" + strings.Join(parts, "\n\n") + "\n</memory>"
}

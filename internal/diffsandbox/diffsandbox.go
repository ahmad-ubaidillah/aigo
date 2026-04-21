package diffsandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Change struct {
	ID        string    `json:"id"`
	File     string    `json:"file"`
	OldContent string `json:"old_content,omitempty"`
	NewContent string `json:"new_content"`
	Patch    string    `json:"patch,omitempty"`
	Status   string    `json:"status"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

type Sandbox struct {
	changes  map[string][]Change
	basePath string
}

func New(basePath string) (*Sandbox, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "sandbox")
	}
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &Sandbox{
		changes:  make(map[string][]Change),
		basePath: basePath,
	}, nil
}

func (s *Sandbox) Add(project, file, oldContent, newContent string) (string, error) {
	id := generateID(file)
	change := Change{
		ID:          id,
		File:       file,
		OldContent: oldContent,
		NewContent: newContent,
		Status:     "pending",
	}
	s.changes[project] = append(s.changes[project], change)
	return id, nil
}

func (s *Sandbox) List(project string) []Change {
	return s.changes[project]
}

func (s *Sandbox) Get(project, id string) *Change {
	for _, c := range s.changes[project] {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

func (s *Sandbox) Apply(project, id string) error {
	for i := range s.changes[project] {
		if s.changes[project][i].ID == id {
			if err := os.WriteFile(s.changes[project][i].File, []byte(s.changes[project][i].NewContent), 0644); err != nil {
				return err
			}
			now := time.Now()
			s.changes[project][i].AppliedAt = &now
			s.changes[project][i].Status = "applied"
			return nil
		}
	}
	return fmt.Errorf("change not found")
}

func (s *Sandbox) ApplyAll(project string) error {
	for i := range s.changes[project] {
		if s.changes[project][i].Status != "pending" {
			continue
		}
		if err := os.WriteFile(s.changes[project][i].File, []byte(s.changes[project][i].NewContent), 0644); err != nil {
			return err
		}
		now := time.Now()
		s.changes[project][i].AppliedAt = &now
		s.changes[project][i].Status = "applied"
	}
	return nil
}

func (s *Sandbox) Reject(project, id string) error {
	for i := range s.changes[project] {
		if s.changes[project][i].ID == id {
			s.changes[project][i].Status = "rejected"
			return nil
		}
	}
	return fmt.Errorf("change not found")
}

func (s *Sandbox) RejectAll(project string) {
	for i := range s.changes[project] {
		s.changes[project][i].Status = "rejected"
	}
}

func (s *Sandbox) Diff(project, id string) (string, error) {
	change := s.Get(project, id)
	if change == nil {
		return "", fmt.Errorf("change not found")
	}
	return fmt.Sprintf("--- %s\n+++ %s\n", change.OldContent, change.NewContent), nil
}

func (s *Sandbox) Pending(project string) []Change {
	var pending []Change
	for _, c := range s.changes[project] {
		if c.Status == "pending" {
			pending = append(pending, c)
		}
	}
	return pending
}

func generateID(seed string) string {
	data := []byte(seed + time.Now().Format("20060102150405"))
	sum := 0
	for _, b := range data {
		sum += int(b)
	}
	return fmt.Sprintf("%x", sum)
}

func (s *Sandbox) Stats() map[string]int {
	stats := make(map[string]int)
	for project, changes := range s.changes {
		stats[project] = len(changes)
	}
	return stats
}
// Package evolution implements the self-evolution system for Aigo.
// It allows the agent to propose, apply, and revert code changes safely.
package evolution

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Proposal represents a proposed code change.
type Proposal struct {
	ID         string    `json:"id"`
	File       string    `json:"file"`
	Find       string    `json:"find,omitempty"`
	Replace    string    `json:"replace,omitempty"`
	NewContent string    `json:"new_content,omitempty"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"` // "pending", "applied", "reverted", "failed"
	Created    time.Time `json:"created"`
}

// Manager manages evolution proposals and changes.
type Manager struct {
	mu             sync.Mutex
	proposals      []*Proposal
	history        []Proposal
	projectDir     string
	maxPerSession  int
	sessionCount   int
}

// New creates a new evolution manager.
func New(projectDir string) *Manager {
	// Clean and resolve the project directory
	projectDir = filepath.Clean(projectDir)
	return &Manager{
		proposals:     make([]*Proposal, 0),
		history:       make([]Proposal, 0),
		projectDir:    projectDir,
		maxPerSession: 3,
	}
}

// ProjectDir returns the project directory path.
func (m *Manager) ProjectDir() string {
	return m.projectDir
}

// validateFile checks that the file is safe to modify.
func (m *Manager) validateFile(file string) error {
	// Resolve to absolute path
	absPath := file
	if !filepath.IsAbs(file) {
		absPath = filepath.Join(m.projectDir, file)
	}
	absPath = filepath.Clean(absPath)

	// Must be within project directory
	rel, err := filepath.Rel(m.projectDir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("file %s is outside project directory %s", file, m.projectDir)
	}

	// Must be a .go file
	if !strings.HasSuffix(absPath, ".go") {
		return fmt.Errorf("can only modify .go files, got: %s", filepath.Ext(absPath))
	}

	// Never modify config or credential files
	base := filepath.Base(absPath)
	lower := strings.ToLower(base)
	blocked := []string{"config", "credential", "secret", "token", "password", ".env"}
	for _, b := range blocked {
		if strings.Contains(lower, b) {
			return fmt.Errorf("cannot modify file containing '%s': %s", b, base)
		}
	}

	return nil
}

func genID() string {
	return fmt.Sprintf("evo-%d", time.Now().UnixNano())
}

// Propose creates a new code change proposal.
func (m *Manager) Propose(file, find, replace, newContent, reason string) (*Proposal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sessionCount >= m.maxPerSession {
		return nil, fmt.Errorf("evolution limit reached: max %d per session", m.maxPerSession)
	}

	// Validate file
	if err := m.validateFile(file); err != nil {
		return nil, err
	}

	// Must have either find+replace or newContent
	if newContent == "" && (find == "" || replace == "") {
		return nil, fmt.Errorf("must provide either new_content or both find and replace")
	}

	// Resolve absolute path
	absPath := file
	if !filepath.IsAbs(file) {
		absPath = filepath.Join(m.projectDir, file)
	}

	p := &Proposal{
		ID:         genID(),
		File:       absPath,
		Find:       find,
		Replace:    replace,
		NewContent: newContent,
		Reason:     reason,
		Status:     "pending",
		Created:    time.Now(),
	}

	m.proposals = append(m.proposals, p)
	log.Printf("🔧 Evolution proposal created: %s — %s", p.ID, truncate(reason, 60))
	return p, nil
}

// Apply applies a proposal by ID. It creates a .bak backup, applies the change,
// runs `go build` to verify, and auto-reverts if the build fails.
func (m *Manager) Apply(id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find proposal
	var proposal *Proposal
	idx := -1
	for i, p := range m.proposals {
		if p.ID == id {
			proposal = p
			idx = i
			break
		}
	}
	if proposal == nil {
		return "", fmt.Errorf("proposal not found: %s", id)
	}
	if proposal.Status != "pending" {
		return "", fmt.Errorf("proposal %s already %s", id, proposal.Status)
	}

	// Read original file
	original, err := os.ReadFile(proposal.File)
	if err != nil {
		proposal.Status = "failed"
		return "", fmt.Errorf("read file: %w", err)
	}

	// Create backup
	bakPath := proposal.File + ".bak"
	if err := os.WriteFile(bakPath, original, 0644); err != nil {
		proposal.Status = "failed"
		return "", fmt.Errorf("create backup: %w", err)
	}

	// Compute new content
	var newFileContent []byte
	if proposal.NewContent != "" {
		newFileContent = []byte(proposal.NewContent)
	} else {
		// Find and replace
		content := string(original)
		if !strings.Contains(content, proposal.Find) {
			proposal.Status = "failed"
			return "", fmt.Errorf("find string not found in %s", proposal.File)
		}
		newContent := strings.ReplaceAll(content, proposal.Find, proposal.Replace)
		newFileContent = []byte(newContent)
	}

	// Write new content
	if err := os.WriteFile(proposal.File, newFileContent, 0644); err != nil {
		proposal.Status = "failed"
		return "", fmt.Errorf("write file: %w", err)
	}

	// Test build
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", "./cmd/aigo/")
	buildCmd.Dir = m.projectDir
	buildCmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/go/bin")
	output, err := buildCmd.CombinedOutput()

	if err != nil {
		// Auto-revert
		log.Printf("🔧 Build failed after applying %s, auto-reverting: %s", id, string(output))
		if revertErr := os.WriteFile(proposal.File, original, 0644); revertErr != nil {
			proposal.Status = "failed"
			return "", fmt.Errorf("build failed AND revert failed: build: %v, revert: %v", err, revertErr)
		}
		proposal.Status = "failed"
		return "", fmt.Errorf("build failed, auto-reverted: %s", string(output))
	}

	proposal.Status = "applied"
	m.sessionCount++

	// Move to history
	historyEntry := *proposal
	m.history = append(m.history, historyEntry)

	// Remove from active proposals
	m.proposals = append(m.proposals[:idx], m.proposals[idx+1:]...)

	log.Printf("🔧 Evolution applied: %s — %s", id, truncate(proposal.Reason, 60))
	return fmt.Sprintf("✅ Applied %s. Build passed. Backup at %s", id, bakPath), nil
}

// Revert restores a file from its .bak backup.
func (m *Manager) Revert(id string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find in history
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].ID == id && m.history[i].Status == "applied" {
			entry := &m.history[i]
			bakPath := entry.File + ".bak"

			backup, err := os.ReadFile(bakPath)
			if err != nil {
				return "", fmt.Errorf("backup not found: %s", bakPath)
			}

			if err := os.WriteFile(entry.File, backup, 0644); err != nil {
				return "", fmt.Errorf("restore failed: %w", err)
			}

			entry.Status = "reverted"
			log.Printf("🔧 Evolution reverted: %s", id)
			return fmt.Sprintf("✅ Reverted %s from %s", id, bakPath), nil
		}
	}

	return "", fmt.Errorf("no applied proposal found with id: %s", id)
}

// History returns all proposals (active + completed).
func (m *Manager) History() []Proposal {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]Proposal, 0, len(m.proposals)+len(m.history))
	for _, p := range m.proposals {
		result = append(result, *p)
	}
	result = append(result, m.history...)
	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

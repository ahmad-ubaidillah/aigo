// Package engram provides a memory backend backed by Engram's structured store.
// It wraps engram_store.Store to implement memory.Backend, adding structured
// observations, topic key deduplication, session lifecycle, and FTS5 search.
package engram

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hermes-v2/aigo/internal/memory/engram_store"
)

// Backend wraps engram's store to implement memory.Backend.
type Backend struct {
	store       *engram_store.Store
	project     string
	sessionID   string
	sessionDir  string
}

// New creates a new engram-backed memory backend.
func New(dataDir string, project string) (*Backend, error) {
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("engram: home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".aigo", "engram")
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("engram: mkdir: %w", err)
	}

	cfg := engram_store.FallbackConfig(dataDir)
	s, err := engram_store.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("engram: open store: %w", err)
	}

	if project == "" {
		project = "aigo"
	}

	b := &Backend{
		store:   s,
		project: project,
	}

	return b, nil
}

// Close closes the underlying store.
func (b *Backend) Close() error {
	return b.store.Close()
}

// --- Session lifecycle ---

// StartSession creates a new engram session and stores the session ID.
func (b *Backend) StartSession(sessionID string) error {
	if sessionID == "" {
		sessionID = fmt.Sprintf("aigo-%s", time.Now().Format("20060102-150405"))
	}
	dir, _ := os.Getwd()
	if err := b.store.CreateSession(sessionID, b.project, dir); err != nil {
		return fmt.Errorf("engram: create session: %w", err)
	}
	b.sessionID = sessionID
	b.sessionDir = dir
	return nil
}

// EndSession ends the current engram session with a summary.
func (b *Backend) EndSession(summary string) error {
	if b.sessionID == "" {
		return nil
	}
	return b.store.EndSession(b.sessionID, summary)
}

// SessionID returns the current session ID.
func (b *Backend) SessionID() string {
	return b.sessionID
}

// --- Observations ---

// SaveObservation saves a structured observation to engram.
func (b *Backend) SaveObservation(obsType, title, content, topicKey string) (int64, error) {
	if b.sessionID == "" {
		if err := b.StartSession(""); err != nil {
			return 0, err
		}
	}
	return b.store.AddObservation(engram_store.AddObservationParams{
		SessionID: b.sessionID,
		Type:      obsType,
		Title:     title,
		Content:   content,
		Project:   b.project,
		Scope:     "project",
		TopicKey:  topicKey,
	})
}

// SaveObservationFull saves with all fields.
func (b *Backend) SaveObservationFull(sessionID, obsType, title, content, toolName, scope, topicKey string) (int64, error) {
	if sessionID == "" {
		sessionID = b.sessionID
	}
	if sessionID == "" {
		if err := b.StartSession(""); err != nil {
			return 0, err
		}
		sessionID = b.sessionID
	}
	return b.store.AddObservation(engram_store.AddObservationParams{
		SessionID: sessionID,
		Type:      obsType,
		Title:     title,
		Content:   content,
		ToolName:  toolName,
		Project:   b.project,
		Scope:     scope,
		TopicKey:  topicKey,
	})
}

// --- Search ---

// Search performs FTS5 search and adapts results to memory.SearchResult.
// This implements memory.Backend.Search().
func (b *Backend) Search(query string, limit int) ([]engram_store.SearchResult, error) {
	return b.store.Search(query, engram_store.SearchOptions{
		Project: b.project,
		Limit:   limit,
	})
}

// --- Context ---

// GetContext returns formatted context from previous sessions.
func (b *Backend) GetContext() (string, error) {
	return b.store.FormatContext(b.project, "project")
}

// GetRecentObservations returns the most recent observations.
func (b *Backend) GetRecentObservations(limit int) ([]engram_store.Observation, error) {
	return b.store.RecentObservations(b.project, "project", limit)
}

// --- Timeline ---

// GetTimeline returns chronological context around an observation.
func (b *Backend) GetTimeline(obsID int64, before, after int) (*engram_store.TimelineResult, error) {
	return b.store.Timeline(obsID, before, after)
}

// --- Stats ---

// GetStats returns memory statistics.
func (b *Backend) GetStats() (*engram_store.Stats, error) {
	return b.store.Stats()
}

// --- Backend interface (for compatibility with memory.Backend) ---

// Context returns memory context to inject into the agent system prompt.
// This implements memory.Backend.Context().
func (b *Backend) Context() string {
	ctx, err := b.GetContext()
	if err != nil {
		return ""
	}
	if ctx == "" {
		return ""
	}
	return "<memory>\n" + ctx + "\n</memory>"
}

// SaveDaily saves content as a "daily" observation.
// This implements memory.Backend.SaveDaily().
func (b *Backend) SaveDaily(content string) error {
	_, err := b.SaveObservation("daily", "Daily Note", content, "")
	return err
}

// SaveLongTerm saves content as a "longterm" observation.
// This implements memory.Backend.SaveLongTerm().
func (b *Backend) SaveLongTerm(content string) error {
	_, err := b.SaveObservation("longterm", "Long-term Note", content, "")
	return err
}

// SearchResults adapts engram search results to memory.SearchResult.
// This implements memory.Backend.Search().
func (b *Backend) SearchResults(query string, limit int) ([]SearchResult, error) {
	results, err := b.Search(query, limit)
	if err != nil {
		return nil, err
	}
	var out []SearchResult
	for _, r := range results {
		out = append(out, SearchResult{
			Content:  fmt.Sprintf("[%s] %s: %s", r.Type, r.Title, truncate(r.Content, 300)),
			Category: r.Type,
			Score:    r.Rank,
		})
	}
	return out, nil
}

// SearchResult is a simplified search result for the Backend interface.
type SearchResult struct {
	Content  string
	Category string
	Score    float64
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// --- Prompt logging ---

// SavePrompt saves a user prompt to engram.
func (b *Backend) SavePrompt(content string) error {
	if b.sessionID == "" {
		if err := b.StartSession(""); err != nil {
			return err
		}
	}
	_, err := b.store.AddPrompt(engram_store.AddPromptParams{
		SessionID: b.sessionID,
		Content:   content,
		Project:   b.project,
	})
	return err
}

// --- Formatting helpers ---

// FormatObservation formats an observation as a string.
func FormatObservation(obs *engram_store.Observation) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s] %s", obs.Type, obs.Title))
	if obs.Content != "" {
		content := obs.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		parts = append(parts, content)
	}
	if obs.TopicKey != nil && *obs.TopicKey != "" {
		parts = append(parts, fmt.Sprintf("topic: %s", *obs.TopicKey))
	}
	return strings.Join(parts, "\n")
}

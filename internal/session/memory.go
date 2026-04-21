// Package session implements automatic session memory.
// Every conversation turn is saved to FTS5 for future recall.
// Corrections are auto-detected and learned.
package session

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/memory/fts5pkg"
)

// AutoMemory manages automatic conversation memory.
type AutoMemory struct {
	fts5      *fts5pkg.Store
	mu        sync.Mutex
	sessionID string
	turns     []Turn
}

// Turn is a single conversation exchange.
type Turn struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// NewAutoMemory creates a new auto-memory system.
func NewAutoMemory(fts5Store *fts5pkg.Store) *AutoMemory {
	sessionID := fmt.Sprintf("session_%s", time.Now().Format("20060102_150405"))
	return &AutoMemory{
		fts5:      fts5Store,
		sessionID: sessionID,
		turns:     make([]Turn, 0, 50),
	}
}

// AddTurn saves a conversation turn and auto-detects learning opportunities.
func (am *AutoMemory) AddTurn(role, content string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	turn := Turn{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
	am.turns = append(am.turns, turn)

	// Save to FTS5 for future search
	if am.fts5 != nil {
		category := "session"
		if role == "user" {
			am.fts5.Save(fmt.Sprintf("User said: %s", content), category, am.sessionID)
		} else {
			am.fts5.Save(fmt.Sprintf("Aigo responded: %s", content), category, am.sessionID)
		}
	}

	// Auto-detect corrections from user
	if role == "user" {
		am.detectAndLearnCorrection(content)
	}

	// Keep turn buffer manageable
	if len(am.turns) > 100 {
		am.turns = am.turns[len(am.turns)-50:]
	}
}

// detectAndLearnCorrection automatically learns from user corrections.
func (am *AutoMemory) detectAndLearnCorrection(text string) {
	if am.fts5 == nil {
		return
	}

	lower := strings.ToLower(text)

	// Correction patterns
	corrections := []struct {
		pattern  string
		category string
	}{
		{"bukan ", "correction"},
		{"salah ", "correction"},
		{"jangan ", "correction"},
		{"itu salah", "correction"},
		{"tidak benar", "correction"},
		{"seharusnya ", "correction"},
		{"yang benar ", "correction"},
		{"lebih baik ", "preference"},
		{"saya lebih suka ", "preference"},
		{"saya suka ", "preference"},
		{"saya tidak suka ", "preference"},
		{"prefer ", "preference"},
		{"selalu ", "preference"},
		{"ingat bahwa ", "fact"},
		{"catat ", "fact"},
	}

	for _, c := range corrections {
		if strings.Contains(lower, c.pattern) {
			am.fts5.Save(
				fmt.Sprintf("[%s] %s", c.category, text),
				"auto_learned",
				am.sessionID,
			)
			break
		}
	}
}

// GetRecentContext returns the last N turns as context for the system prompt.
func (am *AutoMemory) GetRecentContext(n int) string {
	am.mu.Lock()
	defer am.mu.Unlock()

	if len(am.turns) == 0 {
		return ""
	}

	start := len(am.turns) - n
	if start < 0 {
		start = 0
	}

	var parts []string
	parts = append(parts, "## Recent Conversation Context")
	for _, t := range am.turns[start:] {
		role := "User"
		if t.Role == "assistant" {
			role = "Aigo"
		}
		content := t.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s: %s", role, content))
	}

	return strings.Join(parts, "\n")
}

// SearchRelated searches past sessions for relevant context.
func (am *AutoMemory) SearchRelated(query string, limit int) string {
	if am.fts5 == nil {
		return ""
	}

	entries, err := am.fts5.Search(query, limit)
	if err != nil || len(entries) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Related Past Context")
	for _, e := range entries {
		content := e.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		parts = append(parts, fmt.Sprintf("- %s", content))
	}

	return strings.Join(parts, "\n")
}

// GetSessionSummary returns a summary of the current session.
func (am *AutoMemory) GetSessionSummary() string {
	am.mu.Lock()
	defer am.mu.Unlock()

	if len(am.turns) == 0 {
		return "No conversation yet."
	}

	userTurns := 0
	for _, t := range am.turns {
		if t.Role == "user" {
			userTurns++
		}
	}

	topics := extractTopics(am.turns)

	return fmt.Sprintf("Session: %d exchanges, %d user messages. Topics: %s",
		len(am.turns), userTurns, strings.Join(topics, ", "))
}

// extractTopics extracts key topics from conversation turns.
func extractTopics(turns []Turn) []string {
	wordFreq := make(map[string]int)
	stopWords := map[string]bool{
		"the": true, "a": true, "is": true, "it": true, "to": true,
		"dan": true, "di": true, "ya": true, "yang": true, "ini": true,
		"itu": true, "ada": true, "untuk": true, "dengan": true, "saya": true,
		"kamu": true, "dia": true, "apa": true, "bagaimana": true, "bisa": true,
		"akan": true, "dari": true, "pada": true, "ke": true,
	}

	for _, t := range turns {
		words := strings.Fields(strings.ToLower(t.Content))
		for _, w := range words {
			w = strings.Trim(w, ".,!?;:\"'()[]{}")
			if len(w) > 3 && !stopWords[w] {
				wordFreq[w]++
			}
		}
	}

	// Get top 3 most frequent words as topics
	type wordCount struct {
		word  string
		count int
	}
	var sorted []wordCount
	for w, c := range wordFreq {
		if c >= 2 {
			sorted = append(sorted, wordCount{w, c})
		}
	}

	// Simple sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var topics []string
	for i, s := range sorted {
		if i >= 3 {
			break
		}
		topics = append(topics, s.word)
	}

	if len(topics) == 0 {
		topics = []string{"general"}
	}

	return topics
}

// TurnCount returns the number of turns in the current session.
func (am *AutoMemory) TurnCount() int {
	am.mu.Lock()
	defer am.mu.Unlock()
	return len(am.turns)
}

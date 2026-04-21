// Package fts5pkg implements full-text search memory using SQLite FTS5.
// Pure Go — uses modernc.org/sqlite (no CGO needed for static builds).
package fts5pkg

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Entry is a memory entry with metadata.
type Entry struct {
	ID        int64
	Content   string
	Category  string // "daily", "longterm", "session", "skill"
	SessionID string
	CreatedAt time.Time
	Score     float64 // FTS5 rank (lower = better match)
}

// Store manages FTS5-backed memory.
type Store struct {
	db       *sql.DB
	basePath string
}

// New creates or opens an FTS5 memory store.
func New(basePath string) (*Store, error) {
	dbPath := filepath.Join(basePath, "memory.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Create main table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			category TEXT NOT NULL DEFAULT 'daily',
			session_id TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	// Create FTS5 virtual table
	if _, err := db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
			content,
			category,
			session_id,
			content='memories',
			content_rowid='id',
			tokenize='unicode61'
		)
	`); err != nil {
		return nil, fmt.Errorf("create fts5 table: %w", err)
	}

	// Create triggers to keep FTS5 in sync
	triggers := []string{
		`CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
			INSERT INTO memories_fts(rowid, content, category, session_id)
			VALUES (new.id, new.content, new.category, new.session_id);
		END`,
		`CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
			INSERT INTO memories_fts(memories_fts, rowid, content, category, session_id)
			VALUES ('delete', old.id, old.content, old.category, old.session_id);
		END`,
		`CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
			INSERT INTO memories_fts(memories_fts, rowid, content, category, session_id)
			VALUES ('delete', old.id, old.content, old.category, old.session_id);
			INSERT INTO memories_fts(rowid, content, category, session_id)
			VALUES (new.id, new.content, new.category, new.session_id);
		END`,
	}

	for _, trigger := range triggers {
		if _, err := db.Exec(trigger); err != nil {
			log.Printf("Warning: create trigger: %v", err)
		}
	}

	log.Printf("FTS5 memory store: %s", dbPath)

	return &Store{db: db, basePath: basePath}, nil
}

// Save adds a memory entry.
func (s *Store) Save(content, category, sessionID string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO memories (content, category, session_id) VALUES (?, ?, ?)",
		content, category, sessionID,
	)
	if err != nil {
		return 0, fmt.Errorf("save memory: %w", err)
	}
	return result.LastInsertId()
}

// SaveDaily saves a daily memory note (convenience wrapper).
func (s *Store) SaveDaily(content string) (int64, error) {
	return s.Save(content, "daily", "")
}

// SaveLongTerm saves a long-term memory note.
func (s *Store) SaveLongTerm(content string) (int64, error) {
	return s.Save(content, "longterm", "")
}

// SaveSession saves a conversation session entry.
func (s *Store) SaveSession(sessionID, content string) (int64, error) {
	return s.Save(content, "session", sessionID)
}

// Search performs full-text search across all memories.
// Returns ranked results (best match first).
func (s *Store) Search(query string, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 10
	}

	// Use FTS5 for search with bm25 ranking
	rows, err := s.db.Query(`
		SELECT m.id, m.content, m.category, m.session_id, m.created_at, rank
		FROM memories_fts fts
		JOIN memories m ON m.id = fts.rowid
		WHERE memories_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fts5 search: %w", err)
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Content, &e.Category, &e.SessionID, &e.CreatedAt, &e.Score); err != nil {
			continue
		}
		results = append(results, e)
	}
	return results, nil
}

// SearchByCategory searches within a specific category.
func (s *Store) SearchByCategory(query, category string, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.Query(`
		SELECT m.id, m.content, m.category, m.session_id, m.created_at, rank
		FROM memories_fts fts
		JOIN memories m ON m.id = fts.rowid
		WHERE memories_fts MATCH ? AND m.category = ?
		ORDER BY rank
		LIMIT ?
	`, query, category, limit)
	if err != nil {
		return nil, fmt.Errorf("fts5 category search: %w", err)
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Content, &e.Category, &e.SessionID, &e.CreatedAt, &e.Score); err != nil {
			continue
		}
		results = append(results, e)
	}
	return results, nil
}

// GetDaily returns today's daily memories.
func (s *Store) GetDaily() []Entry {
	today := time.Now().Format("2006-01-02")
	rows, err := s.db.Query(`
		SELECT id, content, category, session_id, created_at
		FROM memories
		WHERE category = 'daily' AND date(created_at) = ?
		ORDER BY created_at
	`, today)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Content, &e.Category, &e.SessionID, &e.CreatedAt); err != nil {
			continue
		}
		results = append(results, e)
	}
	return results
}

// GetLongTerm returns all long-term memories (most recent first).
func (s *Store) GetLongTerm(limit int) []Entry {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, content, category, session_id, created_at
		FROM memories
		WHERE category = 'longterm'
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Content, &e.Category, &e.SessionID, &e.CreatedAt); err != nil {
			continue
		}
		results = append(results, e)
	}
	return results
}

// Context returns memory context to inject into the agent system prompt.
func (s *Store) Context() string {
	var parts []string

	// Long-term memory (recent 20)
	longTerm := s.GetLongTerm(20)
	if len(longTerm) > 0 {
		var lines []string
		for _, e := range longTerm {
			lines = append(lines, e.Content)
		}
		parts = append(parts, "## Long-term Memory\n"+strings.Join(lines, "\n"))
	}

	// Today's daily memory
	daily := s.GetDaily()
	if len(daily) > 0 {
		var lines []string
		for _, e := range daily {
			lines = append(lines, e.Content)
		}
		parts = append(parts, "## Today's Notes\n"+strings.Join(lines, "\n"))
	}

	if len(parts) == 0 {
		return ""
	}
	return "<memory>\n" + strings.Join(parts, "\n\n") + "\n</memory>"
}

// Delete removes a memory by ID.
func (s *Store) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM memories WHERE id = ?", id)
	return err
}

// Count returns total memory count.
func (s *Store) Count() int {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM memories").Scan(&count)
	return count
}

// CountByCategory returns memory count per category.
func (s *Store) CountByCategory() map[string]int {
	rows, err := s.db.Query("SELECT category, COUNT(*) FROM memories GROUP BY category")
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var cat string
		var count int
		rows.Scan(&cat, &count)
		result[cat] = count
	}
	return result
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

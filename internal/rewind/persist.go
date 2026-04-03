package rewind

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// PersistStore is a RewindStore with SQLite persistence.
type PersistStore struct {
	*RewindStore
	dbPath string
	db     *sql.DB
	mu     sync.RWMutex
}

// NewPersistStore creates a new persistent rewind store.
// If dbPath is empty, it defaults to ~/.aigo/rewind.db
func NewPersistStore(dbPath string) (*PersistStore, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(home, ".aigo", "rewind.db")
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table
	schema := `
	CREATE TABLE IF NOT EXISTS rewind_entries (
		hash TEXT PRIMARY KEY,
		full_hash TEXT NOT NULL,
		content TEXT NOT NULL,
		content_type TEXT NOT NULL,
		original_size INTEGER NOT NULL,
		compressed_size INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		session_id TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_session_id ON rewind_entries(session_id);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON rewind_entries(timestamp);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	ps := &PersistStore{
		RewindStore: NewRewindStore(),
		dbPath:      dbPath,
		db:          db,
	}

	// Load existing entries into memory
	if err := ps.loadFromDB(); err != nil {
		db.Close()
		return nil, err
	}

	return ps, nil
}

// loadFromDB loads all entries from SQLite into memory.
func (ps *PersistStore) loadFromDB() error {
	rows, err := ps.db.Query(`
		SELECT hash, full_hash, content, content_type, 
		       original_size, compressed_size, timestamp, session_id
		FROM rewind_entries
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var e RewindEntry
		if err := rows.Scan(
			&e.Hash, &e.FullHash, &e.Content, &e.ContentType,
			&e.OriginalSize, &e.CompressedSize, &e.Timestamp, &e.SessionID,
		); err != nil {
			return err
		}
		ps.entries[e.Hash] = &e
	}

	return rows.Err()
}

// Store saves content to both memory and SQLite.
func (ps *PersistStore) Store(content, contentType, sessionID string) string {
	shortHash := ps.RewindStore.Store(content, contentType, sessionID)

	// Get the entry
	ps.mu.RLock()
	entry, ok := ps.entries[shortHash]
	ps.mu.RUnlock()

	if !ok {
		return shortHash
	}

	// Persist to SQLite
	_, err := ps.db.Exec(`
		INSERT OR REPLACE INTO rewind_entries 
		(hash, full_hash, content, content_type, original_size, compressed_size, timestamp, session_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, entry.Hash, entry.FullHash, entry.Content, entry.ContentType,
		entry.OriginalSize, entry.CompressedSize, entry.Timestamp, entry.SessionID)

	if err != nil {
		// Log error but don't fail - memory store is primary
		// In production, would log this
	}

	return shortHash
}

// Delete removes an entry from both memory and SQLite.
func (ps *PersistStore) Delete(shortHash string) error {
	ps.mu.Lock()
	delete(ps.entries, shortHash)
	ps.mu.Unlock()

	_, err := ps.db.Exec("DELETE FROM rewind_entries WHERE hash = ?", shortHash)
	return err
}

// Close closes the database connection.
func (ps *PersistStore) Close() error {
	return ps.db.Close()
}

// Stats returns statistics about the store.
func (ps *PersistStore) Stats() map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var totalSize int64
	for _, e := range ps.entries {
		totalSize += int64(e.OriginalSize)
	}

	return map[string]interface{}{
		"count":       len(ps.entries),
		"total_bytes": totalSize,
		"db_path":     ps.dbPath,
	}
}

// PurgeOlderThan removes entries older than the given duration.
func (ps *PersistStore) PurgeOlderThan(olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)

	ps.mu.Lock()
	defer ps.mu.Unlock()

	var toDelete []string
	for hash, entry := range ps.entries {
		if entry.Timestamp.Before(cutoff) {
			toDelete = append(toDelete, hash)
		}
	}

	for _, hash := range toDelete {
		delete(ps.entries, hash)
	}

	result, err := ps.db.Exec("DELETE FROM rewind_entries WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// GetBySession returns all entries for a session.
func (ps *PersistStore) GetBySession(sessionID string) []RewindEntry {
	return ps.List(sessionID)
}

// GetByTimeRange returns entries within a time range.
func (ps *PersistStore) GetByTimeRange(start, end time.Time) []RewindEntry {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var result []RewindEntry
	for _, e := range ps.entries {
		if e.Timestamp.After(start) && e.Timestamp.Before(end) {
			result = append(result, *e)
		}
	}
	return result
}

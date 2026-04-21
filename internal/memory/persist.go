package memory

import (
	"database/sql"
	"fmt"
	"time"
)

type SQLiteStore struct {
	db       *sql.DB
	dbPath  string
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	store := &SQLiteStore{
		db:      db,
		dbPath: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		category TEXT NOT NULL,
		embedding BLOB,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_memories_category ON memories(category);
	CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) SaveMemory(content, category string, embedding []byte) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO memories(content, category, embedding) VALUES(?, ?, ?)",
		content, category, embedding,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetMemory(id int64) (*MemoryRecord, error) {
	row := s.db.QueryRow(
		"SELECT id, content, category, embedding, created_at FROM memories WHERE id = ?",
		id,
	)

	var record MemoryRecord
	err := row.Scan(&record.ID, &record.Content, &record.Category, &record.Embedding, &record.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (s *SQLiteStore) SearchMemories(query string, limit int) ([]MemoryRecord, error) {
	rows, err := s.db.Query(
		"SELECT id, content, category, embedding, created_at FROM memories WHERE content LIKE ? LIMIT ?",
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]MemoryRecord, 0)
	for rows.Next() {
		var record MemoryRecord
		if err := rows.Scan(&record.ID, &record.Content, &record.Category, &record.Embedding, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *SQLiteStore) UpdateMemory(id int64, content string, embedding []byte) error {
	_, err := s.db.Exec(
		"UPDATE memories SET content = ?, embedding = ?, updated_at = ? WHERE id = ?",
		content, embedding, time.Now(), id,
	)
	return err
}

func (s *SQLiteStore) DeleteMemory(id int64) error {
	_, err := s.db.Exec("DELETE FROM memories WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) ListMemories(category string, limit int) ([]MemoryRecord, error) {
	query := "SELECT id, content, category, embedding, created_at FROM memories"
	args := []interface{}{}

	if category != "" {
		query += " WHERE category = ?"
		args = append(args, category)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]MemoryRecord, 0)
	for rows.Next() {
		var record MemoryRecord
		if err := rows.Scan(&record.ID, &record.Content, &record.Category, &record.Embedding, &record.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

type MemoryRecord struct {
	ID        int64
	Content  string
	Category string
	Embedding []byte
	CreatedAt time.Time
}
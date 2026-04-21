package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type SessionStore struct {
	db      *sql.DB
	dbPath string
}

func NewSessionStore(dbPath string) (*SessionStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	store := &SessionStore{
		db:      db,
		dbPath: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

func (s *SessionStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		state TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		tool_calls TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *SessionStore) SaveSession(id string, state []byte) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT OR REPLACE INTO sessions(id, state, updated_at) VALUES(?, ?, ?)`,
		id, string(state), time.Now(),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SessionStore) LoadSession(id string) ([]byte, error) {
	row := s.db.QueryRow(
		"SELECT state FROM sessions WHERE id = ?",
		id,
	)

	var state []byte
	err := row.Scan(&state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *SessionStore) DeleteSession(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM messages WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SessionStore) AddMessage(sessionID, role, content string, toolCalls []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO messages(session_id, role, content, tool_calls) VALUES(?, ?, ?, ?)`,
		sessionID, role, content, toolCalls,
	)
	return err
}

func (s *SessionStore) GetMessages(sessionID string, limit int) ([]SessionMessage, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, role, content, tool_calls, created_at 
		 FROM messages WHERE session_id = ? ORDER BY created_at ASC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]SessionMessage, 0)
	for rows.Next() {
		var msg SessionMessage
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.ToolCalls, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (s *SessionStore) ListSessions(limit int) ([]SessionInfo, error) {
	rows, err := s.db.Query(
		"SELECT id, created_at, updated_at FROM sessions ORDER BY updated_at DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]SessionInfo, 0)
	for rows.Next() {
		var info SessionInfo
		if err := rows.Scan(&info.ID, &info.CreatedAt, &info.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, info)
	}

	return sessions, rows.Err()
}

func (s *SessionStore) Close() error {
	return s.db.Close()
}

type SessionMessage struct {
	ID        int64
	SessionID string
	Role     string
	Content  string
	ToolCalls []byte
	CreatedAt time.Time
}

type SessionInfo struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func MarshalState(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func UnmarshalState(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
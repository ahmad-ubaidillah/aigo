package planning

import (
	"database/sql"
	"fmt"
	"time"
)

type Decision struct {
	Key     string
	Value   string
	Source  string
	Created time.Time
}

type DecisionStore struct {
	decisions map[string]*Decision
}

type DecisionDBStore struct {
	db      *sql.DB
	dbPath string
}

func NewDecision(key, value string) *Decision {
	return &Decision{
		Key:    key,
		Value:  value,
		Source: "default",
	}
}

func NewDecisionStore() *DecisionStore {
	return &DecisionStore{
		decisions: make(map[string]*Decision),
	}
}

func (ds *DecisionStore) Add(key, value string) {
	ds.decisions[key] = &Decision{
		Key:     key,
		Value:   value,
		Created: time.Now(),
	}
}

func (ds *DecisionStore) Get(key string) *Decision {
	return ds.decisions[key]
}

func (ds *DecisionStore) Len() int {
	return len(ds.decisions)
}

func (ds *DecisionStore) List() []*Decision {
	result := make([]*Decision, 0, len(ds.decisions))
	for _, d := range ds.decisions {
		result = append(result, d)
	}
	return result
}

func NewDecisionDBStore(dbPath string) (*DecisionDBStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	store := &DecisionDBStore{
		db:      db,
		dbPath: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

func (s *DecisionDBStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS decisions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		source TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_decisions_key ON decisions(key);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *DecisionDBStore) Save(key, value, source string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO decisions(key, value, source) VALUES(?, ?, ?)`,
		key, value, source,
	)
	return err
}

func (s *DecisionDBStore) Get(key string) (string, string, error) {
	row := s.db.QueryRow(
		"SELECT value, source FROM decisions WHERE key = ?",
		key,
	)

	var value, source string
	err := row.Scan(&value, &source)
	if err != nil {
		return "", "", err
	}

	return value, source, nil
}

func (s *DecisionDBStore) GetAll() ([]Decision, error) {
	rows, err := s.db.Query(
		"SELECT key, value, source, created_at FROM decisions ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decisions := make([]Decision, 0)
	for rows.Next() {
		var d Decision
		if err := rows.Scan(&d.Key, &d.Value, &d.Source, &d.Created); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}

	return decisions, rows.Err()
}

func (s *DecisionDBStore) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM decisions WHERE key = ?", key)
	return err
}

func (s *DecisionDBStore) Close() error {
	return s.db.Close()
}
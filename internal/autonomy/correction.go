package autonomy

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

type SelfCorrector struct {
	maxRetries  int
	patterns  map[string]string
	mu        sync.RWMutex
}

func NewSelfCorrector() *SelfCorrector {
	return &SelfCorrector{
		maxRetries: 3,
		patterns: map[string]string{
			"undefined":  "check import",
			"no such file": "check path",
			"syntax error": "check syntax",
			"null pointer": "check nil",
		},
	}
}

func (sc *SelfCorrector) AnalyzeError(errMsg string) string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	for pattern, fix := range sc.patterns {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			return fix
		}
	}
	return "manual review required"
}

func (sc *SelfCorrector) ShouldRetry(attempt int) bool {
	return attempt < sc.maxRetries
}

func (sc *SelfCorrector) RetryWithFix(errMsg string) string {
	fix := sc.AnalyzeError(errMsg)
	return "Applied fix: " + fix
}

func (sc *SelfCorrector) AddPattern(pattern, fix string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.patterns[pattern] = fix
}

func (sc *SelfCorrector) GetPatterns() map[string]string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range sc.patterns {
		result[k] = v
	}
	return result
}

type ErrorPattern struct {
	ID          int64
	Pattern    string
	Frequency  int
	SuggestedFix string
	LastSeen  time.Time
}

type ErrorPatternStore struct {
	db      *sql.DB
	dbPath string
}

func NewErrorPatternStore(dbPath string) (*ErrorPatternStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	store := &ErrorPatternStore{
		db:      db,
		dbPath: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

func (s *ErrorPatternStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS error_patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT UNIQUE NOT NULL,
		frequency INTEGER DEFAULT 1,
		suggested_fix TEXT,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_error_patterns ON error_patterns(pattern);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *ErrorPatternStore) Record(pattern, fix string) error {
	_, err := s.db.Exec(
		`INSERT INTO error_patterns(pattern, frequency, suggested_fix) VALUES(?, 1, ?)
		 ON CONFLICT(pattern) DO UPDATE SET frequency = frequency + 1, last_seen = datetime('now')`,
		pattern, fix,
	)
	return err
}

func (s *ErrorPatternStore) GetSuggestions(pattern string) ([]ErrorPattern, error) {
	rows, err := s.db.Query(
		`SELECT id, pattern, frequency, suggested_fix, last_seen 
		 FROM error_patterns WHERE pattern LIKE ? ORDER BY frequency DESC LIMIT 5`,
		"%"+pattern+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := make([]ErrorPattern, 0)
	for rows.Next() {
		var p ErrorPattern
		if err := rows.Scan(&p.ID, &p.Pattern, &p.Frequency, &p.SuggestedFix, &p.LastSeen); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}

	return patterns, rows.Err()
}

func (s *ErrorPatternStore) GetTopPatterns(limit int) ([]ErrorPattern, error) {
	rows, err := s.db.Query(
		`SELECT id, pattern, frequency, suggested_fix, last_seen 
		 FROM error_patterns ORDER BY frequency DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := make([]ErrorPattern, 0)
	for rows.Next() {
		var p ErrorPattern
		if err := rows.Scan(&p.ID, &p.Pattern, &p.Frequency, &p.SuggestedFix, &p.LastSeen); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}

	return patterns, rows.Err()
}

func (s *ErrorPatternStore) Close() error {
	return s.db.Close()
}

type SuccessPattern struct {
	ID           int64
	Pattern     string
	Frequency   int
	Recommendation string
	LastSeen   time.Time
}

type SuccessPatternStore struct {
	db      *sql.DB
	dbPath string
}

func NewSuccessPatternStore(dbPath string) (*SuccessPatternStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	store := &SuccessPatternStore{
		db:      db,
		dbPath: dbPath,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

func (s *SuccessPatternStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS success_patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT UNIQUE NOT NULL,
		frequency INTEGER DEFAULT 1,
		recommendation TEXT,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_success_patterns ON success_patterns(pattern);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SuccessPatternStore) Record(pattern, recommendation string) error {
	_, err := s.db.Exec(
		`INSERT INTO success_patterns(pattern, frequency, recommendation) VALUES(?, 1, ?)
		 ON CONFLICT(pattern) DO UPDATE SET frequency = frequency + 1, last_seen = datetime('now')`,
		pattern, recommendation,
	)
	return err
}

func (s *SuccessPatternStore) GetRecommendations(pattern string) ([]SuccessPattern, error) {
	rows, err := s.db.Query(
		`SELECT id, pattern, frequency, recommendation, last_seen 
		 FROM success_patterns WHERE pattern LIKE ? ORDER BY frequency DESC LIMIT 5`,
		"%"+pattern+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := make([]SuccessPattern, 0)
	for rows.Next() {
		var p SuccessPattern
		if err := rows.Scan(&p.ID, &p.Pattern, &p.Frequency, &p.Recommendation, &p.LastSeen); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}

	return patterns, rows.Err()
}

func (s *SuccessPatternStore) GetTopRecommendations(limit int) ([]SuccessPattern, error) {
	rows, err := s.db.Query(
		`SELECT id, pattern, frequency, recommendation, last_seen 
		 FROM success_patterns ORDER BY frequency DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := make([]SuccessPattern, 0)
	for rows.Next() {
		var p SuccessPattern
		if err := rows.Scan(&p.ID, &p.Pattern, &p.Frequency, &p.Recommendation, &p.LastSeen); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}

	return patterns, rows.Err()
}

func (s *SuccessPatternStore) Close() error {
	return s.db.Close()
}
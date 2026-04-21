package autonomy

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type LearnedPattern struct {
	ID           int
	ErrorPattern string
	Fix          string
	SuccessCount int
	FailCount    int
	CreatedAt    time.Time
	LastUsed     time.Time
}

type LearningErrorAnalyzer struct {
	mu          sync.RWMutex
	db          *sql.DB
	patterns    map[string]string
	baseDir     string
	initialized bool
}

func NewLearningErrorAnalyzer(baseDir string) (*LearningErrorAnalyzer, error) {
	le := &LearningErrorAnalyzer{
		baseDir:  baseDir,
		patterns: make(map[string]string),
	}

	if err := le.initDB(); err != nil {
		return nil, err
	}

	le.loadStaticPatterns()
	le.loadLearnedPatterns()

	return le, nil
}

func (le *LearningErrorAnalyzer) initDB() error {
	dbPath := filepath.Join(le.baseDir, "error_patterns.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS error_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			error_pattern TEXT NOT NULL,
			fix TEXT NOT NULL,
			success_count INTEGER DEFAULT 0,
			fail_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(error_pattern, fix)
		);
		CREATE INDEX IF NOT EXISTS idx_error_pattern ON error_patterns(error_pattern);
	`)
	if err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	le.db = db
	le.initialized = true
	return nil
}

func (le *LearningErrorAnalyzer) loadStaticPatterns() {
	staticPatterns := map[string]string{
		"null pointer":          "check for nil before use",
		"undefined":             "check import statements and declarations",
		"type mismatch":         "check type assertions and conversions",
		"index out of bounds":   "check array length and indices",
		"deadlock":              "check goroutine synchronization and channel usage",
		"connection refused":    "check if service is running and port is correct",
		"permission denied":     "check file/directory permissions",
		"not found":             "verify file path or resource exists",
		"timeout":               "increase timeout or check network connectivity",
		"out of memory":         "optimize memory usage or increase limits",
	}

	for k, v := range staticPatterns {
		le.patterns[k] = v
	}
}

func (le *LearningErrorAnalyzer) loadLearnedPatterns() error {
	if !le.initialized || le.db == nil {
		return nil
	}

	rows, err := le.db.Query(`
		SELECT error_pattern, fix, success_count 
		FROM error_patterns 
		WHERE success_count > 0 
		ORDER BY success_count DESC 
		LIMIT 50
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pattern, fix string
		var successCount int
		if err := rows.Scan(&pattern, &fix, &successCount); err != nil {
			continue
		}
		le.patterns[pattern] = fix
	}

	return nil
}

func (le *LearningErrorAnalyzer) AnalyzePattern(errorMsg string) string {
	le.mu.RLock()
	defer le.mu.RUnlock()

	errorLower := strings.ToLower(errorMsg)

	for pattern, fix := range le.patterns {
		if strings.Contains(errorLower, strings.ToLower(pattern)) {
			le.recordUsage(pattern)
			return fix
		}
	}

	return "manual review required"
}

func (le *LearningErrorAnalyzer) LearnSuccess(errorMsg, fix string) error {
	if !le.initialized || le.db == nil {
		return nil
	}

	errorLower := strings.ToLower(errorMsg)
	pattern := extractPattern(errorLower)

	le.mu.Lock()
	defer le.mu.Unlock()

	_, err := le.db.Exec(`
		INSERT INTO error_patterns (error_pattern, fix, success_count, last_used)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(error_pattern, fix) DO UPDATE SET
			success_count = success_count + 1,
			last_used = CURRENT_TIMESTAMP
	`, pattern, fix)

	if err == nil {
		le.patterns[pattern] = fix
	}

	return err
}

func (le *LearningErrorAnalyzer) LearnFailure(errorMsg, fix string) error {
	if !le.initialized || le.db == nil {
		return nil
	}

	errorLower := strings.ToLower(errorMsg)
	pattern := extractPattern(errorLower)

	le.mu.Lock()
	defer le.mu.Unlock()

	_, err := le.db.Exec(`
		UPDATE error_patterns 
		SET fail_count = fail_count + 1,
		    last_used = CURRENT_TIMESTAMP
		WHERE error_pattern = ? AND fix = ?
	`, pattern, fix)

	return err
}

func (le *LearningErrorAnalyzer) GetSuggestedFixes(errorMsg string) []string {
	if !le.initialized || le.db == nil {
		return nil
	}

	errorLower := strings.ToLower(errorMsg)
	pattern := extractPattern(errorLower)

	rows, err := le.db.Query(`
		SELECT fix, success_count, fail_count 
		FROM error_patterns 
		WHERE error_pattern = ?
		ORDER BY (success_count + 1.0) / (fail_count + 1.0) DESC
		LIMIT 5
	`, pattern)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var fixes []string
	for rows.Next() {
		var fix string
		var success, fail int
		if err := rows.Scan(&fix, &success, &fail); err != nil {
			continue
		}
		fixes = append(fixes, fix)
	}

	return fixes
}

func (le *LearningErrorAnalyzer) GetTopPatterns(limit int) []LearnedPattern {
	if !le.initialized || le.db == nil {
		return nil
	}

	rows, err := le.db.Query(`
		SELECT id, error_pattern, fix, success_count, fail_count, created_at, last_used
		FROM error_patterns
		ORDER BY success_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var patterns []LearnedPattern
	for rows.Next() {
		var p LearnedPattern
		if err := rows.Scan(&p.ID, &p.ErrorPattern, &p.Fix, &p.SuccessCount, &p.FailCount, &p.CreatedAt, &p.LastUsed); err != nil {
			continue
		}
		patterns = append(patterns, p)
	}

	return patterns
}

func (le *LearningErrorAnalyzer) recordUsage(pattern string) {
	if !le.initialized || le.db == nil {
		return
	}

	le.db.Exec(`UPDATE error_patterns SET last_used = CURRENT_TIMESTAMP WHERE error_pattern = ?`, pattern)
}

func extractPattern(errorMsg string) string {
	errorMsg = strings.ToLower(errorMsg)

	commonPatterns := []string{
		"null pointer", "nil pointer", "undefined", "not defined",
		"type mismatch", "cannot convert", "incompatible types",
		"index out of bounds", "array index", "slice index",
		"deadlock", "concurrent map", "send on closed",
		"connection refused", "connection reset", "connection timeout",
		"permission denied", "access denied", "unauthorized",
		"not found", "does not exist", "no such file",
		"timeout", "timed out", "deadline exceeded",
		"out of memory", "memory limit", "heap limit",
		"syntax error", "parse error", "unexpected token",
		"reference error", "undefined reference",
		"no such method", "method not found",
	}

	for _, p := range commonPatterns {
		if strings.Contains(errorMsg, p) {
			return p
		}
	}

	if len(errorMsg) > 50 {
		return errorMsg[:50]
	}
	return errorMsg
}

func (le *LearningErrorAnalyzer) Close() error {
	if le.db != nil {
		return le.db.Close()
	}
	return nil
}

func (le *LearningErrorAnalyzer) Initialize(ctx context.Context) error {
	if le.initialized {
		return nil
	}

	le.mu.Lock()
	defer le.mu.Unlock()

	if le.initialized {
		return nil
	}

	if err := le.initDB(); err != nil {
		return err
	}

	le.loadStaticPatterns()
	le.loadLearnedPatterns()

	return nil
}
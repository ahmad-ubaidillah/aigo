// Package vector implements semantic vector memory backed by SQLite + sqlite-vec.
//
// It stores text observations with lightweight hash-based embeddings and
// supports cosine similarity search via the vec_cosine SQL function from
// github.com/viant/sqlite-vec. No large embedding model needed.
//
// Architecture:
//   - Embeddings: 256-dim SimHash (fast, deterministic, no model download)
//   - Storage: SQLite with BLOB embeddings + FTS5 for hybrid search
//   - Search: vec_cosine() for semantic, FTS5 MATCH for keyword
//   - Hybrid: Combined semantic + keyword scoring (configurable weight)
package vector

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/viant/sqlite-vec/engine"
	_ "modernc.org/sqlite"
)

// VectorStore manages semantic vector memory.
type VectorStore struct {
	db       *sql.DB
	dbPath   string
	dimSize  int
}

// Entry represents a stored vector memory.
type Entry struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	Category  string    `json:"category"`
	Tags      string    `json:"tags"`
	Score     float64   `json:"score,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

const defaultDimSize = 256

// New opens or creates a vector memory store at the given path.
func New(dataDir string) (*VectorStore, error) {
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("vector: home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".aigo", "memory", "vector")
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("vector: mkdir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "vector.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("vector: open db: %w", err)
	}

	// Register vec_cosine / vec_l2 functions
	if err := engine.RegisterVectorFunctions(db); err != nil {
		return nil, fmt.Errorf("vector: register vec functions: %w", err)
	}

	// Enable WAL mode for concurrent reads
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("vector: wal: %w", err)
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS vec_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		category TEXT DEFAULT 'general',
		tags TEXT DEFAULT '',
		embedding BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_vec_category ON vec_entries(category);
	CREATE INDEX IF NOT EXISTS idx_vec_created ON vec_entries(created_at);

	CREATE VIRTUAL TABLE IF NOT EXISTS vec_fts USING fts5(
		text, category, tags,
		content='vec_entries',
		content_rowid='id',
		tokenize='porter unicode61'
	);

	-- Triggers to keep FTS5 in sync
	CREATE TRIGGER IF NOT EXISTS vec_entries_ai AFTER INSERT ON vec_entries BEGIN
		INSERT INTO vec_fts(rowid, text, category, tags)
		VALUES (new.id, new.text, new.category, new.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS vec_entries_ad AFTER DELETE ON vec_entries BEGIN
		INSERT INTO vec_fts(vec_fts, rowid, text, category, tags)
		VALUES ('delete', old.id, old.text, old.category, old.tags);
	END;

	CREATE TABLE IF NOT EXISTS vec_meta (
		key TEXT PRIMARY KEY,
		value TEXT
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("vector: create schema: %w", err)
	}

	// Store dim size in meta
	dimSize := defaultDimSize
	db.Exec("INSERT OR REPLACE INTO vec_meta(key,value) VALUES('dim_size', ?)", fmt.Sprintf("%d", dimSize))

	return &VectorStore{
		db:      db,
		dbPath:  dbPath,
		dimSize: dimSize,
	}, nil
}

// Save stores text with an auto-generated embedding.
func (vs *VectorStore) Save(text, category, tags string) (int64, error) {
	if text == "" {
		return 0, fmt.Errorf("vector: empty text")
	}
	if category == "" {
		category = "general"
	}

	emb := Embed(text, vs.dimSize)
	embBytes := float32ToBytes(emb)

	res, err := vs.db.Exec(
		"INSERT INTO vec_entries(text, category, tags, embedding) VALUES(?, ?, ?, ?)",
		text, category, tags, embBytes,
	)
	if err != nil {
		return 0, fmt.Errorf("vector: insert: %w", err)
	}
	return res.LastInsertId()
}

// Search performs hybrid semantic + keyword search.
// alpha controls the blend: 1.0 = pure semantic, 0.0 = pure keyword.
func (vs *VectorStore) Search(query string, limit int, alpha float64) ([]Entry, error) {
	if limit <= 0 {
		limit = 10
	}
	if alpha < 0 || alpha > 1 {
		alpha = 0.7
	}

	queryEmb := Embed(query, vs.dimSize)
	queryBytes := float32ToBytes(queryEmb)

	// Hybrid search: semantic + FTS5 keyword
	rows, err := vs.db.Query(`
		SELECT
			e.id, e.text, e.category, e.tags, e.created_at,
			COALESCE(vec_cosine(e.embedding, ?), 0) AS sem_score,
			COALESCE(
				(SELECT rank FROM vec_fts WHERE vec_fts.rowid = e.id ORDER BY rank LIMIT 1), 0
			) AS kw_score
		FROM vec_entries e
		WHERE e.text != ''
		ORDER BY (? * sem_score + (1 - ?) * ABS(kw_score)) DESC
		LIMIT ?
	`, queryBytes, alpha, alpha, limit)
	if err != nil {
		return nil, fmt.Errorf("vector: search: %w", err)
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		var createdAt string
		var semScore, kwScore float64
		if err := rows.Scan(&e.ID, &e.Text, &e.Category, &e.Tags, &createdAt, &semScore, &kwScore); err != nil {
			continue
		}
		e.Score = alpha*semScore + (1-alpha)*math.Abs(kwScore)
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, e)
	}
	return results, nil
}

// SemanticSearch performs pure cosine similarity search.
func (vs *VectorStore) SemanticSearch(query string, limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 10
	}
	queryEmb := Embed(query, vs.dimSize)
	queryBytes := float32ToBytes(queryEmb)

	rows, err := vs.db.Query(`
		SELECT id, text, category, tags, created_at,
			COALESCE(vec_cosine(embedding, ?), 0) AS score
		FROM vec_entries
		ORDER BY score DESC
		LIMIT ?
	`, queryBytes, limit)
	if err != nil {
		return nil, fmt.Errorf("vector: semantic search: %w", err)
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		var createdAt string
		if err := rows.Scan(&e.ID, &e.Text, &e.Category, &e.Tags, &createdAt, &e.Score); err != nil {
			continue
		}
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, e)
	}
	return results, nil
}

// Delete removes an entry by ID.
func (vs *VectorStore) Delete(id int64) error {
	_, err := vs.db.Exec("DELETE FROM vec_entries WHERE id = ?", id)
	return err
}

// Count returns the total number of entries.
func (vs *VectorStore) Count() (int, error) {
	var count int
	err := vs.db.QueryRow("SELECT COUNT(*) FROM vec_entries").Scan(&count)
	return count, err
}

// Categories returns distinct categories with counts.
func (vs *VectorStore) Categories() (map[string]int, error) {
	rows, err := vs.db.Query("SELECT category, COUNT(*) FROM vec_entries GROUP BY category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var cat string
		var count int
		rows.Scan(&cat, &count)
		result[cat] = count
	}
	return result, nil
}

// Close closes the database.
func (vs *VectorStore) Close() error {
	return vs.db.Close()
}

// --- Embedding: 256-dim SimHash ---
// Deterministic, fast, no model download. Good enough for semantic similarity.

// Embed generates a SimHash embedding of the given text.
func Embed(text string, dim int) []float32 {
	if dim <= 0 {
		dim = defaultDimSize
	}

	// Tokenize
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return make([]float32, dim)
	}

	// Build weighted feature vector
	features := make(map[string]float64)
	for _, t := range tokens {
		features[t]++
	}

	// SimHash: for each bit, sum weighted feature hashes
	vec := make([]float32, dim)
	for token, weight := range features {
		h := sha256.Sum256([]byte(token))
		for i := 0; i < dim && i < len(h)*8; i++ {
			byteIdx := i / 8
			bitIdx := uint(i % 8)
			if (h[byteIdx]>>(7-bitIdx))&1 == 1 {
				vec[i] += float32(weight)
			} else {
				vec[i] -= float32(weight)
			}
		}
	}

	// Normalize to unit vector
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= float32(norm)
		}
	}

	return vec
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	// Simple split on non-alphanumeric
	var tokens []string
	var current strings.Builder
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r > 127 {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	// Remove very short tokens and common stopwords
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "it": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "and": true, "or": true, "but": true, "not": true,
		"with": true, "this": true, "that": true, "from": true,
		"by": true, "as": true, "be": true, "was": true, "are": true,
		"di": true, "dan": true, "yang": true, "atau": true, "ini": true,
		"itu": true, "untuk": true, "dari": true, "dengan": true,
	}

	var filtered []string
	for _, t := range tokens {
		if len(t) >= 2 && !stopwords[t] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func float32ToBytes(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

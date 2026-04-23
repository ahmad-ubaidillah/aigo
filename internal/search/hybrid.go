package search

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

type HybridSearch struct {
	db           *sql.DB
	dbPath       string
	mu           sync.RWMutex
	basePath     string
	vectorDim    int
	vectorWeight float64
	keywordWeight float64
	rrfK         float64
}

type SearchResult struct {
	ID          string
	Title       string
	Content     string
	URL         string
	Score       float64
	Source      string
	EntityLinks []string
	BacklinkCount int
}

type SearchOptions struct {
	Query         string
	Limit         int
	Intent        string
	IncludeGraph  bool
	BoostCompiled bool
	BoostBacklinks bool
}

func New(basePath string) (*HybridSearch, error) {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, ".aigo", "memory", "search")
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("search mkdir: %w", err)
	}

	dbPath := filepath.Join(basePath, "search.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("search open: %w", err)
	}

	h := &HybridSearch{
		db:            db,
		dbPath:        dbPath,
		basePath:      basePath,
		vectorDim:     256,
		vectorWeight:  0.5,
		keywordWeight: 0.5,
		rrfK:          60,
	}

	if err := h.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return h, nil
}

func (h *HybridSearch) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT,
		url TEXT,
		doc_type TEXT DEFAULT 'note',
		tags TEXT,
		embedding BLOB,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		compiled_truth BOOLEAN DEFAULT 0,
		is_timeline BOOLEAN DEFAULT 0
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
		title,
		content,
		tags,
		content=documents,
		content_rowid=rowid
	);

	CREATE TRIGGER IF NOT EXISTS documents_ai AFTER INSERT ON documents BEGIN
		INSERT INTO documents_fts(rowid, title, content, tags)
		VALUES (NEW.rowid, NEW.title, NEW.content, NEW.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS documents_ad AFTER DELETE ON documents BEGIN
		INSERT INTO documents_fts(documents_fts, rowid, title, content, tags)
		VALUES ('delete', OLD.rowid, OLD.title, OLD.content, OLD.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS documents_au AFTER UPDATE ON documents BEGIN
		INSERT INTO documents_fts(documents_fts, rowid, title, content, tags)
		VALUES ('delete', OLD.rowid, OLD.title, OLD.content, OLD.tags);
		INSERT INTO documents_fts(rowid, title, content, tags)
		VALUES (NEW.rowid, NEW.title, NEW.content, NEW.tags);
	END;

	CREATE INDEX IF NOT EXISTS idx_docs_type ON documents(doc_type);
	CREATE INDEX IF NOT EXISTS idx_docs_compiled ON documents(compiled_truth);
	CREATE INDEX IF NOT EXISTS idx_docs_updated ON documents(updated_at);
	`

	_, err := h.db.Exec(schema)
	return err
}

func (h *HybridSearch) Index(id, title, content, docType, tags string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	embedding := h.computeSimHash(title + " " + content)

	_, err := h.db.Exec(`
		INSERT INTO documents (id, title, content, doc_type, tags, embedding, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			content = excluded.content,
			tags = excluded.tags,
			embedding = excluded.embedding,
			updated_at = CURRENT_TIMESTAMP
	`, id, title, content, docType, tags, embedding)

	return err
}

func (h *HybridSearch) IndexBatch(docs []struct {
	ID    string
	Title string
	Content string
	Type  string
	Tags  string
}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, _ := tx.Prepare(`
		INSERT INTO documents (id, title, content, doc_type, tags, embedding, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			content = excluded.content,
			tags = excluded.tags,
			embedding = excluded.embedding,
			updated_at = CURRENT_TIMESTAMP
	`)

	for _, doc := range docs {
		embedding := h.computeSimHash(doc.Title + " " + doc.Content)
		stmt.Exec(doc.ID, doc.Title, doc.Content, doc.Type, doc.Tags, embedding)
	}

	return tx.Commit()
}

func (h *HybridSearch) computeSimHash(text string) []byte {
	hashes := make([]uint64, h.vectorDim)

	for i := 0; i < len(text)-1; i++ {
		hash := uint64(text[i])<<56 | uint64(text[i+1])<<48
		for j := 0; j < h.vectorDim; j++ {
			if hash&(1<<uint(j)) != 0 {
				hashes[j]++
			} else {
				hashes[j]--
			}
		}
	}

	bytes := make([]byte, h.vectorDim/8)
	for i := 0; i < h.vectorDim/8; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			idx := i*8 + j
			if idx < h.vectorDim && hashes[idx] > 0 {
				b |= 1 << uint(j)
			}
		}
		bytes[i] = b
	}

	return bytes
}

func (h *HybridSearch) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	queries := []string{opts.Query}

	if opts.Intent == "" {
		opts.Intent = h.classifyIntent(opts.Query)
	}

	if opts.Intent == "general" {
		queries = h.expandQuery(opts.Query)
	}

	keywordResults := h.keywordSearch(queries, opts.Limit*2)

	var vectorResults []SearchResult
	if h.hasVectorData() {
		vectorResults = h.vectorSearch(opts.Query, opts.Limit*2)
	}

	combined := h.rrfFusion(keywordResults, vectorResults, opts.Limit)

	if opts.BoostCompiled {
		combined = h.boostCompiledTruth(combined)
	}

	if opts.BoostBacklinks {
		combined = h.boostBacklinks(combined)
	}

	combined = h.deduplicate(combined, opts.Limit)

	return combined, nil
}

func (h *HybridSearch) classifyIntent(query string) string {
	queryLower := strings.ToLower(query)

	entityPatterns := []string{"who is", "who works at", "who founded", "what company", "person"}
	for _, p := range entityPatterns {
		if strings.Contains(queryLower, p) {
			return "entity"
		}
	}

	temporalPatterns := []string{"when", "date", "today", "yesterday", "this week", "last month", "history"}
	for _, p := range temporalPatterns {
		if strings.Contains(queryLower, p) {
			return "temporal"
		}
	}

	eventPatterns := []string{"meeting", "conference", "event", "launch", "release"}
	for _, p := range eventPatterns {
		if strings.Contains(queryLower, p) {
			return "event"
		}
	}

	return "general"
}

func (h *HybridSearch) expandQuery(query string) []string {
	queries := []string{query}

	variants := []struct{ from, to string }{
		{"fix", "bug error issue"},
		{"implement", "add create build"},
		{"find", "search locate look"},
		{"remove", "delete clear"},
		{"update", "change modify"},
	}

	for _, v := range variants {
		if strings.Contains(strings.ToLower(query), v.from) {
			queries = append(queries, strings.ReplaceAll(query, v.from, v.to))
		}
	}

	return queries
}

func (h *HybridSearch) keywordSearch(queries []string, limit int) []SearchResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	type ftsResult struct {
		id    string
		title string
		content string
		rank  int
	}

	var allResults []ftsResult

	for _, query := range queries {
		searchQuery := strings.Join(strings.Fields(query), " ")

		rows, _ := h.db.Query(`
			SELECT d.id, d.title, d.content, fts.rank
			FROM documents_fts fts
			JOIN documents d ON d.rowid = fts.rowid
			WHERE documents_fts MATCH ?
			ORDER BY rank
			LIMIT ?
		`, searchQuery, limit)

		if rows == nil {
			continue
		}

		for rows.Next() {
			var r ftsResult
			rows.Scan(&r.id, &r.title, &r.content, &r.rank)
			allResults = append(allResults, r)
		}
		rows.Close()
	}

	scoreMap := make(map[string]SearchResult)
	for _, r := range allResults {
		rank := float64(r.rank)
		if rank == 0 {
			rank = 1
		}
		score := 1.0 / (h.rrfK + rank)
		score *= h.keywordWeight

		if existing, ok := scoreMap[r.id]; ok {
			existing.Score += score
			scoreMap[r.id] = existing
		} else {
			scoreMap[r.id] = SearchResult{
				ID:      r.id,
				Title:   r.title,
				Content: r.content,
				Score:   score,
				Source:  "keyword",
			}
		}
	}

	var results []SearchResult
	for _, r := range scoreMap {
		results = append(results, r)
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (h *HybridSearch) hasVectorData() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var count int
	h.db.QueryRow("SELECT COUNT(*) FROM documents WHERE embedding IS NOT NULL").Scan(&count)
	return count > 0
}

func (h *HybridSearch) vectorSearch(query string, limit int) []SearchResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	queryEmbedding := h.computeSimHash(query)

	rows, err := h.db.Query(`
		SELECT id, title, content,
			CAST(embedding AS BLOB) as emb
		FROM documents
		WHERE embedding IS NOT NULL
	`, query)

	if err != nil || rows == nil {
		return nil
	}
	defer rows.Close()

	type candidate struct {
		id      string
		title   string
		content string
		sim     float64
	}

	var candidates []candidate

	for rows.Next() {
		var c candidate
		var emb []byte
		rows.Scan(&c.id, &c.title, &c.content, &emb)
		c.sim = h.cosineSimilarity(queryEmbedding, emb)
		candidates = append(candidates, c)
	}

	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].sim > candidates[i].sim {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	var results []SearchResult
	for i, c := range candidates {
		if i >= limit {
			break
		}
		results = append(results, SearchResult{
			ID:      c.id,
			Title:   c.title,
			Content: c.content,
			Score:   c.sim * h.vectorWeight,
			Source:  "vector",
		})
	}

	return results
}

func (h *HybridSearch) cosineSimilarity(a, b []byte) float64 {
	if len(a) != len(b) {
		return 0
	}

	dot := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		va := float64(int(a[i]))
		vb := float64(int(b[i]))
		dot += va * vb
		normA += va * va
		normB += vb * vb
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (h *HybridSearch) rrfFusion(a, b []SearchResult, limit int) []SearchResult {
	scoreMap := make(map[string]*SearchResult)

	for i := range a {
		scoreMap[a[i].ID] = &a[i]
	}

	for i := range b {
		if existing, ok := scoreMap[b[i].ID]; ok {
			existing.Score += b[i].Score
			existing.Source = "hybrid"
		} else {
			scoreMap[b[i].ID] = &b[i]
		}
	}

	var results []SearchResult
	for _, r := range scoreMap {
		results = append(results, *r)
	}

	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (h *HybridSearch) boostCompiledTruth(results []SearchResult) []SearchResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := range results {
		var compiled bool
		h.db.QueryRow("SELECT compiled_truth FROM documents WHERE id = ?", results[i].ID).Scan(&compiled)
		if compiled {
			results[i].Score *= 1.5
		}
	}

	return results
}

func (h *HybridSearch) boostBacklinks(results []SearchResult) []SearchResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for i := range results {
		var count int
		h.db.QueryRow(`
			SELECT COUNT(*) FROM page_entities WHERE entity_id IN 
			(SELECT entity_id FROM page_entities WHERE page_id = 
				(SELECT rowid FROM documents WHERE id = ?))
		`, results[i].ID).Scan(&count)

		results[i].BacklinkCount = count
		results[i].Score *= (1.0 + float64(count)*0.1)
	}

	return results
}

func (h *HybridSearch) deduplicate(results []SearchResult, limit int) []SearchResult {
	seen := make(map[string]bool)
	var unique []SearchResult

	for _, r := range results {
		if !seen[r.ID] {
			seen[r.ID] = true
			unique = append(unique, r)
		}
		if len(unique) >= limit {
			break
		}
	}

	return unique
}

func (h *HybridSearch) GetStats() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]int)

	var total int
	h.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&total)
	stats["total"] = total

	var withEmbedding int
	h.db.QueryRow("SELECT COUNT(*) FROM documents WHERE embedding IS NOT NULL").Scan(&withEmbedding)
	stats["with_vector"] = withEmbedding

	var compiled int
	h.db.QueryRow("SELECT COUNT(*) FROM documents WHERE compiled_truth = 1").Scan(&compiled)
	stats["compiled_truth"] = compiled

	return stats
}

func (h *HybridSearch) Close() error {
	return h.db.Close()
}
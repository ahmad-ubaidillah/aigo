// Package memory provides the common interface and wrapper for memory backends.
package memory

import (
	"path/filepath"

	"github.com/hermes-v2/aigo/internal/memory/engram"
	"github.com/hermes-v2/aigo/internal/memory/engram_store"
	"github.com/hermes-v2/aigo/internal/memory/fts5pkg"
)

// Backend is the common interface for all memory backends.
type Backend interface {
	Context() string
	SaveDaily(content string) error
	SaveLongTerm(content string) error
	Search(query string, limit int) ([]SearchResult, error)
}

// SearchResult is a search result from any memory backend.
type SearchResult struct {
	Content  string
	Category string
	Score    float64
}

// FTS5Wrapper wraps fts5pkg.Store to satisfy Backend interface.
type FTS5Wrapper struct {
	store *fts5pkg.Store
}

// NewFTS5Backend creates a new FTS5-backed memory backend.
func NewFTS5Backend(basePath string) (Backend, error) {
	store, err := fts5pkg.New(basePath)
	if err != nil {
		return nil, err
	}
	return &FTS5Wrapper{store: store}, nil
}

func (w *FTS5Wrapper) Context() string {
	return w.store.Context()
}

func (w *FTS5Wrapper) SaveDaily(content string) error {
	_, err := w.store.SaveDaily(content)
	return err
}

func (w *FTS5Wrapper) SaveLongTerm(content string) error {
	_, err := w.store.SaveLongTerm(content)
	return err
}

func (w *FTS5Wrapper) Search(query string, limit int) ([]SearchResult, error) {
	entries, err := w.store.Search(query, limit)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, e := range entries {
		results = append(results, SearchResult{
			Content:  e.Content,
			Category: e.Category,
			Score:    e.Score,
		})
	}
	return results, nil
}

// FileBackend wraps the existing file-based Store to satisfy Backend interface.
type FileBackend struct {
	store *Store
}

// NewFileBackend creates a new file-based memory backend.
func NewFileBackend(basePath string) (Backend, error) {
	store, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &FileBackend{store: store}, nil
}

func (w *FileBackend) Context() string {
	return w.store.Context()
}

func (w *FileBackend) SaveDaily(content string) error {
	return w.store.SaveDaily(content)
}

func (w *FileBackend) SaveLongTerm(content string) error {
	return w.store.SaveLongTerm(content)
}

func (w *FileBackend) Search(query string, limit int) ([]SearchResult, error) {
	lines, err := w.store.Search(query, limit)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, line := range lines {
		results = append(results, SearchResult{
			Content:  line,
			Category: "mixed",
			Score:    0,
		})
	}
	return results, nil
}

// NewBackend creates the appropriate memory backend based on config.
func NewBackend(basePath string, useFTS5 bool) (Backend, error) {
	if useFTS5 {
		return NewFTS5Backend(basePath)
	}
	return NewFileBackend(filepath.Join(basePath))
}

// EngramBackend wraps the engram backend to satisfy memory.Backend.
// It also exposes engram-specific methods via the underlying *engram.Backend.
type EngramBackend struct {
	b *engram.Backend
}

// NewEngramBackend creates a new engram-backed memory backend.
func NewEngramBackend(dataDir, project string) (*EngramBackend, error) {
	b, err := engram.New(dataDir, project)
	if err != nil {
		return nil, err
	}
	return &EngramBackend{b: b}, nil
}

// Backend returns the underlying engram.Backend for direct access.
func (e *EngramBackend) Backend() *engram.Backend {
	return e.b
}

func (e *EngramBackend) Context() string {
	ctx, err := e.b.GetContext()
	if err != nil || ctx == "" {
		return ""
	}
	return "<memory>\n" + ctx + "\n</memory>"
}

func (e *EngramBackend) SaveDaily(content string) error {
	_, err := e.b.SaveObservation("daily", "Daily Note", content, "")
	return err
}

func (e *EngramBackend) SaveLongTerm(content string) error {
	_, err := e.b.SaveObservation("longterm", "Long-term Note", content, "")
	return err
}

func (e *EngramBackend) Search(query string, limit int) ([]SearchResult, error) {
	results, err := e.b.Search(query, limit)
	if err != nil {
		return nil, err
	}
	var out []SearchResult
	for _, r := range results {
		content := r.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		out = append(out, SearchResult{
			Content:  r.Title + ": " + content,
			Category: r.Type,
			Score:    r.Rank,
		})
	}
	return out, nil
}

// RecentSearchResults returns engram SearchResult directly (with full metadata).
func (e *EngramBackend) RecentSearchResults(query string, limit int) ([]engram_store.SearchResult, error) {
	return e.b.Search(query, limit)
}

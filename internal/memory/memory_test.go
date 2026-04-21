package memory

import (
	"strings"
	"testing"
)

func TestFileBackend(t *testing.T) {
	tmpDir := t.TempDir()
	backend, err := NewFileBackend(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Save daily
	if err := backend.SaveDaily("test daily note"); err != nil {
		t.Fatal(err)
	}

	// Context should include daily
	ctx := backend.Context()
	if !strings.Contains(ctx, "Today's Notes") {
		t.Error("expected 'Today's Notes' in context")
	}
	if !strings.Contains(ctx, "test daily note") {
		t.Error("expected 'test daily note' in context")
	}

	// Save long-term
	if err := backend.SaveLongTerm("important fact"); err != nil {
		t.Fatal(err)
	}

	ctx = backend.Context()
	if !strings.Contains(ctx, "Long-term Memory") {
		t.Error("expected 'Long-term Memory' in context")
	}
	if !strings.Contains(ctx, "important fact") {
		t.Error("expected 'important fact' in context")
	}

	// Search
	results, err := backend.Search("important", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected search results")
	}
}

func TestFileBackendEmptyContext(t *testing.T) {
	tmpDir := t.TempDir()
	backend, err := NewFileBackend(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := backend.Context()
	if ctx != "" {
		t.Errorf("expected empty context, got '%s'", ctx)
	}
}

func TestFTS5Backend(t *testing.T) {
	tmpDir := t.TempDir()
	backend, err := NewFTS5Backend(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Save daily
	if err := backend.SaveDaily("meeting with team at 3pm"); err != nil {
		t.Fatal(err)
	}
	if err := backend.SaveDaily("reviewed code changes"); err != nil {
		t.Fatal(err)
	}

	// Save long-term
	if err := backend.SaveLongTerm("user prefers dark theme"); err != nil {
		t.Fatal(err)
	}

	// Context
	ctx := backend.Context()
	if !strings.Contains(ctx, "Today's Notes") {
		t.Error("expected 'Today's Notes'")
	}
	if !strings.Contains(ctx, "meeting with team") {
		t.Error("expected 'meeting with team'")
	}
	if !strings.Contains(ctx, "Long-term Memory") {
		t.Error("expected 'Long-term Memory'")
	}
	if !strings.Contains(ctx, "dark theme") {
		t.Error("expected 'dark theme'")
	}

	// Search
	results, err := backend.Search("team", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Error("expected search results for 'team'")
	}
	found := false
	for _, r := range results {
		if strings.Contains(r.Content, "meeting with team") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'meeting with team' in search results")
	}
}

func TestFTS5SearchCategory(t *testing.T) {
	tmpDir := t.TempDir()
	backend, err := NewFTS5Backend(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	backend.SaveDaily("daily note about python")
	backend.SaveLongTerm("long term note about python")

	// Search all
	results, err := backend.Search("python", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestNewBackendFile(t *testing.T) {
	tmpDir := t.TempDir()
	backend, err := NewBackend(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	backend.SaveDaily("test")
	ctx := backend.Context()
	if !strings.Contains(ctx, "test") {
		t.Error("file backend should work")
	}
}

func TestNewBackendFTS5(t *testing.T) {
	tmpDir := t.TempDir()
	backend, err := NewBackend(tmpDir, true)
	if err != nil {
		t.Fatal(err)
	}

	backend.SaveDaily("test fts5")
	ctx := backend.Context()
	if !strings.Contains(ctx, "test fts5") {
		t.Error("FTS5 backend should work")
	}
}

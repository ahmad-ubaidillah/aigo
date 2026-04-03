package rewind

import (
	"testing"
)

func TestRewindStore_StoreAndRetrieve(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	hash := s.Store("hello world", "text/plain", "sess-1")
	if len(hash) != 8 {
		t.Errorf("expected 8-char hash, got %q (len=%d)", hash, len(hash))
	}

	entry, err := s.Retrieve(hash)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Content != "hello world" {
		t.Errorf("expected 'hello world', got %q", entry.Content)
	}
	if entry.OriginalSize != 11 {
		t.Errorf("expected size 11, got %d", entry.OriginalSize)
	}
}

func TestRewindStore_RetrieveNotFound(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	_, err := s.Retrieve("nonexist")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRewindStore_List(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	s.Store("content1", "text", "sess-1")
	s.Store("content2", "text", "sess-1")
	s.Store("content3", "text", "sess-2")

	all := s.List("")
	if len(all) != 3 {
		t.Errorf("expected 3 entries, got %d", len(all))
	}

	sess1 := s.List("sess-1")
	if len(sess1) != 2 {
		t.Errorf("expected 2 entries for sess-1, got %d", len(sess1))
	}
}

func TestRewindStore_Count(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	if s.Count() != 0 {
		t.Error("expected 0")
	}
	s.Store("a", "text", "s1")
	s.Store("b", "text", "s1")
	if s.Count() != 2 {
		t.Errorf("expected 2, got %d", s.Count())
	}
}

func TestRewindStore_Clear(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	s.Store("a", "text", "s1")
	s.Store("b", "text", "s1")
	s.Clear()
	if s.Count() != 0 {
		t.Errorf("expected 0 after clear, got %d", s.Count())
	}
}

func TestRewindStore_Show(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	hash := s.Store("secret content", "text", "s1")
	content, err := s.Show(hash)
	if err != nil {
		t.Fatal(err)
	}
	if content != "secret content" {
		t.Errorf("expected 'secret content', got %q", content)
	}
}

func TestRewindStore_ShowNotFound(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	_, err := s.Show("nope")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRewindStore_DeterministicHash(t *testing.T) {
	t.Parallel()

	s := NewRewindStore()
	h1 := s.Store("same content", "text", "s1")
	h2 := s.Store("same content", "text", "s1")
	if h1 != h2 {
		t.Errorf("expected same hash, got %q and %q", h1, h2)
	}
}

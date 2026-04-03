package memory

import (
	"testing"
)

func TestFactExtractor_ExtractFact(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	fact := e.ExtractFact("random conversation")
	if fact.Action != FactNone {
		t.Errorf("expected NONE, got %s", fact.Action)
	}
}

func TestFactExtractor_AddFact(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	e.AddFact("user prefers dark mode", "chat", "user1", "agent1")

	if e.Count() != 1 {
		t.Errorf("expected 1, got %d", e.Count())
	}

	facts := e.ListFacts()
	if facts[0].Content != "user prefers dark mode" {
		t.Errorf("unexpected content: %s", facts[0].Content)
	}
	if facts[0].Action != FactAdd {
		t.Errorf("expected ADD, got %s", facts[0].Action)
	}
}

func TestFactExtractor_GetFacts(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	e.AddFact("fact1", "chat", "user1", "agent1")
	e.AddFact("fact2", "chat", "user2", "agent1")

	facts := e.GetFacts("user1")
	if len(facts) != 1 {
		t.Errorf("expected 1, got %d", len(facts))
	}
}

func TestFactExtractor_UpdateFact(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	e.AddFact("old content", "chat", "user1", "agent1")
	facts := e.ListFacts()

	err := e.UpdateFact(facts[0].ID, "new content")
	if err != nil {
		t.Fatal(err)
	}

	facts = e.ListFacts()
	if facts[0].Content != "new content" {
		t.Errorf("expected 'new content', got %s", facts[0].Content)
	}
	if facts[0].Action != FactUpdate {
		t.Errorf("expected UPDATE, got %s", facts[0].Action)
	}
}

func TestFactExtractor_UpdateFactNotFound(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	err := e.UpdateFact("nonexistent", "content")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFactExtractor_DeleteFact(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	e.AddFact("to delete", "chat", "user1", "agent1")
	facts := e.ListFacts()

	err := e.DeleteFact(facts[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if e.Count() != 0 {
		t.Errorf("expected 0, got %d", e.Count())
	}
}

func TestFactExtractor_DeleteFactNotFound(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	err := e.DeleteFact("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFactExtractor_ListFacts(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	e.AddFact("fact1", "chat", "u1", "a1")
	e.AddFact("fact2", "chat", "u1", "a1")

	facts := e.ListFacts()
	if len(facts) != 2 {
		t.Errorf("expected 2, got %d", len(facts))
	}
}

func TestFactExtractor_Count(t *testing.T) {
	t.Parallel()

	e := NewFactExtractor(nil, "")
	if e.Count() != 0 {
		t.Error("expected 0")
	}
	e.AddFact("f1", "chat", "u1", "a1")
	if e.Count() != 1 {
		t.Errorf("expected 1, got %d", e.Count())
	}
}

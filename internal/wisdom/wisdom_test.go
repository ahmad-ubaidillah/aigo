package wisdom

import (
	"strings"
	"testing"
)

func TestWisdomStore_AddLearning(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	id := w.AddLearning("task1", "always test", "testing")
	if id == "" {
		t.Error("expected ID")
	}
}

func TestWisdomStore_GetLearnings(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	w.AddLearning("auth task", "use JWT", "security")
	w.AddLearning("db task", "use transactions", "database")

	results := w.GetLearnings("auth")
	if len(results) != 1 {
		t.Errorf("expected 1, got %d", len(results))
	}
}

func TestWisdomStore_GetPatterns(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	w.AddLearning("task1", "lesson1", "pattern-a")
	w.AddLearning("task2", "lesson2", "pattern-a")
	w.AddLearning("task3", "lesson3", "pattern-b")

	patterns := w.GetPatterns()
	if len(patterns) != 2 {
		t.Errorf("expected 2 unique patterns, got %d", len(patterns))
	}
}

func TestWisdomStore_FindRelevant(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	w.AddLearning("auth task", "use JWT for auth", "security")
	w.AddLearning("db task", "use connection pooling", "database")

	results := w.FindRelevant("auth", 5)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if !strings.Contains(results[0].Lesson, "JWT") {
		t.Errorf("expected JWT lesson, got %s", results[0].Lesson)
	}
}

func TestWisdomStore_FindRelevant_TopK(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	w.AddLearning("task1", "lesson1", "pattern1")
	w.AddLearning("task2", "lesson2", "pattern2")
	w.AddLearning("task3", "lesson3", "pattern3")

	results := w.FindRelevant("", 2)
	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}
}

func TestWisdomStore_InjectWisdom(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	w.AddLearning("auth", "always validate tokens", "security")

	wisdom := w.InjectWisdom("auth")
	if wisdom == "" {
		t.Error("expected wisdom")
	}
	if !strings.Contains(wisdom, "validate tokens") {
		t.Errorf("expected token validation, got %s", wisdom)
	}
}

func TestWisdomStore_InjectWisdom_None(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	wisdom := w.InjectWisdom("unrelated")
	if wisdom != "" {
		t.Errorf("expected empty, got %s", wisdom)
	}
}

func TestWisdomStore_Clear(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	w.AddLearning("task1", "lesson1", "pattern1")
	w.Clear()
	if w.Count() != 0 {
		t.Errorf("expected 0, got %d", w.Count())
	}
}

func TestWisdomStore_Count(t *testing.T) {
	t.Parallel()

	w := NewWisdomStore()
	if w.Count() != 0 {
		t.Error("expected 0")
	}
	w.AddLearning("t1", "l1", "p1")
	w.AddLearning("t2", "l2", "p2")
	if w.Count() != 2 {
		t.Errorf("expected 2, got %d", w.Count())
	}
}

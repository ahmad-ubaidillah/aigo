package workers

import (
	"context"
	"testing"
)

func TestWorkerPool_RegisterAndGet(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	pool.Register(NewSisyphus())

	w := pool.Get("sisyphus")
	if w == nil {
		t.Fatal("expected sisyphus")
	}
	if w.Name() != "sisyphus" {
		t.Errorf("expected sisyphus, got %s", w.Name())
	}
}

func TestWorkerPool_GetNotFound(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	if pool.Get("nonexistent") != nil {
		t.Error("expected nil")
	}
}

func TestWorkerPool_List(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	pool.Register(NewSisyphus())
	pool.Register(NewHephaestus())
	pool.Register(NewOracle())

	names := pool.List()
	if len(names) != 3 {
		t.Errorf("expected 3 workers, got %d", len(names))
	}
}

func TestWorkerPool_Execute(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	pool.Register(NewSisyphus())

	result, err := pool.Execute(context.Background(), "sisyphus", "test task", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Metadata["task"] != "test task" {
		t.Errorf("expected task in metadata, got %v", result.Metadata)
	}
}

func TestWorkerPool_ExecuteNotFound(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	_, err := pool.Execute(context.Background(), "nonexistent", "task", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestSisyphus_Name(t *testing.T) {
	t.Parallel()

	w := NewSisyphus()
	if w.Name() != "sisyphus" {
		t.Errorf("expected sisyphus, got %s", w.Name())
	}
}

func TestHephaestus_Name(t *testing.T) {
	t.Parallel()

	w := NewHephaestus()
	if w.Name() != "hephaestus" {
		t.Errorf("expected hephaestus, got %s", w.Name())
	}
}

func TestOracle_Name(t *testing.T) {
	t.Parallel()

	w := NewOracle()
	if w.Name() != "oracle" {
		t.Errorf("expected oracle, got %s", w.Name())
	}
}

func TestLibrarian_Name(t *testing.T) {
	t.Parallel()

	w := NewLibrarian()
	if w.Name() != "librarian" {
		t.Errorf("expected librarian, got %s", w.Name())
	}
}

func TestExplore_Name(t *testing.T) {
	t.Parallel()

	w := NewExplore()
	if w.Name() != "explore" {
		t.Errorf("expected explore, got %s", w.Name())
	}
}

func TestHephaestus_HashAnchoredEdit(t *testing.T) {
	t.Parallel()

	w := NewHephaestus()
	err := w.HashAnchoredEdit("file.go", "old", "new")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestHephaestus_Execute(t *testing.T) {
	t.Parallel()

	w := NewHephaestus()
	result, err := w.Execute(context.Background(), "fix bug", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Metadata["task"] != "fix bug" {
		t.Errorf("expected 'fix bug', got %s", result.Metadata["task"])
	}
}

func TestOracle_Execute(t *testing.T) {
	t.Parallel()

	w := NewOracle()
	result, err := w.Execute(context.Background(), "analyze", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestLibrarian_Execute(t *testing.T) {
	t.Parallel()

	w := NewLibrarian()
	result, err := w.Execute(context.Background(), "search docs", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestExplore_Execute(t *testing.T) {
	t.Parallel()

	w := NewExplore()
	result, err := w.Execute(context.Background(), "find patterns", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Error("expected success")
	}
}

func TestWorkerPool_RegisterDuplicate(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	pool.Register(NewSisyphus())
	pool.Register(NewSisyphus())
	names := pool.List()
	if len(names) != 1 {
		t.Errorf("expected 1 worker after duplicate register, got %d", len(names))
	}
}

func TestWorkerPool_GetAfterRegister(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool()
	pool.Register(NewSisyphus())
	w := pool.Get("sisyphus")
	if w == nil {
		t.Error("expected sisyphus after register")
	}
}

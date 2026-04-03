package execution

import (
	"testing"
)

func TestAtlas_AddTodo(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	id := a.AddTodo("write tests")
	if id == "" {
		t.Error("expected todo ID")
	}
}

func TestAtlas_UpdateTodo(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	id := a.AddTodo("write tests")
	err := a.UpdateTodo(id, TodoDone, "all pass")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtlas_UpdateTodoNotFound(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	err := a.UpdateTodo("nonexistent", TodoDone, "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestAtlas_ListTodos(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	a.AddTodo("task1")
	a.AddTodo("task2")
	todos := a.ListTodos()
	if len(todos) != 2 {
		t.Errorf("expected 2, got %d", len(todos))
	}
}

func TestAtlas_NextPending(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	a.AddTodo("pending task")
	todo := a.NextPending()
	if todo == nil {
		t.Fatal("expected pending todo")
	}
	if todo.Status != TodoPending {
		t.Errorf("expected pending, got %s", todo.Status)
	}
}

func TestAtlas_NextPending_None(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	id := a.AddTodo("done task")
	a.UpdateTodo(id, TodoDone, "")
	todo := a.NextPending()
	if todo != nil {
		t.Error("expected nil")
	}
}

func TestAtlas_Wisdom(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	a.AddWisdom("task-1", []string{"always test", "use interfaces"})
	w := a.GetWisdom("task-1")
	if w == nil {
		t.Fatal("expected wisdom")
	}
	if len(w.Learnings) != 2 {
		t.Errorf("expected 2 learnings, got %d", len(w.Learnings))
	}
}

func TestAtlas_WisdomNotFound(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	w := a.GetWisdom("nonexistent")
	if w != nil {
		t.Error("expected nil")
	}
}

func TestAtlas_GetProgress(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	a.AddTodo("task1")
	a.AddTodo("task2")
	progress := a.GetProgress()
	if progress != "0/2 todos completed" {
		t.Errorf("expected '0/2 todos completed', got %q", progress)
	}

	id := a.AddTodo("task3")
	a.UpdateTodo(id, TodoDone, "")
	progress = a.GetProgress()
	if progress != "1/3 todos completed" {
		t.Errorf("expected '1/3 todos completed', got %q", progress)
	}
}

func TestAtlas_IsComplete(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	if a.IsComplete() {
		t.Error("expected false for empty atlas")
	}

	a.AddTodo("task1")
	if a.IsComplete() {
		t.Error("expected false with pending tasks")
	}

	todos := a.ListTodos()
	for _, t := range todos {
		a.UpdateTodo(t.ID, TodoDone, "")
	}
	if !a.IsComplete() {
		t.Error("expected true when all done")
	}
}

func TestAtlas_Reset(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	a.AddTodo("task1")
	a.AddTodo("task2")
	a.AddWisdom("task-1", []string{"learned"})
	a.Reset()
	if len(a.ListTodos()) != 0 {
		t.Error("expected 0 todos after reset")
	}
	if a.GetWisdom("task-1") != nil {
		t.Error("expected nil wisdom after reset")
	}
}

func TestAtlas_SetTask(t *testing.T) {
	t.Parallel()

	a := NewAtlas()
	a.SetTask("build feature X")
	if a.currentTask != "build feature X" {
		t.Errorf("expected 'build feature X', got %q", a.currentTask)
	}
}

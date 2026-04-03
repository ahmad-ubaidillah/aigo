package execution

import (
	"fmt"
	"sync"
	"time"
)

const (
	TodoPending    = "pending"
	TodoInProgress = "in_progress"
	TodoDone       = "done"
	TodoCancelled  = "cancelled"
)

type TodoItem struct {
	ID       string
	Content  string
	Status   string
	Assignee string
	Result   string
}

type Wisdom struct {
	TaskID    string
	Learnings []string
	Patterns  []string
	Timestamp time.Time
}

type Atlas struct {
	todos       []TodoItem
	wisdom      map[string]*Wisdom
	mu          sync.RWMutex
	currentTask string
	nextID      int
}

func NewAtlas() *Atlas {
	return &Atlas{
		todos:  make([]TodoItem, 0),
		wisdom: make(map[string]*Wisdom),
	}
}

func (a *Atlas) SetTask(task string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentTask = task
}

func (a *Atlas) AddTodo(content string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nextID++
	id := fmt.Sprintf("todo-%d", a.nextID)
	a.todos = append(a.todos, TodoItem{
		ID:      id,
		Content: content,
		Status:  TodoPending,
	})
	return id
}

func (a *Atlas) UpdateTodo(id, status, result string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.todos {
		if a.todos[i].ID == id {
			a.todos[i].Status = status
			a.todos[i].Result = result
			return nil
		}
	}
	return fmt.Errorf("todo %s not found", id)
}

func (a *Atlas) ListTodos() []TodoItem {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]TodoItem, len(a.todos))
	copy(out, a.todos)
	return out
}

func (a *Atlas) NextPending() *TodoItem {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for i := range a.todos {
		if a.todos[i].Status == TodoPending {
			return &a.todos[i]
		}
	}
	return nil
}

func (a *Atlas) AddWisdom(taskID string, learnings []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.wisdom[taskID] = &Wisdom{
		TaskID:    taskID,
		Learnings: learnings,
		Timestamp: time.Now(),
	}
}

func (a *Atlas) GetWisdom(taskID string) *Wisdom {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.wisdom[taskID]
}

func (a *Atlas) GetProgress() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	done := 0
	for _, t := range a.todos {
		if t.Status == TodoDone || t.Status == TodoCancelled {
			done++
		}
	}
	return fmt.Sprintf("%d/%d todos completed", done, len(a.todos))
}

func (a *Atlas) IsComplete() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, t := range a.todos {
		if t.Status != TodoDone && t.Status != TodoCancelled {
			return false
		}
	}
	return len(a.todos) > 0
}

func (a *Atlas) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.todos = make([]TodoItem, 0)
	a.wisdom = make(map[string]*Wisdom)
	a.nextID = 0
}

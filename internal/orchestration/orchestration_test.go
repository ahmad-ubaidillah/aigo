package orchestration

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func testOrchestrator(t *testing.T) *Orchestrator {
	t.Helper()
	ctx := context.Background()
	config := OrchestratorConfig{MaxAgents: 5}
	factory := func(ctx context.Context, id string) (any, error) { return id, nil }
	return NewOrchestrator(ctx, config, factory)
}

func TestOrchestrator_ExecutePlan(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	steps := []ExecutionTask{
		{ID: "t1", Description: "task 1", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
		{ID: "t2", Description: "task 2", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
	}
	results, err := o.ExecutePlan(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestOrchestrator_ExecuteParallel(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	tasks := []ExecutionTask{
		{ID: "t1", Description: "task 1", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
		{ID: "t2", Description: "task 2", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
	}
	results, err := o.ExecuteParallel(context.Background(), tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestOrchestrator_ExecuteWithRetry(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	task := ExecutionTask{ID: "t1", Description: "retry test", Execute: func(ctx context.Context) (any, error) { return "done", nil }}
	result, err := o.ExecuteWithRetry(context.Background(), task, 3)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected result")
	}
}

func TestOrchestrator_SpawnAgent(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	id, err := o.SpawnAgent(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Error("expected agent ID")
	}
}

func TestOrchestrator_TerminateAgent(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	id, _ := o.SpawnAgent(context.Background())
	err := o.TerminateAgent(id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrchestrator_SendMessage(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	err := o.SendMessage(Message{From: "a", To: "b", Type: "test", Payload: "data"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrchestrator_Broadcast(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	err := o.Broadcast(Message{From: "a", Type: "broadcast", Payload: "all"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrchestrator_GetResults(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	ch := o.GetResults()
	if ch == nil {
		t.Error("expected results channel")
	}
}

func TestOrchestrator_GetStats(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	stats := o.GetStats()
	if stats["total_tasks"] == nil {
		t.Error("expected total_tasks")
	}
}

func TestOrchestrator_Shutdown(t *testing.T) {
	t.Parallel()
	o := testOrchestrator(t)
	err := o.Shutdown()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMessageBus_Subscribe(t *testing.T) {
	t.Parallel()
	bus := NewMessageBus(10)
	bus.Subscribe("agent1", func(msg Message) {})
	if len(bus.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(bus.handlers))
	}
}

func TestMessageBus_Send(t *testing.T) {
	t.Parallel()
	bus := NewMessageBus(10)
	var received atomic.Bool
	bus.Subscribe("b", func(msg Message) { received.Store(true) })
	err := bus.Send(Message{From: "a", To: "b", Type: "test", Payload: "data"})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if !received.Load() {
		t.Error("expected message received")
	}
}

func TestMessageBus_Broadcast(t *testing.T) {
	t.Parallel()
	bus := NewMessageBus(10)
	var count atomic.Int32
	bus.Subscribe("a", func(msg Message) { count.Add(1) })
	bus.Subscribe("b", func(msg Message) { count.Add(1) })
	err := bus.Broadcast(Message{Type: "all", Payload: "data"})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if count.Load() != 2 {
		t.Errorf("expected 2 broadcasts, got %d", count.Load())
	}
}

func TestMessageBus_Unsubscribe(t *testing.T) {
	t.Parallel()
	bus := NewMessageBus(10)
	bus.Subscribe("agent1", func(msg Message) {})
	bus.Unsubscribe("agent1")
	if len(bus.handlers) != 0 {
		t.Errorf("expected 0 handlers, got %d", len(bus.handlers))
	}
}

func TestSequentialStrategy_Execute(t *testing.T) {
	t.Parallel()
	s := NewSequentialStrategy()
	tasks := []ExecutionTask{
		{ID: "t1", Description: "task 1", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
		{ID: "t2", Description: "task 2", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
	}
	results, err := s.Execute(context.Background(), tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSequentialStrategy_Name(t *testing.T) {
	t.Parallel()
	s := NewSequentialStrategy()
	if s.Name() != "sequential" {
		t.Errorf("expected sequential, got %s", s.Name())
	}
}

func TestParallelStrategy_Execute(t *testing.T) {
	t.Parallel()
	s := NewParallelStrategy(2)
	tasks := []ExecutionTask{
		{ID: "t1", Description: "task 1", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
		{ID: "t2", Description: "task 2", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
	}
	results, err := s.Execute(context.Background(), tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestParallelStrategy_Name(t *testing.T) {
	t.Parallel()
	s := NewParallelStrategy(4)
	if s.Name() != "parallel" {
		t.Errorf("expected parallel, got %s", s.Name())
	}
}

func TestRetryStrategy_Execute(t *testing.T) {
	t.Parallel()
	s := NewRetryStrategy(NewSequentialStrategy())
	tasks := []ExecutionTask{
		{ID: "t1", Description: "task 1", Execute: func(ctx context.Context) (any, error) { return "done", nil }},
	}
	results, err := s.Execute(context.Background(), tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRetryStrategy_WithMaxRetries(t *testing.T) {
	t.Parallel()
	s := NewRetryStrategy(NewSequentialStrategy(), WithMaxRetries(10))
	if s.maxRetries != 10 {
		t.Errorf("expected 10, got %d", s.maxRetries)
	}
}

func TestExecutionTask(t *testing.T) {
	t.Parallel()
	task := ExecutionTask{
		ID:          "t1",
		Description: "test",
		DependsOn:   []string{"d1"},
	}
	if task.ID != "t1" {
		t.Errorf("expected t1, got %s", task.ID)
	}
}

func TestExecutionResult(t *testing.T) {
	t.Parallel()
	r := ExecutionResult{TaskID: "t1"}
	if r.TaskID != "t1" {
		t.Errorf("expected t1, got %s", r.TaskID)
	}
}

func TestMessage(t *testing.T) {
	t.Parallel()
	m := Message{From: "a", To: "b", Type: "test", Payload: "data"}
	if m.From != "a" {
		t.Errorf("expected a, got %s", m.From)
	}
}

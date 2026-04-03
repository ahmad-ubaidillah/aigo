package fleet

import (
	"testing"
	"time"
)

func TestAgentConfig(t *testing.T) {
	t.Parallel()
	cfg := AgentConfig{
		Name:       "test-agent",
		Role:       "worker",
		Priority:   1,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}
	if cfg.Name != "test-agent" {
		t.Errorf("expected test-agent, got %s", cfg.Name)
	}
	if cfg.Priority != 1 {
		t.Errorf("expected 1, got %d", cfg.Priority)
	}
}

func TestAgentState(t *testing.T) {
	t.Parallel()
	state := AgentState{
		Name:        "test",
		Status:      "running",
		CurrentTask: "task-1",
		Retries:     0,
	}
	if state.Name != "test" {
		t.Errorf("expected test, got %s", state.Name)
	}
	if state.Status != "running" {
		t.Errorf("expected running, got %s", state.Status)
	}
}

func TestTaskResult(t *testing.T) {
	t.Parallel()
	result := TaskResult{
		Agent:    "test-agent",
		Success:  true,
		Output:   "done",
		Retries:  0,
		Duration: time.Second,
	}
	if !result.Success {
		t.Error("expected success")
	}
	if result.Agent != "test-agent" {
		t.Errorf("expected test-agent, got %s", result.Agent)
	}
}

func TestFleetTask(t *testing.T) {
	t.Parallel()
	task := FleetTask{
		ID:          "task-1",
		Description: "test task",
		Priority:    1,
	}
	if task.ID != "task-1" {
		t.Errorf("expected task-1, got %s", task.ID)
	}
}

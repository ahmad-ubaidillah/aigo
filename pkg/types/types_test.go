package types

import (
	"testing"
)

func TestSession(t *testing.T) {
	t.Parallel()
	s := Session{ID: "s1", Name: "test", Workspace: "/tmp"}
	if s.ID != "s1" {
		t.Errorf("expected s1, got %s", s.ID)
	}
}

func TestMessage(t *testing.T) {
	t.Parallel()
	m := Message{SessionID: "s1", Role: "user", Content: "hello"}
	if m.Role != "user" {
		t.Errorf("expected user, got %s", m.Role)
	}
}

func TestTask(t *testing.T) {
	t.Parallel()
	task := Task{ID: 1, Description: "test"}
	if task.Description != "test" {
		t.Errorf("expected test, got %s", task.Description)
	}
}

func TestMemory(t *testing.T) {
	t.Parallel()
	m := Memory{ID: 1, Content: "fact", Category: "preference"}
	if m.Category != "preference" {
		t.Errorf("expected preference, got %s", m.Category)
	}
}

func TestProfile(t *testing.T) {
	t.Parallel()
	p := Profile{ID: "p1", Name: "default", DefaultModel: "gpt-4"}
	if p.DefaultModel != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", p.DefaultModel)
	}
}

func TestToolResult(t *testing.T) {
	t.Parallel()
	r := ToolResult{Success: true, Output: "done"}
	if !r.Success {
		t.Error("expected success")
	}
}

func TestConfig(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Model:    ModelConfig{Default: "gpt-4", Coding: "claude-3"},
		OpenCode: OpenCodeConfig{Binary: "opencode", Timeout: 30, MaxTurns: 50},
	}
	if cfg.Model.Default != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", cfg.Model.Default)
	}
}

func TestSkill(t *testing.T) {
	t.Parallel()
	s := Skill{ID: "s1", Name: "test", Version: 1, Enabled: true}
	if !s.Enabled {
		t.Error("expected enabled")
	}
}

func TestGatewayConfig(t *testing.T) {
	t.Parallel()
	g := GatewayConfig{Enabled: true, Platforms: []string{"telegram"}}
	if !g.Enabled {
		t.Error("expected enabled")
	}
}

func TestIntentConstants(t *testing.T) {
	t.Parallel()
	if IntentCoding != "coding" {
		t.Errorf("expected coding, got %s", IntentCoding)
	}
	if IntentWeb != "web" {
		t.Errorf("expected web, got %s", IntentWeb)
	}
}

func TestResearchQuery(t *testing.T) {
	t.Parallel()
	rq := ResearchQuery{ID: "rq1", Query: "test", Sources: []string{"web"}}
	if rq.Query != "test" {
		t.Errorf("expected test, got %s", rq.Query)
	}
}

func TestMemoryConfig(t *testing.T) {
	t.Parallel()
	c := MemoryConfig{MaxL0Items: 20, MaxL1Items: 50, TokenBudget: 8000}
	if c.MaxL0Items != 20 {
		t.Errorf("expected 20, got %d", c.MaxL0Items)
	}
}

func TestOpenCodeConfig(t *testing.T) {
	t.Parallel()
	c := OpenCodeConfig{Binary: "opencode", Timeout: 30, MaxTurns: 50}
	if c.Timeout != 30 {
		t.Errorf("expected 30, got %d", c.Timeout)
	}
}

func TestModelConfig(t *testing.T) {
	t.Parallel()
	c := ModelConfig{Default: "gpt-4", Coding: "claude-3", Intent: "gpt-4o-mini"}
	if c.Default != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", c.Default)
	}
}

func TestWebConfig(t *testing.T) {
	t.Parallel()
	c := WebConfig{Enabled: true, Port: ":8080"}
	if !c.Enabled {
		t.Error("expected enabled")
	}
}

func TestCronSchedule(t *testing.T) {
	t.Parallel()
	cs := CronSchedule{ID: 1, Name: "test", Schedule: "*/5 * * * *", Command: "echo test"}
	if cs.Name != "test" {
		t.Errorf("expected test, got %s", cs.Name)
	}
}

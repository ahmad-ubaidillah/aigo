package planning

import (
	"testing"
)

func TestContextBoost_HotFiles(t *testing.T) {
	c := NewContextBoost()
	c.AddHotFile("agent.go")
	c.AddHotFile("planner.go")

	score := c.GetBoostScore("fix agent.go bug")
	if score == 0 {
		t.Error("Should detect hot file in input")
	}
	t.Logf("Hot file boost score: %d", score)
}

func TestContextBoost_RecentCommands(t *testing.T) {
	c := NewContextBoost()
	c.AddRecentCommand("git commit")
	c.AddRecentCommand("go test")

	score := c.GetBoostScore("go test run")
	if score == 0 {
		t.Error("Should detect recent command in input")
	}
	t.Logf("Recent command boost score: %d", score)
}

func TestContextBoost_DomainHints(t *testing.T) {
	c := NewContextBoost()
	c.AddDomainHint("planning")
	c.AddDomainHint("agent")

	score := c.GetBoostScore("improve planning")
	if score == 0 {
		t.Error("Should detect domain hint in input")
	}
	t.Logf("Domain hint boost score: %d", score)
}

func TestContextBoost_GetBoostedPrompt(t *testing.T) {
	c := NewContextBoost()
	c.AddHotFile("agent.go")
	c.AddRecentCommand("go build")
	c.AddDomainHint("planning")

	prompt := c.GetBoostedPrompt("Create a plan")
	if prompt == "Create a plan" {
		t.Error("Should add context to prompt")
	}
	t.Logf("Boosted prompt: %s", prompt)
}

func TestContextBoost_Clear(t *testing.T) {
	c := NewContextBoost()
	c.AddHotFile("agent.go")
	c.AddRecentCommand("go test")
	c.Clear()

	score := c.GetBoostScore("agent.go")
	if score != 0 {
		t.Error("Should have no boost after clear")
	}
}
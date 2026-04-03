package planning

import (
	"strings"
	"testing"
)

func TestPrometheus_CreatePlan(t *testing.T) {
	t.Parallel()

	p := NewPrometheus()
	plan := p.CreatePlan("build a web app")

	if plan.Task != "build a web app" {
		t.Errorf("expected 'build a web app', got %q", plan.Task)
	}
	if len(plan.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(plan.Steps))
	}
	if plan.Status != "pending" {
		t.Errorf("expected pending, got %q", plan.Status)
	}
}

func TestPrometheus_CreatePlan_Empty(t *testing.T) {
	t.Parallel()

	p := NewPrometheus()
	plan := p.CreatePlan("")
	if plan.Task != "default-task" {
		t.Errorf("expected default-task, got %q", plan.Task)
	}
}

func TestPrometheus_Interview(t *testing.T) {
	t.Parallel()

	p := NewPrometheus()
	q := p.Interview("some task")
	if !strings.Contains(q, "expected outcome") {
		t.Errorf("expected outcome question, got %q", q)
	}
}

func TestMetis_Analyze(t *testing.T) {
	t.Parallel()

	p := NewPrometheus()
	m := NewMetis()
	plan := p.CreatePlan("build something")
	risks := m.Analyze(plan)

	if len(risks) == 0 {
		t.Error("expected risks")
	}
	if !strings.Contains(risks[0], "No error handling") {
		t.Errorf("expected error handling risk, got %q", risks[0])
	}
}

func TestMetis_AnalyzeNilPlan(t *testing.T) {
	t.Parallel()

	m := NewMetis()
	risks := m.Analyze(nil)
	if len(risks) != 0 {
		t.Errorf("expected 0 risks for nil plan, got %d", len(risks))
	}
}

func TestMomus_Review_Approve(t *testing.T) {
	t.Parallel()

	p := NewPrometheus()
	m := NewMomus()
	plan := p.CreatePlan("build something")

	approved, feedback := m.Review(plan)
	if !approved {
		t.Errorf("expected approval, got feedback: %s", feedback)
	}
}

func TestMomus_Review_RejectNil(t *testing.T) {
	t.Parallel()

	m := NewMomus()
	approved, feedback := m.Review(nil)
	if approved {
		t.Error("expected rejection for nil plan")
	}
	if feedback == "" {
		t.Error("expected feedback")
	}
}

func TestMomus_Review_RejectFewSteps(t *testing.T) {
	t.Parallel()

	m := NewMomus()
	plan := &Plan{Task: "x", Steps: []Step{{ID: "s1", Description: "one step", Status: StatusPending}}}
	approved, feedback := m.Review(plan)
	if approved {
		t.Errorf("expected rejection, got %s", feedback)
	}
}

func TestMomus_Review_RejectEmptyDesc(t *testing.T) {
	t.Parallel()

	m := NewMomus()
	plan := &Plan{Task: "x", Steps: []Step{
		{ID: "s1", Description: "step one"},
		{ID: "s2", Description: "  "},
	}}
	approved, feedback := m.Review(plan)
	if approved {
		t.Errorf("expected rejection, got %s", feedback)
	}
}

func TestMomus_Review_RejectCircular(t *testing.T) {
	t.Parallel()

	m := NewMomus()
	plan := &Plan{Task: "x", Steps: []Step{
		{ID: "s1", Description: "step one", DependsOn: []string{"s2"}},
		{ID: "s2", Description: "step two", DependsOn: []string{"s1"}},
	}}
	approved, feedback := m.Review(plan)
	if approved {
		t.Errorf("expected rejection, got %s", feedback)
	}
}

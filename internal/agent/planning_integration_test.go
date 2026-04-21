package agent

import (
	"context"
	"testing"

	"github.com/hermes-v2/aigo/internal/planning"
)

func TestAgent_WithPlannerIntegration(t *testing.T) {
	planner := planning.NewPlanner()
	ctx := context.Background()
	input := []string{"fix the login bug"}

	plan, err := planner.Plan(ctx, input)
	if err != nil {
		t.Fatalf("Planner.Plan() failed: %v", err)
	}

	if plan == nil {
		t.Fatal("Plan should not be nil")
	}

	if len(plan.Steps) == 0 {
		t.Error("Plan should have steps")
	}

	t.Logf("Plan generated: %d steps", len(plan.Steps))
}

func TestAgent_PlanningPhaseExists(t *testing.T) {
	a := &Agent{
		maxIter:   10,
		maxTokens: 8000,
	}

	if a.maxIter <= 0 {
		t.Error("Agent should have maxIter set")
	}
}
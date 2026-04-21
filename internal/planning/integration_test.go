package planning

import (
	"context"
	"testing"
)

func TestPlanningIntegration(t *testing.T) {
	ctx := context.Background()

	planner := NewPlanner()
	metis := NewMetis()
	momus := NewMomus()

	input := []string{"implement login"}

	plan, err := planner.Plan(ctx, input)
	if err != nil {
		t.Fatalf("Planner.Plan() failed: %v", err)
	}

	gaps, _ := metis.AnalyzeGaps(ctx, plan)
	if len(gaps) > 0 {
		t.Logf("Found %d gaps", len(gaps))
	}

	review, _ := momus.ReviewPlan(ctx, plan)
	if review == nil {
		t.Fatal("Review is nil")
	}

	t.Logf("Plan: %d steps, Review score: %d", len(plan.Steps), review.Score)
}
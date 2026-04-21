package planning

import (
	"context"
	"testing"
)

type mockLLMProvider struct {
	response string
	err      error
}

func (m *mockLLMProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestPlanner_SimplePlan(t *testing.T) {
	p := NewPlanner()
	if p == nil {
		t.Fatal("NewPlanner() returned nil")
	}

	ctx := context.Background()
	input := []string{"fix the bug"}

	result, err := p.Plan(ctx, input)
	if err != nil {
		t.Fatalf("Plan() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Plan() returned nil result")
	}

	if len(result.Steps) == 0 {
		t.Log("Warning: Plan returned empty steps (acceptable for skeleton)")
	}
}

func TestPlanner_LLMBased(t *testing.T) {
	p := NewPlanner()

	mock := &mockLLMProvider{
		response: "Step 1: Analyze\nStep 2: Implement\nStep 3: Test",
	}
	p.SetLLMProvider(mock)

	input := []string{"implement login"}
	plan, err := p.Plan(context.Background(), input)

	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	if len(plan.Steps) == 0 {
		t.Error("Plan should have steps")
	}

	if plan.Confidence != 80 {
		t.Errorf("Expected confidence 80, got %d", plan.Confidence)
	}

	t.Logf("LLM Plan: %d steps, confidence: %d", len(plan.Steps), plan.Confidence)
}

func TestPlanner_FallbackToRuleBased(t *testing.T) {
	p := NewPlanner()

	input := []string{"fix the bug"}
	plan, err := p.Plan(context.Background(), input)

	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	if len(plan.Steps) == 0 {
		t.Error("Plan should have steps")
	}

	if plan.Confidence != 50 {
		t.Errorf("Expected confidence 50 for rule-based, got %d", plan.Confidence)
	}

	t.Logf("Rule-based Plan: %d steps", len(plan.Steps))
}

func TestPlanner_RuleBasedFix(t *testing.T) {
	p := NewPlanner()
	steps := p.ruleBasedPlan("fix the login bug")

	if len(steps) != 4 {
		t.Errorf("Expected 4 steps for 'fix', got %d", len(steps))
	}
}

func TestPlanner_RuleBasedCreate(t *testing.T) {
	p := NewPlanner()
	steps := p.ruleBasedPlan("create new API")

	if len(steps) != 3 {
		t.Errorf("Expected 3 steps for 'create', got %d", len(steps))
	}
}

func TestPlanner_RuleBasedRefactor(t *testing.T) {
	p := NewPlanner()
	steps := p.ruleBasedPlan("refactor the code")

	if len(steps) != 4 {
		t.Errorf("Expected 4 steps for 'refactor', got %d", len(steps))
	}
}
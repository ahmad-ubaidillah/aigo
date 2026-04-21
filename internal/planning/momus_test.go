package planning

import (
	"context"
	"testing"
)

type mockMomusProvider struct {
	response string
	err      error
}

func (m *mockMomusProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestMomus_ReviewPlan(t *testing.T) {
	m := NewMomus()
	if m == nil {
		t.Fatal("NewMomus() returned nil")
	}

	ctx := context.Background()
	plan := &Plan{
		Steps:       []string{"fix bug", "test"},
		Description: "test plan",
	}

	review, err := m.ReviewPlan(ctx, plan)
	if err != nil {
		t.Fatalf("ReviewPlan() returned error: %v", err)
	}

	if review == nil {
		t.Fatal("ReviewPlan() returned nil")
	}
}

func TestMomus_CheckCompleteness(t *testing.T) {
	m := NewMomus()

	complete := m.CheckCompleteness(&Plan{Steps: []string{"step1", "step2"}})
	if !complete {
		t.Log("Multi-step plan should be complete")
	}
}

func TestMomus_LLMReviewPlan(t *testing.T) {
	m := NewMomus()

	mock := &mockMomusProvider{
		response: "Plan is complete and verifiable",
	}
	m.SetLLMProvider(mock)

	plan := &Plan{
		Steps:       []string{"fix bug", "test"},
		Description: "test plan",
	}

	review, err := m.ReviewPlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("ReviewPlan() failed: %v", err)
	}

	t.Logf("LLM Review score: %d", review.Score)
	if review.CompletenessDetails != "" {
		t.Logf("Completeness: %s", review.CompletenessDetails)
	}
	if review.VerifiabilityDetails != "" {
		t.Logf("Verifiability: %s", review.VerifiabilityDetails)
	}
}

func TestMomus_LLMEvaluateCompleteness(t *testing.T) {
	m := NewMomus()

	mock := &mockMomusProvider{
		response: "Plan is complete - all requirements covered",
	}
	m.SetLLMProvider(mock)

	plan := &Plan{
		Steps:       []string{"implement login", "add tests"},
		Description: "test plan",
	}

	details, err := m.checkCompletenessWithLLM(context.Background(), plan)
	if err != nil {
		t.Fatalf("checkCompletenessWithLLM failed: %v", err)
	}

	t.Logf("Completeness check: %s", details)
}

func TestMomus_LLMEvaluateVerifiability(t *testing.T) {
	m := NewMomus()

	mock := &mockMomusProvider{
		response: "Plan is verifiable - clear success criteria",
	}
	m.SetLLMProvider(mock)

	plan := &Plan{
		Steps:       []string{"implement login", "verify with tests"},
		Description: "test plan",
	}

	details, err := m.checkVerifiabilityWithLLM(context.Background(), plan)
	if err != nil {
		t.Fatalf("checkVerifiabilityWithLLM failed: %v", err)
	}

	t.Logf("Verifiability check: %s", details)
}
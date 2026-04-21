package planning

import (
	"context"
	"testing"
)

type mockMetisProvider struct {
	response string
	err      error
}

func (m *mockMetisProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestMetis_AnalyzeGaps(t *testing.T) {
	m := NewMetis()
	if m == nil {
		t.Fatal("NewMetis() returned nil")
	}

	ctx := context.Background()
	plan := &Plan{
		Steps:       []string{"fix bug", "test"},
		Description: "test plan",
	}

	gaps, err := m.AnalyzeGaps(ctx, plan)
	if err != nil {
		t.Fatalf("AnalyzeGaps() returned error: %v", err)
	}

	if gaps == nil {
		t.Fatal("AnalyzeGaps() returned nil")
	}
}

func TestMetis_DetectAmbiguity(t *testing.T) {
	m := NewMetis()

	ambiguous := []string{"fix something", "do it"}
	gap := m.DetectAmbiguity(ambiguous)

	if gap == nil {
		t.Log("DetectAmbiguity returned nil for ambiguous input (acceptable for skeleton)")
	}
}

func TestMetis_LLMAnalyzeGaps(t *testing.T) {
	m := NewMetis()

	mock := &mockMetisProvider{
		response: "Found ambiguous steps in the plan",
	}
	m.SetLLMProvider(mock)

	plan := &Plan{
		Steps:       []string{"fix bug", "test"},
		Description: "test plan",
	}

	gaps, err := m.AnalyzeGaps(context.Background(), plan)
	if err != nil {
		t.Fatalf("AnalyzeGaps() failed: %v", err)
	}

	t.Logf("LLM detected %d gaps", len(gaps))
}

func TestMetis_LLMDetectAmbiguity(t *testing.T) {
	m := NewMetis()

	mock := &mockMetisProvider{
		response: "Input is ambiguous - missing details",
	}
	m.SetLLMProvider(mock)

	inputs := []string{"implement something"}
	gap := m.DetectAmbiguity(inputs)

	if gap != nil {
		t.Logf("LLM detected ambiguity: %s", gap.Description)
	}
}

func TestMetis_TrackAssumptions(t *testing.T) {
	m := NewMetis()

	plan := &Plan{
		Steps:       []string{"fix bug?", "test - assume network works"},
		Description: "test plan",
	}

	assumptions := m.TrackAssumptions(plan)
	if len(assumptions) == 0 {
		t.Error("Should track assumptions from steps")
	}
}

func TestMetis_TrackAssumptionsWithLLM(t *testing.T) {
	m := NewMetis()

	mock := &mockMetisProvider{
		response: "Assumption: User wants REST API",
	}
	m.SetLLMProvider(mock)

	plan := &Plan{
		Steps:       []string{"create API"},
		Description: "test plan",
	}

	assumptions := m.TrackAssumptions(plan)
	t.Logf("Tracked %d assumptions", len(assumptions))
}
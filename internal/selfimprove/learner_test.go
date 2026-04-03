package selfimprove

import (
	"context"
	"testing"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type mockStore struct {
	logs   []*types.SelfImproveLog
	stats  LearnerStats
}

func (m *mockStore) SaveLog(log *types.SelfImproveLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockStore) ListLogs(sessionID string, limit int) ([]types.SelfImproveLog, error) {
	result := make([]types.SelfImproveLog, 0, len(m.logs))
	for _, l := range m.logs {
		if sessionID == "" || l.SessionID == sessionID {
			result = append(result, *l)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockStore) GetStats() (LearnerStats, error) {
	return m.stats, nil
}

func TestNewLearner(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	if l == nil {
		t.Error("expected learner")
	}
}

func TestLearner_LogTurn(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	err := l.LogTurn(context.Background(), "s1", "input", "output", "success", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(store.logs))
	}
}

func TestLearner_GetRecentTurns(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	l.LogTurn(context.Background(), "s1", "input1", "output1", "success", false)
	l.LogTurn(context.Background(), "s1", "input2", "output2", "failure", false)
	logs, err := l.GetRecentTurns(context.Background(), "s1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}
}

func TestLearner_AnalyzeForSkillCreation(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	logs := []types.SelfImproveLog{
		{TurnInput: "run test", TurnOutput: "error", Outcome: "failure"},
		{TurnInput: "run test", TurnOutput: "error2", Outcome: "failure"},
	}
	proposals, err := l.AnalyzeForSkillCreation(context.Background(), logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(proposals))
	}
}

func TestLearner_AnalyzeForSkillCreation_NoPatterns(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	logs := []types.SelfImproveLog{
		{TurnInput: "run test", TurnOutput: "ok", Outcome: "success"},
	}
	proposals, err := l.AnalyzeForSkillCreation(context.Background(), logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 0 {
		t.Errorf("expected 0 proposals, got %d", len(proposals))
	}
}

func TestLearner_GetStats(t *testing.T) {
	t.Parallel()
	store := &mockStore{stats: LearnerStats{TotalTurns: 10}}
	l := NewLearner(store, nil)
	stats, err := l.GetStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalTurns != 10 {
		t.Errorf("expected 10, got %d", stats.TotalTurns)
	}
}

func TestLearner_SuggestSkills(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	l.LogTurn(context.Background(), "s1", "run test", "error", "failure", false)
	l.LogTurn(context.Background(), "s1", "run test", "error2", "failure", false)
	proposals, err := l.SuggestSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(proposals))
	}
}

func TestNewGenerator(t *testing.T) {
	t.Parallel()
	g := NewGenerator()
	if g == nil {
		t.Error("expected generator")
	}
	if len(g.proposals) != 0 {
		t.Errorf("expected 0 proposals, got %d", len(g.proposals))
	}
}

func TestGenerator_AddProposal(t *testing.T) {
	t.Parallel()
	g := NewGenerator()
	g.AddProposal(SkillProposal{Trigger: "test", Suggestion: "test skill"})
	if len(g.proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(g.proposals))
	}
}

func TestGenerator_GenerateSkill(t *testing.T) {
	t.Parallel()
	g := NewGenerator()
	g.AddProposal(SkillProposal{Trigger: "test", Suggestion: "test skill", SampleError: "error"})
	skill, err := g.GenerateSkill()
	if err != nil {
		t.Fatal(err)
	}
	if skill.Name == "" {
		t.Error("expected skill name")
	}
	if skill.Code == "" {
		t.Error("expected skill code")
	}
}

func TestGenerator_GenerateSkillEmpty(t *testing.T) {
	t.Parallel()
	g := NewGenerator()
	_, err := g.GenerateSkill()
	if err == nil {
		t.Error("expected error for empty proposals")
	}
}

func TestGenerator_GetProposals(t *testing.T) {
	t.Parallel()
	g := NewGenerator()
	g.AddProposal(SkillProposal{Trigger: "test"})
	proposals := g.GetProposals()
	if len(proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(proposals))
	}
}

func TestSanitizeName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"run test", "auto_run_test"},
		{"test/path", "auto_test_path"},
		{"test-name", "auto_test_name"},
		{"VERY_LONG_NAME_THAT_EXCEEDS_THIRTY_CHARACTERS", "auto_very_long_name_that_exceeds_th"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestGenerateLogID(t *testing.T) {
	t.Parallel()
	id := generateLogID()
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGenerateSkillID(t *testing.T) {
	t.Parallel()
	id := generateSkillID()
	if id == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGenerateSkillCode(t *testing.T) {
	t.Parallel()
	code := generateSkillCode("test trigger", "sample error")
	if code == "" {
		t.Error("expected non-empty code")
	}
}

func TestMin(t *testing.T) {
	t.Parallel()
	if min(3, 5) != 3 {
		t.Errorf("expected 3, got %d", min(3, 5))
	}
	if min(5, 3) != 3 {
		t.Errorf("expected 3, got %d", min(5, 3))
	}
}

func TestExtractTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"run this test now", "run this test now"},
		{"a b c d e f g", "a b c d e"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractTrigger(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestSkillProposal(t *testing.T) {
	t.Parallel()
	p := SkillProposal{
		Trigger:     "test",
		Suggestion:  "suggestion",
		SampleError: "error",
		Confidence:  0.5,
	}
	if p.Trigger != "test" {
		t.Errorf("expected test, got %s", p.Trigger)
	}
	if p.Confidence != 0.5 {
		t.Errorf("expected 0.5, got %f", p.Confidence)
	}
}

func TestLearnerStats(t *testing.T) {
	t.Parallel()
	s := LearnerStats{
		TotalTurns:     10,
		SuccessRate:    0.8,
		SkillsCreated:  2,
		FailurePattern: 3,
	}
	if s.TotalTurns != 10 {
		t.Errorf("expected 10, got %d", s.TotalTurns)
	}
	if s.SuccessRate != 0.8 {
		t.Errorf("expected 0.8, got %f", s.SuccessRate)
	}
}

func TestLearner_LogTurnWithSkillGen(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	err := l.LogTurn(context.Background(), "s1", "input", "output", "success", true)
	if err != nil {
		t.Fatal(err)
	}
	if !store.logs[0].SkillGen {
		t.Error("expected SkillGen to be true")
	}
}

func TestLearner_GetRecentTurnsFiltered(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	l.LogTurn(context.Background(), "s1", "input1", "output1", "success", false)
	l.LogTurn(context.Background(), "s2", "input2", "output2", "success", false)
	logs, err := l.GetRecentTurns(context.Background(), "s1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestLearner_GetRecentTurnsLimited(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	l.LogTurn(context.Background(), "s1", "input1", "output1", "success", false)
	l.LogTurn(context.Background(), "s1", "input2", "output2", "success", false)
	logs, err := l.GetRecentTurns(context.Background(), "s1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestLearner_AnalyzeForSkillCreation_PartialOutcome(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	logs := []types.SelfImproveLog{
		{TurnInput: "partial test", TurnOutput: "partial", Outcome: "partial"},
		{TurnInput: "partial test", TurnOutput: "partial2", Outcome: "partial"},
	}
	proposals, err := l.AnalyzeForSkillCreation(context.Background(), logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(proposals))
	}
}

func TestLearner_AnalyzeForSkillCreation_MixedOutcomes(t *testing.T) {
	t.Parallel()
	store := &mockStore{}
	l := NewLearner(store, nil)
	logs := []types.SelfImproveLog{
		{TurnInput: "mixed test", TurnOutput: "error", Outcome: "failure"},
		{TurnInput: "mixed test", TurnOutput: "partial", Outcome: "partial"},
		{TurnInput: "mixed test", TurnOutput: "ok", Outcome: "success"},
	}
	proposals, err := l.AnalyzeForSkillCreation(context.Background(), logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 1 {
		t.Errorf("expected 1 proposal, got %d", len(proposals))
	}
}

func TestGenerator_GenerateSkillRemovesProposal(t *testing.T) {
	t.Parallel()
	g := NewGenerator()
	g.AddProposal(SkillProposal{Trigger: "test"})
	g.GenerateSkill()
	if len(g.proposals) != 0 {
		t.Errorf("expected 0 proposals after generation, got %d", len(g.proposals))
	}
}

func TestGenerateSkillCode_ContainsTrigger(t *testing.T) {
	t.Parallel()
	code := generateSkillCode("my trigger", "my error")
	if !containsStr(code, "my trigger") {
		t.Error("expected code to contain trigger")
	}
	if !containsStr(code, "my error") {
		t.Error("expected code to contain error")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSelfImproveLog(t *testing.T) {
	t.Parallel()
	log := types.SelfImproveLog{
		ID:         "log1",
		SessionID:  "s1",
		TurnInput:  "input",
		TurnOutput: "output",
		Outcome:    "success",
		SkillGen:   false,
		CreatedAt:  time.Now(),
	}
	if log.ID != "log1" {
		t.Errorf("expected log1, got %s", log.ID)
	}
}

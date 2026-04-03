package planning

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
)

// MomusAgent reviews plans with LLM-powered constitution checks.
type MomusAgent struct {
	client llm.LLMClient
	model  string
}

// ReviewResult contains the outcome of a plan review.
type ReviewResult struct {
	Approved    bool
	Score       int
	Violations  []Violation
	Suggestions []string
}

// Violation represents a constitution violation.
type Violation struct {
	Criterion   string
	Severity    string
	StepID      string
	Description string
	Suggestion  string
}

// NewMomusAgent creates a Momus agent.
func NewMomusAgent(client llm.LLMClient, model string) *MomusAgent {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &MomusAgent{client: client, model: model}
}

// ReviewWithLLM reviews a plan using LLM with constitution criteria.
func (m *MomusAgent) ReviewWithLLM(ctx context.Context, plan *Plan) (*ReviewResult, error) {
	if m.client == nil {
		return m.reviewFallback(plan), nil
	}

	prompt := fmt.Sprintf("Review this plan against these criteria:\n1. LibraryFirst - starts as standalone library?\n2. TestFirst - tests before implementation?\n3. Simplicity - no over-engineering?\n4. Clarity - steps unambiguous?\n5. Feasibility - achievable with available tools?\n6. Safety - no destructive operations?\n\nTask: %s\nSteps: %d\n", plan.Task, len(plan.Steps))
	for _, s := range plan.Steps {
		prompt += fmt.Sprintf("- %s\n", s.Description)
	}

	messages := []llm.Message{
		{Role: "system", Content: "You are a ruthless plan reviewer. Find every flaw."},
		{Role: "user", Content: prompt},
	}

	resp, err := m.client.Chat(ctx, messages)
	if err != nil {
		return m.reviewFallback(plan), nil
	}

	return m.parseReviewResponse(plan, resp.Content), nil
}

func (m *MomusAgent) reviewFallback(plan *Plan) *ReviewResult {
	result := &ReviewResult{Approved: true, Score: 100}

	if len(plan.Steps) < 2 {
		result.Approved = false
		result.Score = 20
		result.Violations = append(result.Violations, Violation{
			Criterion: "Clarity", Severity: "critical",
			Description: "Plan has fewer than 2 steps",
		})
	}

	hasTest := false
	for _, s := range plan.Steps {
		if strings.Contains(strings.ToLower(s.Description), "test") {
			hasTest = true
		}
	}
	if !hasTest {
		result.Score -= 15
		result.Violations = append(result.Violations, Violation{
			Criterion: "TestFirst", Severity: "major",
			Description: "No test steps found",
		})
	}

	if result.Score < 50 {
		result.Approved = false
	}
	return result
}

func (m *MomusAgent) parseReviewResponse(plan *Plan, content string) *ReviewResult {
	result := m.reviewFallback(plan)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result.Suggestions = append(result.Suggestions, line)
		}
	}
	result.Score = 75
	return result
}

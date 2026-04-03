package planning

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
)

// MetisAgent analyzes plans for gaps and risks using LLM.
type MetisAgent struct {
	client llm.LLMClient
	model  string
}

// NewMetisAgent creates a Metis agent.
func NewMetisAgent(client llm.LLMClient, model string) *MetisAgent {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &MetisAgent{client: client, model: model}
}

// GapReport contains detected gaps in a plan.
type GapReport struct {
	Gaps        []Gap
	Suggestions []string
	Confidence  float64
}

// Gap represents a detected gap in a plan.
type Gap struct {
	Type        string
	StepID      string
	Description string
	Severity    string
	Suggestion  string
}

// AnalyzeWithLLM analyzes a plan using LLM for gap detection.
func (m *MetisAgent) AnalyzeWithLLM(ctx context.Context, plan *Plan) (*GapReport, error) {
	if m.client == nil {
		return m.analyzeFallback(plan), nil
	}

	prompt := fmt.Sprintf("Review this plan for gaps, risks, and missing information:\nTask: %s\nSteps: %d\n", plan.Task, len(plan.Steps))
	for _, s := range plan.Steps {
		prompt += fmt.Sprintf("- %s (deps: %v)\n", s.Description, s.DependsOn)
	}

	messages := []llm.Message{
		{Role: "system", Content: "You are a gap analysis expert. Identify missing information, risks, and vague steps."},
		{Role: "user", Content: prompt},
	}

	resp, err := m.client.Chat(ctx, messages)
	if err != nil {
		return m.analyzeFallback(plan), nil
	}

	return m.parseGapResponse(plan, resp.Content), nil
}

// EnhancePlan adds missing steps and clarifies vague steps.
func (m *MetisAgent) EnhancePlan(plan *Plan, gaps []Gap) *Plan {
	enhanced := *plan
	enhanced.Steps = make([]Step, len(plan.Steps))
	copy(enhanced.Steps, plan.Steps)

	for _, gap := range gaps {
		if gap.Type == "MissingTests" {
			enhanced.AddStep(Step{
				ID:          fmt.Sprintf("test-%s", gap.StepID),
				Description: "Add tests for: " + gap.Description,
				Status:      StatusPending,
				DependsOn:   []string{gap.StepID},
			})
		}
		if gap.Type == "StepTooVague" {
			for i, s := range enhanced.Steps {
				if s.ID == gap.StepID {
					enhanced.Steps[i].Description = s.Description + " — " + gap.Suggestion
				}
			}
		}
	}
	return &enhanced
}

func (m *MetisAgent) analyzeFallback(plan *Plan) *GapReport {
	report := &GapReport{Confidence: 0.5}
	for _, s := range plan.Steps {
		if len(s.DependsOn) == 0 && s.ID != "step-1" {
			report.Gaps = append(report.Gaps, Gap{
				Type: "MissingDependencies", StepID: s.ID,
				Description: s.Description, Severity: "warning",
				Suggestion: "Declare explicit dependencies",
			})
		}
		if !strings.Contains(strings.ToLower(s.Description), "test") {
			report.Gaps = append(report.Gaps, Gap{
				Type: "MissingTests", StepID: s.ID,
				Description: s.Description, Severity: "info",
				Suggestion: "Add test step",
			})
		}
	}
	return report
}

func (m *MetisAgent) parseGapResponse(plan *Plan, content string) *GapReport {
	report := m.analyzeFallback(plan)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			report.Suggestions = append(report.Suggestions, line)
		}
	}
	report.Confidence = 0.8
	return report
}

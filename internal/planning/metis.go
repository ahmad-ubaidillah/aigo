package planning

import (
	"context"
	"strings"
)

type Metis struct {
	strictness int
	llmEnabled bool
	provider  LLMProvider
}

type Gap struct {
	Type         string
	Description string
	Severity    int
	Assumptions []string
}

func NewMetis() *Metis {
	return &Metis{
		strictness: 5,
	}
}

func (m *Metis) SetLLMProvider(provider LLMProvider) {
	m.provider = provider
	m.llmEnabled = true
}

func (m *Metis) AnalyzeGaps(ctx context.Context, plan *Plan) ([]*Gap, error) {
	gaps := make([]*Gap, 0)

	if plan == nil {
		gaps = append(gaps, &Gap{
			Type:        "missing_plan",
			Description: "Plan is nil",
			Severity:    3,
		})
		return gaps, nil
	}

	if len(plan.Steps) == 0 {
		gaps = append(gaps, &Gap{
			Type:        "empty_steps",
			Description: "Plan has no steps",
			Severity:    2,
		})
	}

	if m.llmEnabled && m.provider != nil {
		return m.analyzeGapsWithLLM(ctx, plan, gaps)
	}

	return gaps, nil
}

func (m *Metis) analyzeGapsWithLLM(ctx context.Context, plan *Plan, gaps []*Gap) ([]*Gap, error) {
	prompt := "Analyze this plan for gaps, ambiguities, and missing information.\n\nPlan: " + strings.Join(plan.Steps, ", ")

	messages := []Message{
		{Role: "system", Content: "You are a gap analysis assistant. Identify missing requirements, unclear steps, and potential issues."},
		{Role: "user", Content: prompt},
	}

	response, err := m.provider.Chat(ctx, messages)
	if err != nil {
		return gaps, nil
	}

	if strings.Contains(strings.ToLower(response), "ambiguous") || strings.Contains(strings.ToLower(response), "unclear") {
		gaps = append(gaps, &Gap{
			Type:         "llm_detected",
			Description:  response,
			Severity:     3,
			Assumptions: []string{"LLM detected issues"},
		})
	}

	return gaps, nil
}

func (m *Metis) DetectAmbiguity(inputs []string) *Gap {
	ambiguousPhrases := []string{"something", "thing", "do it"}

	for _, input := range inputs {
		for _, phrase := range ambiguousPhrases {
			if len(input) < 10 && phrase == input {
				return &Gap{
					Type:        "ambiguous",
					Description: "Input too short: " + input,
					Severity:    4,
				}
			}
		}
	}

	if m.llmEnabled && m.provider != nil {
		return m.detectAmbiguityWithLLM(inputs)
	}

	return nil
}

func (m *Metis) detectAmbiguityWithLLM(inputs []string) *Gap {
	prompt := "Detect ambiguity in these inputs: " + strings.Join(inputs, ", ")

	messages := []Message{
		{Role: "system", Content: "You are an ambiguity detection assistant."},
		{Role: "user", Content: prompt},
	}

	response, err := m.provider.Chat(context.Background(), messages)
	if err != nil {
		return nil
	}

	if strings.Contains(strings.ToLower(response), "ambiguous") {
		return &Gap{
			Type:         "llm_ambiguous",
			Description:  response,
			Severity:     3,
			Assumptions:   inputs,
		}
	}

	return nil
}

func (m *Metis) TrackAssumptions(plan *Plan) []string {
	assumptions := make([]string, 0)

	for _, step := range plan.Steps {
		if strings.Contains(step, "?") || strings.Contains(step, "assume") {
			assumptions = append(assumptions, step)
		}
	}

	if m.llmEnabled && m.provider != nil {
		llmAssumptions := m.extractAssumptionsWithLLM(plan)
		assumptions = append(assumptions, llmAssumptions...)
	}

	return assumptions
}

func (m *Metis) extractAssumptionsWithLLM(plan *Plan) []string {
	prompt := "Extract implicit assumptions from this plan: " + strings.Join(plan.Steps, ", ")

	messages := []Message{
		{Role: "system", Content: "You are an assumption extraction assistant."},
		{Role: "user", Content: prompt},
	}

	response, err := m.provider.Chat(context.Background(), messages)
	if err != nil {
		return []string{}
	}

	return []string{response}
}
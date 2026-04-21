package planning

import (
	"context"
	"strings"
)

type Momus struct {
	threshold int
	llmEnabled bool
	provider  LLMProvider
}

type Review struct {
	Complete    bool
	Verifiable  bool
	Score      int
	Notes      []string
	CompletenessDetails string
	VerifiabilityDetails string
}

func NewMomus() *Momus {
	return &Momus{
		threshold: 3,
	}
}

func (m *Momus) SetLLMProvider(provider LLMProvider) {
	m.provider = provider
	m.llmEnabled = true
}

func (m *Momus) ReviewPlan(ctx context.Context, plan *Plan) (*Review, error) {
	review := &Review{
		Complete:   m.CheckCompleteness(plan),
		Verifiable: m.CheckVerifiability(plan),
		Score:     0,
		Notes:     make([]string, 0),
	}

	if plan == nil {
		review.Notes = append(review.Notes, "plan is nil")
		return review, nil
	}

	if review.Complete {
		review.Score += 5
	}
	if review.Verifiable {
		review.Score += 3
	}

	if m.llmEnabled && m.provider != nil {
		return m.reviewPlanWithLLM(ctx, plan, review)
	}

	return review, nil
}

func (m *Momus) reviewPlanWithLLM(ctx context.Context, plan *Plan, review *Review) (*Review, error) {
	completenessDetails, err := m.checkCompletenessWithLLM(ctx, plan)
	if err == nil && completenessDetails != "" {
		review.CompletenessDetails = completenessDetails
		if strings.Contains(strings.ToLower(completenessDetails), "complete") {
			review.Score += 5
		}
	}

	verifiabilityDetails, err := m.checkVerifiabilityWithLLM(ctx, plan)
	if err == nil && verifiabilityDetails != "" {
		review.VerifiabilityDetails = verifiabilityDetails
		if strings.Contains(strings.ToLower(verifiabilityDetails), "verifiable") {
			review.Score += 3
		}
	}

	return review, nil
}

func (m *Momus) checkCompletenessWithLLM(ctx context.Context, plan *Plan) (string, error) {
	prompt := "Check if this plan is complete. Consider: Are all requirements covered? Are there missing steps?\n\nPlan: " + strings.Join(plan.Steps, ", ")

	messages := []Message{
		{Role: "system", Content: "You are a plan completeness reviewer."},
		{Role: "user", Content: prompt},
	}

	return m.provider.Chat(ctx, messages)
}

func (m *Momus) checkVerifiabilityWithLLM(ctx context.Context, plan *Plan) (string, error) {
	prompt := "Check if this plan is verifiable. Consider: Can each step be verified? Are success criteria clear?\n\nPlan: " + strings.Join(plan.Steps, ", ")

	messages := []Message{
		{Role: "system", Content: "You are a plan verifiability reviewer."},
		{Role: "user", Content: prompt},
	}

	return m.provider.Chat(ctx, messages)
}

func (m *Momus) CheckCompleteness(plan *Plan) bool {
	if plan == nil {
		return false
	}
	return len(plan.Steps) >= 1
}

func (m *Momus) CheckVerifiability(plan *Plan) bool {
	if plan == nil {
		return false
	}
	for _, step := range plan.Steps {
		if len(step) > 0 {
			return true
		}
	}
	return false
}
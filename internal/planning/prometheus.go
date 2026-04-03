package planning

import (
	"context"
	"fmt"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
)

// PrometheusAgent generates strategic plans using LLM.
type PrometheusAgent struct {
	client      llm.LLMClient
	model       string
	temperature float64
}

// NewPrometheusAgent creates a Prometheus agent with an LLM client.
func NewPrometheusAgent(client llm.LLMClient, model string) *PrometheusAgent {
	if model == "" {
		model = "gpt-4"
	}
	return &PrometheusAgent{
		client:      client,
		model:       model,
		temperature: 0.3,
	}
}

// GeneratePlanWithLLM generates a plan using LLM.
func (p *PrometheusAgent) GeneratePlanWithLLM(ctx context.Context, task string) (*Plan, error) {
	if p.client == nil {
		return NewPlan(task), nil
	}

	prompt := fmt.Sprintf(`Break down this task into 3-5 concrete, actionable steps.
Return a JSON object with:
- "steps": array of {id, description, tool, depends_on}
- "risks": array of potential issues
- "alternatives": array of fallback approaches
- "estimated_tokens": integer

Task: %s`, task)

	messages := []llm.Message{
		{Role: "system", Content: "You are a strategic planner. Break tasks into clear, sequential steps with dependencies."},
		{Role: "user", Content: prompt},
	}

	resp, err := p.client.Chat(ctx, messages)
	if err != nil {
		return NewPlan(task), nil
	}

	plan := p.parseLLMResponse(task, resp.Content)
	return plan, nil
}

// InterviewWithLLM generates clarifying questions using LLM.
func (p *PrometheusAgent) InterviewWithLLM(ctx context.Context, task string) ([]string, error) {
	if p.client == nil {
		return []string{"What is the expected outcome? Any constraints? Deadline?"}, nil
	}

	messages := []llm.Message{
		{Role: "system", Content: "Ask 3 clarifying questions to understand the task better."},
		{Role: "user", Content: fmt.Sprintf("Task: %s\nWhat questions do you need answered before planning?", task)},
	}

	resp, err := p.client.Chat(ctx, messages)
	if err != nil {
		return []string{"What is the expected outcome? Any constraints? Deadline?"}, nil
	}

	lines := strings.Split(resp.Content, "\n")
	var questions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "?") {
			questions = append(questions, strings.TrimPrefix(line, "- "))
		}
	}
	if len(questions) == 0 {
		questions = append(questions, resp.Content)
	}
	return questions, nil
}

func (p *PrometheusAgent) parseLLMResponse(task, content string) *Plan {
	plan := NewPlan(task)
	plan.Risks = []string{"No risk analysis performed"}
	plan.Alternatives = []string{"No alternatives identified"}

	steps := strings.Split(content, "\n")
	for i, step := range steps {
		step = strings.TrimSpace(step)
		if step == "" {
			continue
		}
		plan.AddStep(Step{
			ID:          fmt.Sprintf("step-%d", i+1),
			Description: step,
			Status:      StatusPending,
		})
	}

	if len(plan.Steps) == 0 {
		plan.AddStep(Step{ID: "step-1", Description: "Execute: " + task, Status: StatusPending})
	}
	return plan
}

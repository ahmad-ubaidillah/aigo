package planning

import (
	"context"
	"strings"
)

// LLMProvider defines interface for LLM calls
type LLMProvider interface {
	Chat(ctx context.Context, messages []Message) (string, error)
}

// Message represents a chat message
type Message struct {
	Role    string
	Content string
}

// Planner handles execution planning
type Planner struct {
	maxSteps    int
	llmEnabled bool
	provider  LLMProvider
}

// Plan represents an execution plan
type Plan struct {
	Steps       []string
	Description string
	Confidence int
	Metadata   map[string]string
}

// NewPlanner creates a new Planner
func NewPlanner() *Planner {
	return &Planner{
		maxSteps:    50,
		llmEnabled: false,
	}
}

// SetLLMProvider sets the LLM provider for intelligent planning
func (p *Planner) SetLLMProvider(provider LLMProvider) {
	p.provider = provider
	p.llmEnabled = true
}

// Plan creates an execution plan for the given inputs
func (p *Planner) Plan(ctx context.Context, inputs []string) (*Plan, error) {
	plan := &Plan{
		Steps:       make([]string, 0),
		Description: "Execution plan",
		Confidence: 50,
		Metadata:   make(map[string]string),
	}

	// Use LLM if available
	if p.llmEnabled && p.provider != nil {
		return p.planWithLLM(ctx, inputs)
	}

	// Fallback to rule-based planning
	for _, input := range inputs {
		steps := p.ruleBasedPlan(input)
		plan.Steps = append(plan.Steps, steps...)
	}

	return plan, nil
}

// planWithLLM generates a plan using LLM
func (p *Planner) planWithLLM(ctx context.Context, inputs []string) (*Plan, error) {
	plan := &Plan{
		Steps:       make([]string, 0),
		Description: "LLM-generated execution plan",
		Confidence:  80,
		Metadata:   make(map[string]string),
	}

	prompt := "Create a detailed execution plan for the following task. Break it down into specific steps.\n\nTask: " + strings.Join(inputs, ", ")

	messages := []Message{
		{Role: "system", Content: "You are a planning assistant. Create detailed execution plans."},
		{Role: "user", Content: prompt},
	}

	response, err := p.provider.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Parse LLM response into steps
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			plan.Steps = append(plan.Steps, line)
		}
	}

	plan.Metadata["llm_response"] = response
	return plan, nil
}

// ruleBasedPlan creates a plan using rules when LLM is unavailable
func (p *Planner) ruleBasedPlan(input string) []string {
	steps := make([]string, 0)
	inputLower := strings.ToLower(input)

	// Analyze intent and create appropriate steps
	if strings.Contains(inputLower, "fix") || strings.Contains(inputLower, "bug") {
		steps = append(steps, "analyze: identify root cause")
		steps = append(steps, "locate: find relevant code")
		steps = append(steps, "implement: apply fix")
		steps = append(steps, "verify: test the fix")
	} else if strings.Contains(inputLower, "create") || strings.Contains(inputLower, "implement") {
		steps = append(steps, "design: create specification")
		steps = append(steps, "implement: write code")
		steps = append(steps, "test: verify functionality")
	} else if strings.Contains(inputLower, "refactor") {
		steps = append(steps, "analyze: understand current code")
		steps = append(steps, "plan: design new structure")
		steps = append(steps, "execute: apply refactoring")
		steps = append(steps, "verify: ensure tests pass")
	} else if strings.Contains(inputLower, "test") {
		steps = append(steps, "identify: find code to test")
		steps = append(steps, "write: create test cases")
		steps = append(steps, "run: execute tests")
	} else {
		steps = append(steps, "analyze: "+input)
		steps = append(steps, "execute: complete task")
		steps = append(steps, "verify: confirm completion")
	}

	return steps
}
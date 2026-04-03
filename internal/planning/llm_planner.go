package planning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/llm"
)

// LLMPlanner uses an LLM to generate plans
type LLMPlanner struct {
	client   llm.LLMClient
	provider TemplateProvider
}

// NewLLMPlanner creates a new LLM-based planner
func NewLLMPlanner(client llm.LLMClient) *LLMPlanner {
	return &LLMPlanner{
		client:   client,
		provider: NewDefaultTemplateProvider(),
	}
}

// NewLLMPlannerWithTemplates creates a planner with custom templates
func NewLLMPlannerWithTemplates(client llm.LLMClient, provider TemplateProvider) *LLMPlanner {
	return &LLMPlanner{
		client:   client,
		provider: provider,
	}
}

// Plan generates a plan from a task description using the LLM
func (p *LLMPlanner) Plan(ctx context.Context, task string) (*Plan, error) {
	// Detect workflow type
	workflowType := p.detectWorkflowType(task)
	
	// Get the appropriate template
	template := p.provider.GetTemplate(workflowType)
	
	// Build the prompt
	systemPrompt := template.SystemPrompt
	userPrompt := fmt.Sprintf("Task: %s\n\nPlease create a detailed plan with specific steps.", task)

	// Call the LLM
	response, err := p.client.CompleteWithSystem(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse the response into a plan
	plan, err := p.parsePlanFromLLM(response.Content, task)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	return plan, nil
}

// RefinePlan modifies an existing plan based on feedback
func (p *LLMPlanner) RefinePlan(ctx context.Context, plan *Plan, feedback string) (*Plan, error) {
	// Serialize the current plan
	planJSON, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize plan: %w", err)
	}

	systemPrompt := `You are an expert at refining execution plans. Given an existing plan and feedback, modify the plan to address the feedback. Maintain the same JSON structure with id, task, steps array, dependencies map, and status.

Output only valid JSON, no markdown formatting.`

	userPrompt := fmt.Sprintf(`Current Plan:
%s

Feedback: %s

Please provide an updated plan that addresses this feedback.`, string(planJSON), feedback)

	response, err := p.client.CompleteWithSystem(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to refine plan: %w", err)
	}

	updatedPlan, err := p.parsePlanFromLLM(response.Content, plan.Task)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refined plan: %w", err)
	}

	updatedPlan.ID = plan.ID
	updatedPlan.CreatedAt = plan.CreatedAt
	updatedPlan.UpdatedAt = time.Now()

	return updatedPlan, nil
}

// GetNextStep returns the next step that can be executed
func (p *LLMPlanner) GetNextStep(plan *Plan) *Step {
	ready := plan.GetReadySteps()
	if len(ready) == 0 {
		return nil
	}
	return &ready[0]
}

// detectWorkflowType attempts to determine the type of workflow from the task
func (p *LLMPlanner) detectWorkflowType(task string) WorkflowType {
	taskLower := strings.ToLower(task)
	
	// Bug fix keywords
	if strings.Contains(taskLower, "bug") || 
	   strings.Contains(taskLower, "fix") ||
	   strings.Contains(taskLower, "error") ||
	   strings.Contains(taskLower, "crash") ||
	   strings.Contains(taskLower, "issue") ||
	   strings.Contains(taskLower, "problem") ||
	   strings.Contains(taskLower, "doesn't work") ||
	   strings.Contains(taskLower, "not working") {
		return WorkflowTypeBugFix
	}

	// Refactor keywords
	if strings.Contains(taskLower, "refactor") ||
	   strings.Contains(taskLower, "restructure") ||
	   strings.Contains(taskLower, "reorganize") ||
	   strings.Contains(taskLower, "clean up") ||
	   strings.Contains(taskLower, "cleanup") ||
	   strings.Contains(taskLower, "improve performance") ||
	   strings.Contains(taskLower, "optimize") {
		return WorkflowTypeRefactor
	}

	// Default to feature
	return WorkflowTypeFeature
}

// parsePlanFromLLM parses an LLM response into a Plan
func (p *LLMPlanner) parsePlanFromLLM(content, task string) (*Plan, error) {
	// Try to extract JSON from the content
	jsonContent := extractJSON(content)
	
	var llmPlan struct {
		Task         string                   `json:"task"`
		Steps        []llmStep                `json:"steps"`
		Dependencies map[string][]string      `json:"dependencies"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &llmPlan); err != nil {
		// If JSON parsing fails, create a basic plan from the content
		return p.createBasicPlan(content, task)
	}

	// Create plan from parsed content
	planID := generatePlanID()
	plan := NewPlanWithID(planID, task)
	
	// Use the task from LLM if available
	if llmPlan.Task != "" {
		plan.Task = llmPlan.Task
	}

	// Add steps
	for i, s := range llmPlan.Steps {
		step := Step{
			ID:          fmt.Sprintf("step-%d", i+1),
			Description: s.Description,
			Tool:        s.Tool,
			Parameters:  s.Parameters,
			Status:      StatusPending,
		}
		plan.AddStep(step)
	}

	// Add dependencies
	if llmPlan.Dependencies != nil {
		plan.Dependencies = llmPlan.Dependencies
	}

	return plan, nil
}

// llmStep represents a step as parsed from LLM output
type llmStep struct {
	Description string         `json:"description"`
	Tool        string         `json:"tool"`
	Parameters  map[string]any `json:"parameters"`
}

// extractJSON extracts JSON content from text that may contain markdown
func extractJSON(content string) string {
	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)
	
	// Check for markdown code block
	if strings.HasPrefix(content, "```") {
		// Find the end of the opening ```
		startIdx := strings.Index(content, "\n")
		if startIdx != -1 {
			content = content[startIdx+1:]
		}
		
		// Remove closing ```
		endIdx := strings.LastIndex(content, "```")
		if endIdx != -1 {
			content = content[:endIdx]
		}
	}
	
	return strings.TrimSpace(content)
}

// createBasicPlan creates a simple plan when JSON parsing fails
func (p *LLMPlanner) createBasicPlan(content, task string) (*Plan, error) {
	planID := generatePlanID()
	plan := NewPlanWithID(planID, task)

	// Split content into lines and create steps from numbered items
	lines := strings.Split(content, "\n")
	stepNum := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for numbered list items (1., 1), etc.)
		if len(line) > 2 {
			// Handle "1." or "1)" style numbering
			if (line[0] >= '1' && line[0] <= '9') && 
			   (line[1] == '.' || line[1] == ')' || line[1] == ' ') {
				// Extract description
				desc := strings.TrimSpace(line[2:])
				if desc != "" {
					step := Step{
						ID:          fmt.Sprintf("step-%d", stepNum),
						Description: desc,
						Tool:        "generic",
						Parameters:  make(map[string]any),
						Status:      StatusPending,
					}
					plan.AddStep(step)
					stepNum++
				}
			}
		}
	}

	// If no steps were created, add a single step with the full content
	if len(plan.Steps) == 0 {
		plan.AddStep(Step{
			ID:          "step-1",
			Description: content,
			Tool:        "generic",
			Parameters:  make(map[string]any),
			Status:      StatusPending,
		})
	}

	return plan, nil
}

// generatePlanID creates a unique plan ID
func generatePlanID() string {
	return fmt.Sprintf("plan-%d", time.Now().UnixNano())
}

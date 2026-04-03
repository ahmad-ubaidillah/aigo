// Package planning provides plan generation and management for AI workflows
package planning

import (
	"time"
)

// PlanStatus represents the status of a plan or step
type PlanStatus string

const (
	StatusPending    PlanStatus = "pending"
	StatusInProgress PlanStatus = "in_progress"
	StatusCompleted  PlanStatus = "completed"
	StatusFailed     PlanStatus = "failed"
)

// Plan represents a structured execution plan
type Plan struct {
	// ID is a unique identifier for the plan
	ID string `json:"id"`

	// Task is the original task description
	Task string `json:"task"`

	// Steps are the individual steps in the plan
	Steps []Step `json:"steps"`

	// Dependencies maps step IDs to their dependency step IDs
	Dependencies map[string][]string `json:"dependencies,omitempty"`

	// Status is the current status of the plan
	Status PlanStatus `json:"status"`

	// CreatedAt is when the plan was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the plan was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// Wisdom contains insights or notes about the plan
	Wisdom string `json:"wisdom,omitempty"`

	// EstimatedTokens is the estimated token budget for the plan
	EstimatedTokens int `json:"estimated_tokens,omitempty"`

	// Risks lists potential issues with the plan
	Risks []string `json:"risks,omitempty"`

	// Alternatives lists fallback approaches
	Alternatives []string `json:"alternatives,omitempty"`
}

// Step represents a single step in a plan
type Step struct {
	// ID is a unique identifier for the step
	ID string `json:"id"`

	// Description is a human-readable description of the step
	Description string `json:"description"`

	// Tool is the tool to use for this step (optional)
	Tool string `json:"tool,omitempty"`

	// Parameters are the parameters for the tool (optional)
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Status is the current status of the step
	Status PlanStatus `json:"status"`

	// Result is the result of executing the step (optional)
	Result string `json:"result,omitempty"`

	// Error is any error that occurred during execution
	Error string `json:"error,omitempty"`

	// DependsOn lists step IDs that must complete before this step
	DependsOn []string `json:"depends_on,omitempty"`

	// IsParallel indicates if this step can run in parallel with siblings
	IsParallel bool `json:"is_parallel,omitempty"`

	// Depth indicates the decomposition depth of this step
	Depth int `json:"depth,omitempty"`
}

// NewPlan creates a new plan with the given task
func NewPlan(task string) *Plan {
	now := time.Now()
	return &Plan{
		ID:           generateID(),
		Task:         task,
		Steps:        []Step{},
		Dependencies: make(map[string][]string),
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewPlanWithID creates a new plan with a specific ID
func NewPlanWithID(id, task string) *Plan {
	now := time.Now()
	return &Plan{
		ID:           id,
		Task:         task,
		Steps:        []Step{},
		Dependencies: make(map[string][]string),
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AddStep adds a step to the plan
func (p *Plan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
	p.UpdatedAt = time.Now()
}

// GetStep retrieves a step by ID
func (p *Plan) GetStep(id string) *Step {
	for i := range p.Steps {
		if p.Steps[i].ID == id {
			return &p.Steps[i]
		}
	}
	return nil
}

// UpdateStepStatus updates the status of a step
func (p *Plan) UpdateStepStatus(id string, status PlanStatus) {
	for i := range p.Steps {
		if p.Steps[i].ID == id {
			p.Steps[i].Status = status
			p.UpdatedAt = time.Now()
			return
		}
	}
}

// IsComplete returns true if all steps are completed
func (p *Plan) IsComplete() bool {
	for _, step := range p.Steps {
		if step.Status != StatusCompleted {
			return false
		}
	}
	return true
}

// HasFailed returns true if any step has failed
func (p *Plan) HasFailed() bool {
	for _, step := range p.Steps {
		if step.Status == StatusFailed {
			return true
		}
	}
	return false
}

// GetReadySteps returns steps that are ready to execute (dependencies satisfied)
func (p *Plan) GetReadySteps() []Step {
	var ready []Step
	for _, step := range p.Steps {
		if step.Status != StatusPending {
			continue
		}
		if p.areDependenciesMet(step) {
			ready = append(ready, step)
		}
	}
	return ready
}

// areDependenciesMet checks if all dependencies of a step are completed
func (p *Plan) areDependenciesMet(step Step) bool {
	for _, depID := range step.DependsOn {
		depStep := p.GetStep(depID)
		if depStep == nil || depStep.Status != StatusCompleted {
			return false
		}
	}
	return true
}

// NewStep creates a new step with the given description
func NewStep(description string) Step {
	return Step{
		ID:          generateID(),
		Description: description,
		Status:      StatusPending,
		DependsOn:   []string{},
	}
}

// generateID generates a unique identifier
func generateID() string {
	return time.Now().Format("20060102_150405.000000")
}

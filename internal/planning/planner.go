package planning

import (
	"context"
)

// Planner defines the interface for plan generation and management
type Planner interface {
	// Plan generates a plan from a task description
	Plan(ctx context.Context, task string) (*Plan, error)

	// RefinePlan modifies an existing plan based on feedback
	RefinePlan(ctx context.Context, plan *Plan, feedback string) (*Plan, error)

	// ValidatePlan checks if a plan is valid and executable
	ValidatePlan(plan *Plan) error
}

// PlannerOptions contains options for plan generation
type PlannerOptions struct {
	// MaxDepth is the maximum decomposition depth
	MaxDepth int

	// EnableParallel enables parallel step detection
	EnableParallel bool

	// ModelHint suggests which model to use
	ModelHint string
}

// DefaultPlannerOptions returns default planner options
func DefaultPlannerOptions() PlannerOptions {
	return PlannerOptions{
		MaxDepth:       3,
		EnableParallel: true,
		ModelHint:      "",
	}
}

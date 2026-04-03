package planning

import (
	"fmt"
	"strings"
	"time"
)

// Prometheus is a basic plan creator
type Prometheus struct{}

// NewPrometheus creates a new Prometheus instance
func NewPrometheus() *Prometheus { return &Prometheus{} }

// CreatePlan creates a simple plan from a task
func (p *Prometheus) CreatePlan(task string) *Plan {
	t := strings.TrimSpace(task)
	if t == "" {
		t = "default-task"
	}
	plan := &Plan{
		ID:        fmt.Sprintf("plan-%d", time.Now().UnixNano()),
		Task:      t,
		CreatedAt: time.Now(),
		Status:    StatusPending,
		Steps:     make([]Step, 0),
	}
	steps := []Step{
		{ID: "step-1", Description: "step-1: " + t, DependsOn: []string{}, Status: StatusPending},
		{ID: "step-2", Description: "step-2: verify results", DependsOn: []string{"step-1"}, Status: StatusPending},
		{ID: "step-3", Description: "step-3: finalize", DependsOn: []string{"step-2"}, Status: StatusPending},
	}
	plan.Steps = steps
	return plan
}

// Interview returns questions for the user
func (p *Prometheus) Interview(task string) string {
	return "What is the expected outcome? Any constraints? Deadline?"
}

// Metis analyzes plans for risks
type Metis struct{}

// NewMetis creates a new Metis instance
func NewMetis() *Metis { return &Metis{} }

// Analyze analyzes a plan for potential risks
func (m *Metis) Analyze(plan *Plan) []string {
	if plan == nil {
		return []string{}
	}
	var risks []string
	for _, s := range plan.Steps {
		risks = append(risks, fmt.Sprintf("No error handling in %s", s.ID))
		risks = append(risks, fmt.Sprintf("No rollback strategy in %s", s.ID))
	}
	return risks
}

// Momus reviews plans for validity
type Momus struct{}

// NewMomus creates a new Momus instance
func NewMomus() *Momus { return &Momus{} }

// Review reviews a plan for validity
func (m *Momus) Review(plan *Plan) (bool, string) {
	if plan == nil {
		return false, "Plan is nil"
	}
	if len(plan.Steps) < 2 {
		return false, "Plan must have at least 2 steps"
	}
	for _, s := range plan.Steps {
		if strings.TrimSpace(s.Description) == "" {
			return false, fmt.Sprintf("Step %s has empty description", s.ID)
		}
	}
	byID := make(map[string]Step)
	for _, s := range plan.Steps {
		byID[s.ID] = s
	}
	colors := make(map[string]int)
	var visit func(string) bool
	visit = func(id string) bool {
		if colors[id] == 1 {
			return true
		}
		if colors[id] == 2 {
			return false
		}
		colors[id] = 1
		if s, ok := byID[id]; ok {
			for _, dep := range s.DependsOn {
				if visit(dep) {
					return true
				}
			}
		}
		colors[id] = 2
		return false
	}
	for _, s := range plan.Steps {
		if visit(s.ID) {
			return false, "Circular dependency detected in plan steps"
		}
	}
	return true, "Approved"
}

package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Orchestrator manages multiple agents and coordinates task execution.
type Orchestrator struct {
	pool     *AgentPool
	bus      *MessageBus
	resolver *ConflictResolver
	strategy ExecutionStrategy
	mu       sync.RWMutex
	tasks    map[string]*TaskStatus
	results  chan ExecutionResult
}

// TaskStatus tracks the status of a task being orchestrated.
type TaskStatus struct {
	ID          string
	Task        string
	Status      string // pending, running, completed, failed
	AssignedTo  string // Agent ID
	StartedAt   time.Time
	CompletedAt time.Time
	Error       string
}

// OrchestratorConfig holds configuration for the orchestrator.
type OrchestratorConfig struct {
	MaxAgents         int
	Strategy          ExecutionStrategy
	ConflictStrategy  ResolutionStrategy
	MessageBufferSize int
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(ctx context.Context, config OrchestratorConfig, factory AgentFactory) *Orchestrator {
	if config.Strategy == nil {
		config.Strategy = NewSequentialStrategy()
	}
	if config.ConflictStrategy == 0 {
		config.ConflictStrategy = ResolutionFIFO
	}
	if config.MessageBufferSize == 0 {
		config.MessageBufferSize = 1000
	}

	return &Orchestrator{
		pool:     NewAgentPool(ctx, config.MaxAgents, factory),
		bus:      NewMessageBus(config.MessageBufferSize),
		resolver: NewConflictResolver(config.ConflictStrategy),
		strategy: config.Strategy,
		tasks:    make(map[string]*TaskStatus),
		results:  make(chan ExecutionResult, 100),
	}
}

// ExecutePlan executes a plan with steps.
func (o *Orchestrator) ExecutePlan(ctx context.Context, steps []ExecutionTask) ([]ExecutionResult, error) {
	// Check for resource conflicts
	for _, step := range steps {
		if err := o.checkResources(step); err != nil {
			return nil, fmt.Errorf("resource conflict for task %s: %w", step.ID, err)
		}
	}

	// Execute using the configured strategy
	results, err := o.strategy.Execute(ctx, steps)
	if err != nil {
		return nil, err
	}

	// Process results
	for _, result := range results {
		o.results <- result
	}

	return results, nil
}

// ExecuteParallel executes independent tasks in parallel.
func (o *Orchestrator) ExecuteParallel(ctx context.Context, tasks []ExecutionTask) ([]ExecutionResult, error) {
	parallel := NewParallelStrategy(4)
	return parallel.Execute(ctx, tasks)
}

// ExecuteWithRetry executes a task with retry logic.
func (o *Orchestrator) ExecuteWithRetry(ctx context.Context, task ExecutionTask, maxRetries int) (*ExecutionResult, error) {
	retry := NewRetryStrategy(o.strategy, WithMaxRetries(maxRetries))
	results, err := retry.Execute(ctx, []ExecutionTask{task})
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return &results[0], nil
	}
	return nil, fmt.Errorf("no result returned")
}

// checkResources checks for resource conflicts and reserves them.
func (o *Orchestrator) checkResources(task ExecutionTask) error {
	for _, dep := range task.DependsOn {
		if !o.resolver.TryLock(dep, task.ID, ConflictTypeFileAccess) {
			return fmt.Errorf("resource %s is locked", dep)
		}
	}
	return nil
}

// SpawnAgent creates a new agent in the pool.
func (o *Orchestrator) SpawnAgent(ctx context.Context) (string, error) {
	info, err := o.pool.Spawn()
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

// TerminateAgent removes an agent from the pool.
func (o *Orchestrator) TerminateAgent(agentID string) error {
	return o.pool.Terminate(agentID)
}

// SendMessage sends a message between agents.
func (o *Orchestrator) SendMessage(msg Message) error {
	return o.bus.Send(msg)
}

// Broadcast sends a message to all agents.
func (o *Orchestrator) Broadcast(msg Message) error {
	return o.bus.Broadcast(msg)
}

// GetResults returns the results channel.
func (o *Orchestrator) GetResults() <-chan ExecutionResult {
	return o.results
}

// GetStats returns orchestrator statistics.
func (o *Orchestrator) GetStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	completed := 0
	failed := 0
	for _, task := range o.tasks {
		if task.Status == "completed" {
			completed++
		} else if task.Status == "failed" {
			failed++
		}
	}

	return map[string]interface{}{
		"total_tasks":     len(o.tasks),
		"completed_tasks": completed,
		"failed_tasks":    failed,
		"active_agents":   o.pool.ActiveCount(),
		"total_agents":    o.pool.TotalCount(),
	}
}

// Shutdown gracefully shuts down the orchestrator.
func (o *Orchestrator) Shutdown() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Clear message bus handlers
	o.bus.handlers = make(map[string]func(Message))

	// Clear tasks
	o.tasks = make(map[string]*TaskStatus)

	return nil
}

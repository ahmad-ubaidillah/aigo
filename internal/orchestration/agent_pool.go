package orchestration

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AgentState represents the current state of an agent.
type AgentState int

const (
	AgentStateIdle AgentState = iota
	AgentStateBusy
	AgentStateTerminated
)

// AgentInfo holds information about a managed agent.
type AgentInfo struct {
	ID           string
	State        AgentState
	CreatedAt    time.Time
	LastActiveAt time.Time
	TaskCount    int
	Metadata     map[string]any
}

// AgentFactory is a function that creates a new agent instance.
type AgentFactory func(ctx context.Context, id string) (any, error)

// AgentPool manages a pool of agents for the orchestrator.
type AgentPool struct {
	mu         sync.RWMutex
	agents     map[string]*AgentInfo
	instances  map[string]any
	available  chan string
	maxAgents  int
	factory    AgentFactory
	ctx        context.Context
	active     atomic.Int32
	totalSpawn atomic.Int32
}

// NewAgentPool creates a new agent pool.
func NewAgentPool(ctx context.Context, maxAgents int, factory AgentFactory) *AgentPool {
	if maxAgents <= 0 {
		maxAgents = 10
	}
	return &AgentPool{
		agents:    make(map[string]*AgentInfo),
		instances: make(map[string]any),
		available: make(chan string, maxAgents),
		maxAgents: maxAgents,
		factory:   factory,
		ctx:       ctx,
	}
}

// Spawn creates a new agent instance in the pool.
func (p *AgentPool) Spawn() (*AgentInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check max limit
	if len(p.agents) >= p.maxAgents {
		return nil, fmt.Errorf("agent pool at maximum capacity (%d)", p.maxAgents)
	}

	// Generate agent ID
	id := generateAgentID()

	// Create agent instance
	instance, err := p.factory(p.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	now := time.Now()
	info := &AgentInfo{
		ID:           id,
		State:        AgentStateIdle,
		CreatedAt:    now,
		LastActiveAt: now,
		TaskCount:    0,
		Metadata:     make(map[string]any),
	}

	p.agents[id] = info
	p.instances[id] = instance
	p.totalSpawn.Add(1)

	// Add to available pool
	p.available <- id

	return info, nil
}

// Acquire gets an available agent from the pool.
// Returns the agent ID and instance, or an error if none available.
func (p *AgentPool) Acquire(ctx context.Context) (string, any, error) {
	select {
	case id := <-p.available:
		p.mu.Lock()
		defer p.mu.Unlock()

		info, exists := p.agents[id]
		if !exists {
			return "", nil, fmt.Errorf("agent %s not found", id)
		}

		instance, exists := p.instances[id]
		if !exists {
			return "", nil, fmt.Errorf("agent %s instance not found", id)
		}

		info.State = AgentStateBusy
		info.LastActiveAt = time.Now()
		p.active.Add(1)

		return id, instance, nil

	case <-ctx.Done():
		return "", nil, fmt.Errorf("acquire timeout: %w", ctx.Err())
	}
}

// Release returns an agent to the pool for reuse.
func (p *AgentPool) Release(agentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	info, exists := p.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if info.State == AgentStateTerminated {
		return fmt.Errorf("agent %s is terminated", agentID)
	}

	info.State = AgentStateIdle
	info.LastActiveAt = time.Now()
	p.active.Add(-1)

	// Return to available pool
	select {
	case p.available <- agentID:
	default:
		// Pool is full, this shouldn't happen normally
	}

	return nil
}

// Terminate removes an agent from the pool permanently.
func (p *AgentPool) Terminate(agentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	info, exists := p.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	info.State = AgentStateTerminated
	delete(p.agents, agentID)
	delete(p.instances, agentID)

	if info.State == AgentStateBusy {
		p.active.Add(-1)
	}

	return nil
}

// SetMaxAgents updates the maximum number of agents allowed in the pool.
// If the new max is less than current count, existing agents are not terminated.
func (p *AgentPool) SetMaxAgents(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxAgents = n
}

// GetAgentInfo returns information about a specific agent.
func (p *AgentPool) GetAgentInfo(agentID string) (*AgentInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	info, exists := p.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}
	return info, nil
}

// ListAgents returns information about all agents in the pool.
func (p *AgentPool) ListAgents() []*AgentInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*AgentInfo, 0, len(p.agents))
	for _, info := range p.agents {
		result = append(result, info)
	}
	return result
}

// ActiveCount returns the number of currently active agents.
func (p *AgentPool) ActiveCount() int {
	return int(p.active.Load())
}

// TotalCount returns the total number of agents in the pool.
func (p *AgentPool) TotalCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.agents)
}

// TotalSpawned returns the total number of agents spawned since creation.
func (p *AgentPool) TotalSpawned() int {
	return int(p.totalSpawn.Load())
}

// Close terminates all agents in the pool.
func (p *AgentPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id := range p.agents {
		delete(p.agents, id)
		delete(p.instances, id)
	}

	p.active.Store(0)
	return nil
}

// generateAgentID creates a unique agent identifier.
func generateAgentID() string {
	return fmt.Sprintf("agent-%s", time.Now().Format("20060102150405.999"))
}

// Package subagent implements smart sub-agent delegation inspired by
// Oh My OpenAgent (OMO) architecture. It provides specialized agent roles
// that can be spawned in parallel for complex tasks.
//
// Architecture (OMO-inspired):
//   - Sisyphus: Main orchestrator (task decomposition, delegation)
//   - Hephaestus: Builder (code writing, implementation)
//   - Oracle: Reasoner (analysis, architecture, logic)
//   - Explorer: Scout (code exploration, research)
//   - Librarian: Memory keeper (context, knowledge management)
//
// Each role has:
//   - Specialized system prompt
//   - Category-based model routing
//   - Lean context passing
//   - Parallel execution with goroutines
package subagent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Role defines a specialized agent persona.
type Role string

const (
	RoleOrchestrator Role = "orchestrator" // Sisyphus — plans, delegates, drives
	RoleBuilder      Role = "builder"      // Hephaestus — implements, codes
	RoleReasoner     Role = "reasoner"     // Oracle — analyzes, architects
	RoleExplorer     Role = "explorer"     // Explore — researches, scouts
	RoleMemory       Role = "memory"       // Librarian — manages knowledge
)

// Category maps to model selection (OMO-style).
type Category string

const (
	CategoryDeep       Category = "deep"        // Autonomous research + execution
	CategoryQuick      Category = "quick"       // Single-file changes, quick fixes
	CategoryBrain      Category = "ultrabrain"  // Hard logic, architecture
	CategoryVisual     Category = "visual"      // Frontend, UI/UX
	CategoryGeneral    Category = "general"     // Default routing
)

// SubAgent represents a specialized agent instance.
type SubAgent struct {
	ID         string    `json:"id"`
	Role       Role      `json:"role"`
	Category   Category  `json:"category"`
	Goal       string    `json:"goal"`
	Status     string    `json:"status"` // pending, running, done, failed
	Result     string    `json:"result,omitempty"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	Iterations int       `json:"iterations"`
}

// Task is a unit of work assigned to a sub-agent.
type Task struct {
	ID          string   `json:"id"`
	Goal        string   `json:"goal"`
	Role        Role     `json:"role"`
	Category    Category `json:"category"`
	Context     string   `json:"context"`      // Lean context passed to sub-agent
	DependsOn   []string `json:"depends_on"`   // Task IDs this depends on
	Priority    int      `json:"priority"`     // Higher = more important
}

// Orchestrator manages sub-agent lifecycle and task decomposition.
type Orchestrator struct {
	mu         sync.RWMutex
	agents     map[string]*SubAgent
	tasks      map[string]*Task
	maxAgents  int
	brainFunc  BrainFunc
	sendFunc   SendFunc
	results    map[string]string // taskID → result
}

// BrainFunc is the function that calls the LLM for reasoning.
type BrainFunc func(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error)

// SendFunc delivers a message to the user.
type SendFunc func(message string)

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(brainFunc BrainFunc, sendFunc SendFunc) *Orchestrator {
	return &Orchestrator{
		agents:    make(map[string]*SubAgent),
		tasks:     make(map[string]*Task),
		maxAgents: 5, // OMO fires 5+ agents in parallel
		brainFunc: brainFunc,
		sendFunc:  sendFunc,
		results:   make(map[string]string),
	}
}

// Decompose breaks a complex goal into subtasks.
// This is the "Prometheus Planner" equivalent — interview mode analysis.
func (o *Orchestrator) Decompose(ctx context.Context, goal string) ([]Task, error) {
	prompt := fmt.Sprintf(`You are a task decomposer. Break this goal into independent subtasks that can run in parallel.

Goal: %s

For each subtask, specify:
1. A clear, specific goal
2. Which role should handle it (builder/reasoner/explorer/memory)
3. Which category fits (deep/quick/ultrabrain/visual/general)
4. Dependencies on other subtask IDs (if any)

Respond as JSON array:
[{"id":"t1","goal":"...","role":"builder","category":"deep","depends_on":[],"priority":1}]

Rules:
- Max 5 subtasks
- Minimize dependencies (prefer parallel)
- "quick" for simple fixes, "deep" for autonomous work
- "ultrabrain" for architecture decisions
- Be specific, not vague`, goal)

	response, err := o.brainFunc(ctx, SystemPrompt(RoleOrchestrator), prompt, 2000)
	if err != nil {
		return nil, fmt.Errorf("decompose: %w", err)
	}

	// Parse tasks from response
	tasks, err := parseTasks(response)
	if err != nil {
		// Fallback: single task
		return []Task{{
			ID:       "t1",
			Goal:     goal,
			Role:     RoleBuilder,
			Category: CategoryDeep,
			Priority: 1,
		}}, nil
	}

	return tasks, nil
}

// Execute runs tasks with parallel sub-agents.
// Follows OMO pattern: fire specialists in parallel, lean context, results when ready.
func (o *Orchestrator) Execute(ctx context.Context, tasks []Task) (map[string]string, error) {
	o.mu.Lock()
	// Build dependency graph
	completed := make(map[string]bool)
	for _, t := range tasks {
		o.tasks[t.ID] = &t
	}
	o.mu.Unlock()

	// Execute in waves based on dependencies
	wave := 0
	maxWaves := 10
	for len(completed) < len(tasks) && wave < maxWaves {
		wave++

		// Find tasks whose dependencies are all satisfied
		var ready []Task
		o.mu.RLock()
		for _, t := range tasks {
			if completed[t.ID] {
				continue
			}
			allDepsDone := true
			for _, dep := range t.DependsOn {
				if !completed[dep] {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				ready = append(ready, t)
			}
		}
		o.mu.RUnlock()

		if len(ready) == 0 {
			break // Deadlock or all done
		}

		// Enrich context with results from dependencies
		for i := range ready {
			for _, dep := range ready[i].DependsOn {
				if res, ok := o.results[dep]; ok {
					ready[i].Context += fmt.Sprintf("\n\n[Result from %s]:\n%s", dep, res)
				}
			}
		}

		// Execute ready tasks in parallel
		var wg sync.WaitGroup
		for _, t := range ready {
			if len(o.agents) >= o.maxAgents {
				break
			}
			wg.Add(1)
			go func(task Task) {
				defer wg.Done()
				result, err := o.runAgent(ctx, task)
				o.mu.Lock()
				if err != nil {
					completed[task.ID] = true // Mark as done even on error
					log.Printf("🤖 Sub-agent %s failed: %v", task.ID, err)
				} else {
					o.results[task.ID] = result
					completed[task.ID] = true
				}
				o.mu.Unlock()
			}(t)
		}
		wg.Wait()
	}

	return o.results, nil
}

// runAgent spawns a single sub-agent for a task.
func (o *Orchestrator) runAgent(ctx context.Context, task Task) (string, error) {
	agentID := fmt.Sprintf("%s-%d", task.ID, time.Now().UnixNano())
	agent := &SubAgent{
		ID:        agentID,
		Role:      task.Role,
		Category:  task.Category,
		Goal:      task.Goal,
		Status:    "running",
		StartedAt: time.Now(),
	}

	o.mu.Lock()
	o.agents[agentID] = agent
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		agent.FinishedAt = time.Now()
		agent.Status = "done"
		o.mu.Unlock()
	}()

	// Build specialized system prompt
	systemPrompt := SystemPrompt(task.Role)

	// Build lean user prompt with context
	userPrompt := buildLeanPrompt(task)

	// Execute with the brain function
	result, err := o.brainFunc(ctx, systemPrompt, userPrompt, 4000)
	if err != nil {
		agent.Status = "failed"
		agent.Error = err.Error()
		return "", err
	}

	agent.Result = result
	agent.Iterations++

	return result, nil
}

// GetAgent returns a sub-agent by ID.
func (o *Orchestrator) GetAgent(id string) (*SubAgent, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	a, ok := o.agents[id]
	return a, ok
}

// ListAgents returns all sub-agents.
func (o *Orchestrator) ListAgents() []*SubAgent {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var result []*SubAgent
	for _, a := range o.agents {
		result = append(result, a)
	}
	return result
}

// GetResults returns all task results.
func (o *Orchestrator) GetResults() map[string]string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	cp := make(map[string]string)
	for k, v := range o.results {
		cp[k] = v
	}
	return cp
}

// buildLeanPrompt creates a lean context prompt (OMO-style).
func buildLeanPrompt(task Task) string {
	var sb string
	sb += fmt.Sprintf("## Your Goal\n%s\n", task.Goal)
	if task.Context != "" {
		sb += fmt.Sprintf("\n## Context\n%s\n", task.Context)
	}
	sb += "\n## Instructions\n"
	sb += "- Be autonomous. Explore if needed. Don't ask for permission.\n"
	sb += "- Be specific. Include file paths, line numbers, exact changes.\n"
	sb += "- Report what you did in a structured summary.\n"
	return sb
}

// IntentAnalysis analyzes what the user really wants (OMO IntentGate).
func IntentAnalysis(ctx context.Context, brainFunc BrainFunc, userMessage string) (*IntentResult, error) {
	prompt := fmt.Sprintf(`Analyze the true intent behind this user message. Don't interpret literally — understand what they REALLY want.

Message: "%s"

Respond as JSON:
{
  "true_intent": "what the user actually wants",
  "category": "deep|quick|ultrabrain|visual|general",
  "complexity": 1-10,
  "needs_decomposition": true/false,
  "suggested_roles": ["builder", "reasoner", "explorer"],
  "risks": ["potential misunderstanding"],
  "clarification_needed": false
}`, userMessage)

	response, err := brainFunc(ctx, "You are an intent analysis expert. Be precise and concise.", prompt, 1000)
	if err != nil {
		return nil, err
	}

	intent, err := parseIntent(response)
	if err != nil {
		return &IntentResult{
			TrueIntent:  userMessage,
			Category:    CategoryGeneral,
			Complexity:  5,
			NeedsDecomp: false,
		}, nil
	}

	return intent, nil
}

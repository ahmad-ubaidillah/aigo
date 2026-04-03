package fleet

import (
	stdctx "context"
	"fmt"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/agent"
	"github.com/ahmad-ubaidillah/aigo/internal/cli"
	contextengine "github.com/ahmad-ubaidillah/aigo/internal/context"
	"github.com/ahmad-ubaidillah/aigo/internal/intent"
	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type AgentConfig struct {
	Name       string
	Role       string
	Priority   int
	MaxRetries int
	Timeout    time.Duration
}

type AgentState struct {
	Name        string
	Status      string
	CurrentTask string
	LastError   error
	Retries     int
	StartedAt   time.Time
	CompletedAt time.Time
}

type TaskResult struct {
	Agent    string
	Success  bool
	Output   string
	Error    error
	Retries  int
	Duration time.Duration
}

type AigoFleet struct {
	agents  map[string]*agent.Agent
	configs map[string]AgentConfig
	states  map[string]*AgentState
	queue   chan *FleetTask
	results map[string]*TaskResult
	mu      sync.RWMutex
	wg      sync.WaitGroup
	running bool
	ctx     stdctx.Context
	cancel  stdctx.CancelFunc
	db      *memory.SessionDB
}

type FleetTask struct {
	ID          string
	Description string
	Priority    int
	AssignedTo  string
	Result      *TaskResult
}

func NewAigoFleet(db *memory.SessionDB) *AigoFleet {
	ctx, cancel := stdctx.WithCancel(stdctx.Background())
	f := &AigoFleet{
		agents:  make(map[string]*agent.Agent),
		configs: make(map[string]AgentConfig),
		states:  make(map[string]*AgentState),
		queue:   make(chan *FleetTask, 100),
		results: make(map[string]*TaskResult),
		ctx:     ctx,
		cancel:  cancel,
		db:      db,
	}
	f.registerDefaultAgents()
	return f
}

func (f *AigoFleet) registerDefaultAgents() {
	defaultAgents := []AgentConfig{
		{
			Name:       "coordinator",
			Role:       "orchestrator",
			Priority:   10,
			MaxRetries: 3,
			Timeout:    5 * time.Minute,
		},
		{
			Name:       "coder",
			Role:       "coding",
			Priority:   8,
			MaxRetries: 5,
			Timeout:    10 * time.Minute,
		},
		{
			Name:       "researcher",
			Role:       "research",
			Priority:   7,
			MaxRetries: 3,
			Timeout:    3 * time.Minute,
		},
		{
			Name:       "reviewer",
			Role:       "review",
			Priority:   6,
			MaxRetries: 2,
			Timeout:    2 * time.Minute,
		},
		{
			Name:       "runner",
			Role:       "execution",
			Priority:   5,
			MaxRetries: 4,
			Timeout:    5 * time.Minute,
		},
	}

	for _, cfg := range defaultAgents {
		f.RegisterAgent(cfg)
	}
}

func (f *AigoFleet) RegisterAgent(cfg AgentConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.configs[cfg.Name] = cfg
	f.states[cfg.Name] = &AgentState{
		Name:      cfg.Name,
		Status:    "idle",
		Retries:   0,
		StartedAt: time.Now(),
	}
}

func (f *AigoFleet) Start() {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return
	}
	f.running = true
	f.mu.Unlock()

	for i := 0; i < 3; i++ {
		f.wg.Add(1)
		go f.worker(i)
	}
}

func (f *AigoFleet) Stop() {
	f.mu.Lock()
	if !f.running {
		f.mu.Unlock()
		return
	}
	f.running = false
	f.cancel()
	f.mu.Unlock()

	f.wg.Wait()
}

func (f *AigoFleet) worker(id int) {
	defer f.wg.Done()

	for {
		select {
		case <-f.ctx.Done():
			return
		case task := <-f.queue:
			f.executeTask(task)
		}
	}
}

func (f *AigoFleet) executeTask(task *FleetTask) {
	cfg := f.selectAgent(task.Description)
	if cfg.Name == "" {
		task.Result = &TaskResult{
			Success: false,
			Error:   fmt.Errorf("no suitable agent found"),
		}
		return
	}

	f.mu.Lock()
	f.states[cfg.Name].Status = "running"
	f.states[cfg.Name].CurrentTask = task.Description
	f.mu.Unlock()

	result := f.runWithRetry(task, cfg)
	task.Result = result

	f.mu.Lock()
	if result.Success {
		f.states[cfg.Name].Status = "idle"
	} else {
		f.states[cfg.Name].Status = "error"
		f.states[cfg.Name].LastError = result.Error
	}
	f.states[cfg.Name].CurrentTask = ""
	f.mu.Unlock()

	f.results[task.ID] = result
}

func (f *AigoFleet) runWithRetry(task *FleetTask, cfg AgentConfig) *TaskResult {
	start := time.Now()
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result := f.executeOnAgent(task.Description, cfg)

		if result.Success {
			result.Retries = attempt
			result.Duration = time.Since(start)
			return result
		}

		lastErr = result.Error

		if attempt < cfg.MaxRetries {
			delay := f.calculateRetryDelay(attempt, result.Error)
			select {
			case <-f.ctx.Done():
				return &TaskResult{Success: false, Error: f.ctx.Err()}
			case <-time.After(delay):
			}
		}
	}

	return &TaskResult{
		Agent:    cfg.Name,
		Success:  false,
		Error:    lastErr,
		Retries:  cfg.MaxRetries,
		Duration: time.Since(start),
	}
}

func (f *AigoFleet) calculateRetryDelay(attempt int, err error) time.Duration {
	baseDelays := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
	}

	if attempt < len(baseDelays) {
		return baseDelays[attempt]
	}

	exponential := time.Duration(attempt*attempt) * time.Second
	if exponential > 30*time.Second {
		exponential = 30 * time.Second
	}
	return exponential
}

func (f *AigoFleet) executeOnAgent(taskDesc string, cfg AgentConfig) *TaskResult {
	classifier := intent.NewClassifier(cli.DefaultConfig())
	ctxEngine := contextengine.NewContextEngine(f.db, cli.DefaultConfig())
	router := agent.NewRouter(f.db, cli.DefaultConfig(), nil)

	a := agent.NewAgent(classifier, router, ctxEngine, f.db, cli.DefaultConfig(), cfg.Name)

	ctx, cancel := stdctx.WithTimeout(f.ctx, cfg.Timeout)
	defer cancel()

	result, err := a.RunSession(ctx, cfg.Name, taskDesc)
	if err != nil {
		return &TaskResult{
			Agent:   cfg.Name,
			Success: false,
			Error:   err,
		}
	}

	return &TaskResult{
		Agent:   cfg.Name,
		Success: result.Success,
		Output:  result.Output,
		Error:   nil,
	}
}

func (f *AigoFleet) selectAgent(task string) AgentConfig {
	classifier := intent.NewClassifier(cli.DefaultConfig())
	classification := classifier.Classify(task)

	f.mu.RLock()
	defer f.mu.RUnlock()

	for name, cfg := range f.configs {
		if f.states[name].Status == "idle" || f.states[name].Status == "error" {
			switch classification.Intent {
			case types.IntentCoding:
				if cfg.Role == "coding" || cfg.Role == "orchestrator" {
					return cfg
				}
			case types.IntentWeb, types.IntentResearch:
				if cfg.Role == "research" {
					return cfg
				}
			case types.IntentSkill:
				if cfg.Role == "execution" {
					return cfg
				}
			}
		}
	}

	for name, cfg := range f.configs {
		if f.states[name].Status == "idle" {
			return cfg
		}
	}

	return AgentConfig{}
}

func (f *AigoFleet) Submit(taskDescription string) string {
	task := &FleetTask{
		ID:          fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Description: taskDescription,
		Priority:    1,
	}

	f.queue <- task
	return task.ID
}

func (f *AigoFleet) SubmitWithAgent(taskDescription, agentName string) string {
	task := &FleetTask{
		ID:          fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Description: taskDescription,
		AssignedTo:  agentName,
	}

	f.queue <- task
	return task.ID
}

func (f *AigoFleet) GetResult(taskID string) (*TaskResult, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result, ok := f.results[taskID]
	return result, ok
}

func (f *AigoFleet) GetAgentStates() map[string]*AgentState {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]*AgentState)
	for k, v := range f.states {
		result[k] = v
	}
	return result
}

func (f *AigoFleet) GetQueueLength() int {
	return len(f.queue)
}

func (f *AigoFleet) WaitForResult(taskID string, timeout time.Duration) (*TaskResult, error) {
	start := time.Now()
	for {
		result, ok := f.GetResult(taskID)
		if ok {
			return result, nil
		}

		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout waiting for result")
		}

		select {
		case <-f.ctx.Done():
			return nil, f.ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

type RetryStrategy interface {
	ShouldRetry(attempt int, err error) bool
	GetDelay(attempt int) time.Duration
}

type MantisClawStrategy struct {
	maxRetries     int
	baseDelays     []time.Duration
	retryableCodes []int
}

func NewMantisClawStrategy() *MantisClawStrategy {
	return &MantisClawStrategy{
		maxRetries: 5,
		baseDelays: []time.Duration{
			100 * time.Millisecond,
			500 * time.Millisecond,
			1 * time.Second,
			2 * time.Second,
			5 * time.Second,
		},
		retryableCodes: []int{
			types.ErrCodeTimeout,
			types.ErrCodeRateLimit,
			types.ErrCodeExternal,
		},
	}
}

func (m *MantisClawStrategy) ShouldRetry(attempt int, err error) bool {
	if attempt >= m.maxRetries {
		return false
	}

	if err == nil {
		return false
	}

	e, ok := err.(*types.Error)
	if !ok {
		return true
	}

	for _, code := range m.retryableCodes {
		if e.Code == code {
			return true
		}
	}

	return false
}

func (m *MantisClawStrategy) GetDelay(attempt int) time.Duration {
	if attempt < len(m.baseDelays) {
		return m.baseDelays[attempt]
	}

	exponential := time.Duration(attempt*attempt) * time.Second
	if exponential > 30*time.Second {
		exponential = 30 * time.Second
	}
	return exponential
}

func (f *AigoFleet) ExecuteWithRetry(task string, strategy RetryStrategy) *TaskResult {
	attempt := 0
	start := time.Now()

	for {
		result := f.executeSingle(task)

		if result.Success || !strategy.ShouldRetry(attempt, result.Error) {
			result.Retries = attempt
			result.Duration = time.Since(start)
			return result
		}

		attempt++

		delay := strategy.GetDelay(attempt)
		select {
		case <-f.ctx.Done():
			return &TaskResult{Success: false, Error: f.ctx.Err()}
		case <-time.After(delay):
		}
	}
}

func (f *AigoFleet) executeSingle(task string) *TaskResult {
	cfg := f.selectAgent(task)
	if cfg.Name == "" {
		return &TaskResult{Success: false, Error: fmt.Errorf("no agent available")}
	}

	return f.executeOnAgent(task, cfg)
}

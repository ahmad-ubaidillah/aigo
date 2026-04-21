package hooks

import (
	"fmt"
	"sync"
	"time"
)

var HookTypes = []string{
	"on_start",
	"on_stop",
	"on_error",
	"on_tool_start",
	"on_tool_end",
	"on_tool_error",
	"on_session_start",
	"on_session_end",
	"on_message",
	"on_response",
	"on_planning_start",
	"on_planning_end",
	"on_execution_start",
	"on_execution_end",
	"on_distill_start",
	"on_distill_end",
	"on_memory_store",
	"on_memory_retrieve",
	"on_plan_create",
	"on_plan_approve",
	"on_plan_reject",
	"on_agent_spawn",
	"on_agent_exit",
	"on_subagent_create",
	"on_subagent_result",
	"on_gap_detected",
	"on_ambiguity_detected",
	"on_decision_made",
	"on_context_expand",
	"on_context_compact",
	"on_token_limit",
	"on_doom_loop",
	"on_retry",
	"on_retry_success",
	"on_retry_fail",
	"on_user_input",
	"on_user_feedback",
	"on_skill_load",
	"on_skill_unload",
	"on_mcp_connect",
	"on_mcp_disconnect",
	"on_gateway_connect",
	"on_gateway_disconnect",
	"on_config_change",
	"on_hot_reload",
	"pre_planning",
	"post_planning",
}

type HookFunc func() error

type HookContext struct {
	Event     string
	Timestamp time.Time
	Data      map[string]interface{}
}

type HookRegistry struct {
	hooks map[string]HookFunc
	mu    sync.RWMutex
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{hooks: make(map[string]HookFunc)}
}

func (r *HookRegistry) Register(name string, fn HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[name] = fn
}

func (r *HookRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.hooks, name)
}

func (r *HookRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.hooks[name]
	return ok
}

func (r *HookRegistry) Execute(name string) error {
	r.mu.RLock()
	fn, ok := r.hooks[name]
	r.mu.RUnlock()

	if !ok {
		return nil
	}
	return fn()
}

func (r *HookRegistry) ExecuteWithContext(name string, ctx *HookContext) error {
	r.mu.RLock()
	fn, ok := r.hooks[name]
	r.mu.RUnlock()

	if !ok {
		return nil
	}
	return fn()
}

func (r *HookRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.hooks))
	for name := range r.hooks {
		result = append(result, name)
	}
	return result
}

type HookExecutor struct {
	registry *HookRegistry
	log      []HookLog
	mu       sync.Mutex
}

type HookLog struct {
	HookName   string    `json:"hook_name"`
	ExecutedAt time.Time `json:"executed_at"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

func NewHookExecutor(reg *HookRegistry) *HookExecutor {
	return &HookExecutor{
		registry: reg,
		log:      make([]HookLog, 0),
	}
}

func (e *HookExecutor) Trigger(event string) error {
	start := time.Now()
	err := e.registry.Execute(event)

	e.mu.Lock()
	defer e.mu.Unlock()
	e.log = append(e.log, HookLog{
		HookName:   event,
		ExecutedAt: start,
		Success:   err == nil,
		Error:     fmt.Sprintf("%v", err),
	})

	return err
}

func (e *HookExecutor) TriggerOnStart() error {
	return e.Trigger("on_start")
}

func (e *HookExecutor) TriggerOnStop() error {
	return e.Trigger("on_stop")
}

func (e *HookExecutor) TriggerOnError(err error) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, l := range e.log {
		l.Error = fmt.Sprintf("%v", err)
	}

	return e.Trigger("on_error")
}

func (e *HookExecutor) TriggerOnToolStart(name string) error {
	return e.Trigger("on_tool_start")
}

func (e *HookExecutor) TriggerOnToolEnd(name string) error {
	return e.Trigger("on_tool_end")
}

func (e *HookExecutor) TriggerOnSessionStart(id string) error {
	return e.Trigger("on_session_start")
}

func (e *HookExecutor) TriggerOnSessionEnd(id string) error {
	return e.Trigger("on_session_end")
}

func (e *HookExecutor) GetLog(limit int) []HookLog {
	e.mu.Lock()
	defer e.mu.Unlock()

	if limit > 0 && limit < len(e.log) {
		return e.log[len(e.log)-limit:]
	}
	return e.log
}

func (e *HookExecutor) GetLogByHook(hookName string) []HookLog {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]HookLog, 0)
	for _, l := range e.log {
		if l.HookName == hookName {
			result = append(result, l)
		}
	}
	return result
}

func (e *HookExecutor) ClearLog() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.log = make([]HookLog, 0)
}
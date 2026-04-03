package hooks

import (
	"fmt"
	"sync"
	"time"
)

const (
	HookGatewayStartup = "gateway:startup"
	HookSessionStart   = "session:start"
	HookSessionEnd     = "session:end"
	HookAgentStart     = "agent:start"
	HookAgentStep      = "agent:step"
	HookAgentEnd       = "agent:end"
)

type HookEvent struct {
	Name      string
	Payload   map[string]string
	Timestamp time.Time
}

type HookHandler interface {
	Handle(event HookEvent) error
}

type HookFunc func(event HookEvent) error

func (f HookFunc) Handle(event HookEvent) error { return f(event) }

type HookRegistry struct {
	hooks map[string][]HookHandler
	mu    sync.RWMutex
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{hooks: make(map[string][]HookHandler)}
}

func (r *HookRegistry) Register(hookType string, handler HookHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[hookType] = append(r.hooks[hookType], handler)
}

func (r *HookRegistry) Fire(hookType string, payload map[string]string) []error {
	r.mu.RLock()
	handlers := r.hooks[hookType]
	r.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	event := HookEvent{
		Name:      hookType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	var errs []error
	for _, h := range handlers {
		if err := h.Handle(event); err != nil {
			errs = append(errs, fmt.Errorf("hook %s: %w", hookType, err))
		}
	}
	return errs
}

func (r *HookRegistry) ListHooks(hookType string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hooks[hookType])
}

func (r *HookRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks = make(map[string][]HookHandler)
}

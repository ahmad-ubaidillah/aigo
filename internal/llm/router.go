package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type LLMRouter struct {
	providers []NamedProvider
	fallback  []string
	health    map[string]*ProviderHealth
	mu        sync.RWMutex
}

type ProviderHealth struct {
	Name             string
	LastCheck        time.Time
	IsHealthy        bool
	ConsecutiveFails int
	LastError        string
}

func NewLLMRouter(providers []NamedProvider, fallback []string) *LLMRouter {
	r := &LLMRouter{
		providers: providers,
		fallback:  fallback,
		health:    make(map[string]*ProviderHealth),
	}
	for _, p := range providers {
		r.health[p.Name] = &ProviderHealth{
			Name:             p.Name,
			LastCheck:        time.Now(),
			IsHealthy:        true,
			ConsecutiveFails: 0,
			LastError:        "",
		}
	}
	return r
}

func (r *LLMRouter) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	r.mu.RLock()
	ordered := r.orderedProviders()
	r.mu.RUnlock()

	var lastErr error
	for _, p := range ordered {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := r.chatWithProvider(ctx, p, messages, opts)
		if err == nil {
			r.markSuccess(p.Name)
			return resp, nil
		}

		r.markFailure(p.Name, err)
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed (last error: %w)", lastErr)
}

func (r *LLMRouter) orderedProviders() []NamedProvider {
	result := make([]NamedProvider, 0)
	seen := make(map[string]bool)

	for _, p := range r.providers {
		if seen[p.Name] {
			continue
		}
		seen[p.Name] = true

		h := r.health[p.Name]
		if h.IsHealthy && h.ConsecutiveFails < 3 {
			result = append(result, p)
		}
	}

	for _, name := range r.fallback {
		if !seen[name] {
			for _, p := range r.providers {
				if p.Name == name {
					seen[name] = true
					result = append(result, p)
					break
				}
			}
		}
	}

	for _, p := range r.providers {
		if !seen[p.Name] {
			result = append(result, p)
		}
	}

	return result
}

func (r *LLMRouter) chatWithProvider(ctx context.Context, p NamedProvider, messages []Message, opts ChatOptions) (*ChatResponse, error) {
	timeout := p.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if chatter, ok := p.Client.(Chatter); ok {
		return chatter.ChatWithOptions(ctx, messages, opts)
	}

	if basic, ok := p.Client.(interface {
		Chat(context.Context, []Message) (*Response, error)
	}); ok {
		resp, err := basic.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}
		return &ChatResponse{
			Content: resp.Content,
			Usage:   TokenUsage(resp.Usage),
		}, nil
	}

	return nil, fmt.Errorf("provider %s does not support chat", p.Name)
}

func (r *LLMRouter) markSuccess(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.health[name]; ok {
		h.ConsecutiveFails = 0
		h.IsHealthy = true
		h.LastCheck = time.Now()
		h.LastError = ""
	}
}

func (r *LLMRouter) markFailure(name string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.health[name]; ok {
		h.ConsecutiveFails++
		h.LastError = err.Error()
		h.LastCheck = time.Now()
		if h.ConsecutiveFails >= 3 {
			h.IsHealthy = false
		}
	}
}

func (r *LLMRouter) Health() map[string]*ProviderHealth {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]*ProviderHealth)
	for k, v := range r.health {
		copied := *v
		result[k] = &copied
	}
	return result
}

func (r *LLMRouter) CheckHealth(ctx context.Context) map[string]bool {
	results := make(map[string]bool)
	for _, p := range r.providers {
		healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := r.chatWithProvider(healthCtx, p, []Message{{Role: "user", Content: "ping"}}, ChatOptions{})
		cancel()
		results[p.Name] = err == nil
	}
	return results
}

func (r *LLMRouter) Complete(ctx context.Context, prompt string) (*Response, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	resp, err := r.Chat(ctx, messages, ChatOptions{})
	if err != nil {
		return nil, err
	}
	return &Response{Content: resp.Content, Usage: Usage(resp.Usage)}, nil
}

func (r *LLMRouter) CompleteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	resp, err := r.Chat(ctx, messages, ChatOptions{})
	if err != nil {
		return nil, err
	}
	return &Response{Content: resp.Content, Usage: Usage(resp.Usage)}, nil
}

func NewRouterFromConfig(cfg types.LLMConfig) (*LLMRouter, error) {
	providers, err := NewProviders(cfg)
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}
	return NewLLMRouter(providers, cfg.Fallback), nil
}

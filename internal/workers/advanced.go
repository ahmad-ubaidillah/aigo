package workers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BoulderMode enables persistent execution that cannot stop until task complete.
type BoulderMode struct {
	maxAttempts    int
	currentAttempt int
	task           string
}

// NewBoulderMode creates a new boulder mode config.
func NewBoulderMode(maxAttempts int) *BoulderMode {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	return &BoulderMode{maxAttempts: maxAttempts}
}

// Execute runs the task with boulder continuation.
func (b *BoulderMode) Execute(ctx context.Context, fn func() error) error {
	for b.currentAttempt < b.maxAttempts {
		b.currentAttempt++
		err := fn()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return fmt.Errorf("boulder mode: %d attempts exhausted for task: %s", b.maxAttempts, b.task)
}

// Attempt returns the current attempt number.
func (b *BoulderMode) Attempt() int {
	return b.currentAttempt
}

// MultiProviderWorker tries multiple providers with fallback.
type MultiProviderWorker struct {
	providers []Worker
	mu        sync.RWMutex
	active    int
}

// NewMultiProviderWorker creates a worker with fallback providers.
func NewMultiProviderWorker(providers ...Worker) *MultiProviderWorker {
	return &MultiProviderWorker{providers: providers}
}

// Execute tries each provider in order until one succeeds.
func (m *MultiProviderWorker) Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error) {
	var lastErr error
	for i, p := range m.providers {
		m.mu.Lock()
		m.active = i
		m.mu.Unlock()

		result, err := p.Execute(ctx, task, params)
		if err == nil && result.Success {
			return result, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	return &WorkerResult{
		Success:  false,
		Error:    fmt.Sprintf("all %d providers failed, last error: %v", len(m.providers), lastErr),
		Duration: time.Second,
	}, lastErr
}

// ActiveProvider returns the index of the currently active provider.
func (m *MultiProviderWorker) ActiveProvider() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.active
}

// ProviderCount returns the number of configured providers.
func (m *MultiProviderWorker) ProviderCount() int {
	return len(m.providers)
}

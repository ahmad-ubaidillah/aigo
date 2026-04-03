package workers

import (
	"context"
	"fmt"
	"time"
)

// WorkerResult holds the result from a worker execution
type WorkerResult struct {
	Success  bool
	Output   string
	Error    string
	Metadata map[string]string
	Duration time.Duration
}

// Worker is the generic interface for all worker types
type Worker interface {
	Name() string
	Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error)
}

// Sisyphus orchestrates tasks across workers
type Sisyphus struct{}

func NewSisyphus() *Sisyphus { return &Sisyphus{} }

func (s *Sisyphus) Name() string { return "sisyphus" }

func (s *Sisyphus) Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error) {
	res := &WorkerResult{
		Success:  true,
		Output:   "Sisyphus orchestrator: task queued",
		Metadata: map[string]string{"task": task},
	}
	// We simulate duration minimally
	res.Duration = time.Millisecond * 10
	return res, nil
}

// Hephaestus coding agent
type Hephaestus struct{}

func NewHephaestus() *Hephaestus { return &Hephaestus{} }

func (h *Hephaestus) Name() string { return "hephaestus" }

func (h *Hephaestus) Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error) {
	res := &WorkerResult{
		Success:  true,
		Output:   "Hephaestus coding agent: task received",
		Metadata: map[string]string{"task": task},
	}
	res.Duration = time.Millisecond * 15
	return res, nil
}

func (h *Hephaestus) HashAnchoredEdit(path, oldStr, newStr string) error {
	// Stub for anchoring edits
	return nil
}

// Oracle consultant
type Oracle struct{}

func NewOracle() *Oracle { return &Oracle{} }

func (o *Oracle) Name() string { return "oracle" }

func (o *Oracle) Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error) {
	res := &WorkerResult{
		Success:  true,
		Output:   "Oracle analysis: reviewing architecture",
		Metadata: map[string]string{"task": task},
	}
	res.Duration = time.Millisecond * 12
	return res, nil
}

// Librarian searches documentation
type Librarian struct{}

func NewLibrarian() *Librarian { return &Librarian{} }

func (l *Librarian) Name() string { return "librarian" }

func (l *Librarian) Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error) {
	res := &WorkerResult{
		Success:  true,
		Output:   "Librarian search: query received",
		Metadata: map[string]string{"task": task},
	}
	res.Duration = time.Millisecond * 14
	return res, nil
}

// Explore exploration worker
type Explore struct{}

func NewExplore() *Explore { return &Explore{} }

func (e *Explore) Name() string { return "explore" }

func (e *Explore) Execute(ctx context.Context, task string, params map[string]any) (*WorkerResult, error) {
	res := &WorkerResult{
		Success:  true,
		Output:   "Explore: scanning codebase",
		Metadata: map[string]string{"task": task},
	}
	res.Duration = time.Millisecond * 20
	return res, nil
}

// WorkerPool manages registered workers
type WorkerPool struct {
	workers map[string]Worker
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{workers: make(map[string]Worker)}
}

func (wp *WorkerPool) Register(w Worker) {
	if wp.workers == nil {
		wp.workers = make(map[string]Worker)
	}
	wp.workers[w.Name()] = w
}

func (wp *WorkerPool) Get(name string) Worker {
	if wp.workers == nil {
		return nil
	}
	return wp.workers[name]
}

func (wp *WorkerPool) List() []string {
	names := make([]string, 0, len(wp.workers))
	for name := range wp.workers {
		names = append(names, name)
	}
	return names
}

func (wp *WorkerPool) Execute(ctx context.Context, name, task string, params map[string]any) (*WorkerResult, error) {
	w := wp.Get(name)
	if w == nil {
		return nil, fmt.Errorf("worker %s not found", name)
	}
	return w.Execute(ctx, task, params)
}

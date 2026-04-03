package orchestration

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ExecutionTask represents a task to be executed by a strategy.
type ExecutionTask struct {
	ID          string
	Name        string
	Description string
	Execute     func(ctx context.Context) (any, error)
	Timeout     time.Duration
	Priority    int
	DependsOn   []string // Task IDs this task depends on
}

// ExecutionResult represents the result of executing a task.
type ExecutionResult struct {
	TaskID    string
	Result    any
	Error     error
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// ExecutionStrategy defines how tasks are executed.
type ExecutionStrategy interface {
	Execute(ctx context.Context, tasks []ExecutionTask) ([]ExecutionResult, error)
	Name() string
}

// SequentialStrategy executes tasks one by one in order.
type SequentialStrategy struct {
	name string
}

// NewSequentialStrategy creates a new sequential execution strategy.
func NewSequentialStrategy() *SequentialStrategy {
	return &SequentialStrategy{name: "sequential"}
}

// Execute runs tasks one at a time.
func (s *SequentialStrategy) Execute(ctx context.Context, tasks []ExecutionTask) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, len(tasks))

	for i, task := range tasks {
		startTime := time.Now()

		// Create timeout context if specified
		execCtx := ctx
		if task.Timeout > 0 {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, task.Timeout)
			defer cancel()
		}

		result := ExecutionResult{
			TaskID:    task.ID,
			StartTime: startTime,
		}

		// Execute the task
		res, err := task.Execute(execCtx)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		result.Result = res
		result.Error = err
		results[i] = result

		// If task failed, stop execution
		if err != nil {
			return results, fmt.Errorf("task %s failed: %w", task.ID, err)
		}

		// Check if context is cancelled
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
	}

	return results, nil
}

// Name returns the strategy name.
func (s *SequentialStrategy) Name() string {
	return s.name
}

// ParallelStrategy executes independent tasks concurrently.
type ParallelStrategy struct {
	name        string
	maxWorkers  int
}

// NewParallelStrategy creates a new parallel execution strategy.
func NewParallelStrategy(maxWorkers int) *ParallelStrategy {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	return &ParallelStrategy{
		name:       "parallel",
		maxWorkers: maxWorkers,
	}
}

// Execute runs tasks concurrently, respecting dependencies.
func (s *ParallelStrategy) Execute(ctx context.Context, tasks []ExecutionTask) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, len(tasks))
	resultsMap := make(map[string]*ExecutionResult)
	taskMap := make(map[string]*ExecutionTask)
	taskIndices := make(map[string]int)

	for i := range tasks {
		task := &tasks[i]
		results[i] = ExecutionResult{TaskID: task.ID}
		resultsMap[task.ID] = &results[i]
		taskMap[task.ID] = task
		taskIndices[task.ID] = i
	}

	// Build dependency graph
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)
	for _, task := range tasks {
		inDegree[task.ID] = len(task.DependsOn)
		for _, dep := range task.DependsOn {
			dependents[dep] = append(dependents[dep], task.ID)
		}
	}

	// Find tasks with no dependencies (ready to run)
	ready := make(chan string, len(tasks))
	for _, task := range tasks {
		if inDegree[task.ID] == 0 {
			ready <- task.ID
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var execErr error
	activeWorkers := 0

	// Process tasks
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return results, ctx.Err()

		case taskID := <-ready:
			if taskID == "" {
				// Channel closed
				continue
			}

			mu.Lock()
			if execErr != nil {
				mu.Unlock()
				continue
			}
			if activeWorkers >= s.maxWorkers {
				// Put back and wait
				go func() {
					time.Sleep(10 * time.Millisecond)
					ready <- taskID
				}()
				mu.Unlock()
				continue
			}
			activeWorkers++
			mu.Unlock()

			wg.Add(1)
			go func(tid string) {
				defer wg.Done()
				defer func() {
					mu.Lock()
					activeWorkers--
					mu.Unlock()
				}()

				task := taskMap[tid]
				startTime := time.Now()

				// Create timeout context if specified
				execCtx := ctx
				if task.Timeout > 0 {
					var cancel context.CancelFunc
					execCtx, cancel = context.WithTimeout(ctx, task.Timeout)
					defer cancel()
				}

				res, err := task.Execute(execCtx)
				endTime := time.Now()

				mu.Lock()
				result := resultsMap[tid]
				result.StartTime = startTime
				result.EndTime = endTime
				result.Duration = endTime.Sub(startTime)
				result.Result = res
				result.Error = err

				if err != nil {
					execErr = fmt.Errorf("task %s failed: %w", tid, err)
					mu.Unlock()
					return
				}
				mu.Unlock()

				// Update dependents
				mu.Lock()
				for _, depID := range dependents[tid] {
					inDegree[depID]--
					if inDegree[depID] == 0 {
						ready <- depID
					}
				}
				mu.Unlock()
			}(taskID)

		default:
			// Check if done
			mu.Lock()
			done := activeWorkers == 0 && len(ready) == 0
			mu.Unlock()

			if done {
				wg.Wait()
				if execErr != nil {
					return results, execErr
				}

				// Check if all tasks completed
				allComplete := true
				for _, t := range tasks {
					if _, exists := resultsMap[t.ID]; !exists || resultsMap[t.ID].EndTime.IsZero() {
						allComplete = false
						break
					}
				}

				if !allComplete {
					return results, fmt.Errorf("circular dependency detected")
				}

				return results, nil
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Name returns the strategy name.
func (s *ParallelStrategy) Name() string {
	return s.name
}

// RetryStrategy wraps another strategy with exponential backoff retry.
type RetryStrategy struct {
	name       string
	delegate   ExecutionStrategy
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
	multiplier float64
}

// RetryOption configures the retry strategy.
type RetryOption func(*RetryStrategy)

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) RetryOption {
	return func(r *RetryStrategy) {
		r.maxRetries = n
	}
}

// WithBaseDelay sets the initial delay between retries.
func WithBaseDelay(d time.Duration) RetryOption {
	return func(r *RetryStrategy) {
		r.baseDelay = d
	}
}

// WithMaxDelay sets the maximum delay between retries.
func WithMaxDelay(d time.Duration) RetryOption {
	return func(r *RetryStrategy) {
		r.maxDelay = d
	}
}

// WithMultiplier sets the delay multiplier for exponential backoff.
func WithMultiplier(m float64) RetryOption {
	return func(r *RetryStrategy) {
		r.multiplier = m
	}
}

// NewRetryStrategy creates a new retry execution strategy.
func NewRetryStrategy(delegate ExecutionStrategy, opts ...RetryOption) *RetryStrategy {
	r := &RetryStrategy{
		name:       "retry",
		delegate:   delegate,
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   30 * time.Second,
		multiplier: 2.0,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Execute runs tasks with retry on failure.
func (s *RetryStrategy) Execute(ctx context.Context, tasks []ExecutionTask) ([]ExecutionResult, error) {
	// Wrap each task with retry logic
	wrappedTasks := make([]ExecutionTask, len(tasks))
	for i, task := range tasks {
		task := task // capture
		wrappedTasks[i] = ExecutionTask{
			ID:          task.ID,
			Name:        task.Name,
			Description: task.Description,
			Timeout:     task.Timeout,
			Priority:    task.Priority,
			DependsOn:   task.DependsOn,
			Execute: func(ctx context.Context) (any, error) {
				var lastErr error
				for attempt := 0; attempt <= s.maxRetries; attempt++ {
					if attempt > 0 {
						// Calculate backoff delay
						delay := time.Duration(float64(s.baseDelay) * math.Pow(s.multiplier, float64(attempt-1)))
						if delay > s.maxDelay {
							delay = s.maxDelay
						}

						select {
						case <-time.After(delay):
						case <-ctx.Done():
							return nil, ctx.Err()
						}
					}

					// Create timeout context if specified
					execCtx := ctx
					if task.Timeout > 0 {
						var cancel context.CancelFunc
						execCtx, cancel = context.WithTimeout(ctx, task.Timeout)
						defer cancel()
					}

					result, err := task.Execute(execCtx)
					if err == nil {
						return result, nil
					}
					lastErr = err

					// Check if error is retryable
					if !isRetryable(err) {
						return nil, err
					}

					// Check if context is cancelled
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
				}
				return nil, fmt.Errorf("max retries (%d) exceeded: %w", s.maxRetries, lastErr)
			},
		}
	}

	return s.delegate.Execute(ctx, wrappedTasks)
}

// Name returns the strategy name.
func (s *RetryStrategy) Name() string {
	return s.name
}

// isRetryable determines if an error is retryable.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	// Check for context cancellation
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	// Most errors are retryable by default
	return true
}

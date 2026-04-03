package execution

import (
	"fmt"
	"sync"
	"time"
)

// ProgressReporter tracks and reports execution progress.
type ProgressReporter struct {
	mu           sync.RWMutex
	completed    int
	inProgress   int
	blocked      int
	pending      int
	total        int
	startedAt    time.Time
	lastReportAt time.Time
}

// NewProgressReporter creates a new progress reporter.
func NewProgressReporter(total int) *ProgressReporter {
	return &ProgressReporter{
		total:     total,
		pending:   total,
		startedAt: time.Now(),
	}
}

func (r *ProgressReporter) Update(completed, inProgress, blocked, pending int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed = completed
	r.inProgress = inProgress
	r.blocked = blocked
	r.pending = pending
	r.lastReportAt = time.Now()
}

// MarkComplete marks a task as completed.
func (r *ProgressReporter) MarkComplete() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed++
	if r.pending > 0 {
		r.pending--
	}
	r.lastReportAt = time.Now()
}

// MarkInProgress marks a task as in progress.
func (r *ProgressReporter) MarkInProgress() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inProgress++
	if r.pending > 0 {
		r.pending--
	}
}

// MarkBlocked marks a task as blocked.
func (r *ProgressReporter) MarkBlocked() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blocked++
	if r.inProgress > 0 {
		r.inProgress--
	}
}

// Percentage returns the completion percentage.
func (r *ProgressReporter) Percentage() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.total == 0 {
		return 0
	}
	return float64(r.completed) / float64(r.total) * 100
}

// ETA estimates time to completion based on current rate.
func (r *ProgressReporter) ETA() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.completed == 0 {
		return 0
	}
	elapsed := time.Since(r.startedAt)
	rate := float64(r.completed) / elapsed.Seconds()
	remaining := r.total - r.completed
	return time.Duration(float64(remaining)/rate) * time.Second
}

// Summary returns a human-readable progress summary.
func (r *ProgressReporter) Summary() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return fmt.Sprintf("%d/%d completed (%.0f%%) | %d in progress | %d blocked | %d pending",
		r.completed, r.total, r.Percentage(), r.inProgress, r.blocked, r.pending)
}

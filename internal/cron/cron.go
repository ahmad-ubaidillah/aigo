// Package cron implements job scheduling for Aigo.
// Supports recurring and one-shot jobs.
package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Job represents a scheduled task.
type Job struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Schedule    string    `json:"schedule"` // cron expression or ISO timestamp
	Prompt      string    `json:"prompt"`   // What to ask the agent
	Enabled     bool      `json:"enabled"`
	LastRun     time.Time `json:"last_run,omitempty"`
	NextRun     time.Time `json:"next_run,omitempty"`
	RunCount    int       `json:"run_count"`
}

// Handler is called when a job fires.
type Handler func(ctx context.Context, job Job) (string, error)

// Scheduler manages scheduled jobs.
type Scheduler struct {
	jobs     map[string]*Job
	mu       sync.RWMutex
	handler  Handler
	storagePath string
	running  bool
}

// New creates a new scheduler.
func New(storagePath string, handler Handler) *Scheduler {
	return &Scheduler{
		jobs:        make(map[string]*Job),
		handler:     handler,
		storagePath: storagePath,
	}
}

// Start begins the scheduler loop.
func (s *Scheduler) Start(ctx context.Context) {
	s.running = true
	s.loadJobs()

	log.Printf("Cron scheduler started with %d jobs", len(s.jobs))

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ctx.Done():
			s.running = false
			return
		case <-ticker.C:
			s.checkAndRun(ctx)
		}
	}
}

func (s *Scheduler) checkAndRun(ctx context.Context) {
	now := time.Now()

	s.mu.RLock()
	var toRun []*Job
	for _, job := range s.jobs {
		if job.Enabled && !job.NextRun.IsZero() && now.After(job.NextRun) {
			toRun = append(toRun, job)
		}
	}
	s.mu.RUnlock()

	for _, job := range toRun {
		go s.runJob(ctx, job)
	}
}

func (s *Scheduler) runJob(ctx context.Context, job *Job) {
	log.Printf("Cron: running job '%s'", job.Name)

	if s.handler != nil {
		result, err := s.handler(ctx, *job)
		if err != nil {
			log.Printf("Cron: job '%s' failed: %v", job.Name, err)
		} else {
			log.Printf("Cron: job '%s' completed: %s", job.Name, truncate(result, 100))
		}
	}

	s.mu.Lock()
	job.LastRun = time.Now()
	job.RunCount++
	job.NextRun = parseNextRun(job.Schedule, job.LastRun)
	s.mu.Unlock()

	s.saveJobs()
}

// Add adds a job.
func (s *Scheduler) Add(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job.ID == "" {
		job.ID = fmt.Sprintf("job_%d", time.Now().UnixNano())
	}
	if job.NextRun.IsZero() {
		job.NextRun = parseNextRun(job.Schedule, time.Now())
	}
	job.Enabled = true

	s.jobs[job.ID] = &job
	s.saveJobs()
	log.Printf("Cron: added job '%s' (next: %s)", job.Name, job.NextRun.Format(time.RFC3339))
}

// Remove removes a job.
func (s *Scheduler) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
	s.saveJobs()
}

// List returns all jobs.
func (s *Scheduler) List() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, *job)
	}
	return jobs
}

func (s *Scheduler) saveJobs() {
	if s.storagePath == "" {
		return
	}
	os.MkdirAll(filepath.Dir(s.storagePath), 0755)
	data, _ := json.MarshalIndent(s.jobs, "", "  ")
	os.WriteFile(s.storagePath, data, 0644)
}

func (s *Scheduler) loadJobs() {
	if s.storagePath == "" {
		return
	}
	data, err := os.ReadFile(s.storagePath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.jobs)
}

// parseNextRun calculates next run time from schedule.
// Supports: "@every 1h", "@daily", "@hourly", or ISO timestamp.
func parseNextRun(schedule string, from time.Time) time.Time {
	switch schedule {
	case "@hourly":
		return from.Add(1 * time.Hour)
	case "@daily":
		return from.Add(24 * time.Hour)
	case "@weekly":
		return from.Add(7 * 24 * time.Hour)
	}

	if len(schedule) > 7 && schedule[:7] == "@every " {
		d, err := time.ParseDuration(schedule[7:])
		if err == nil {
			return from.Add(d)
		}
	}

	// Try ISO timestamp
	if t, err := time.Parse(time.RFC3339, schedule); err == nil {
		return t
	}

	// Default: 1 hour
	return from.Add(1 * time.Hour)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

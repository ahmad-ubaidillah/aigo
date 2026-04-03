package cron

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Scheduler struct {
	jobs    map[int64]*types.CronSchedule
	queue   chan *types.CronSchedule
	runner  *Runner
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
	db      interface {
		ListCronJobs() ([]types.CronSchedule, error)
		UpdateCronJob(id int64, enabled bool, lastRun, nextRun time.Time) error
	}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:   make(map[int64]*types.CronSchedule),
		queue:  make(chan *types.CronSchedule, 100),
		runner: NewRunner(),
		stopCh: make(chan struct{}),
	}
}

func (s *Scheduler) SetDB(db interface {
	ListCronJobs() ([]types.CronSchedule, error)
	UpdateCronJob(id int64, enabled bool, lastRun, nextRun time.Time) error
}) {
	s.db = db
}

func (s *Scheduler) AddJob(job *types.CronSchedule) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	next := calculateNextRun(job.Schedule, time.Now())
	job.NextRun = &next
	s.jobs[job.ID] = job

	return nil
}

func (s *Scheduler) RemoveJob(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jobs[id]; !ok {
		return fmt.Errorf("job not found: %d", id)
	}

	delete(s.jobs, id)
	return nil
}

func (s *Scheduler) ListJobs() []types.CronSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]types.CronSchedule, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, *job)
	}
	return result
}

func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	if s.db != nil {
		go s.loadJobsFromDB(ctx)
	}

	go s.runLoop(ctx)
	go s.runner.ProcessQueue(ctx, s.queue)
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopCh)
}

func (s *Scheduler) loadJobsFromDB(ctx context.Context) {
	jobs, err := s.db.ListCronJobs()
	if err != nil {
		return
	}

	for i := range jobs {
		s.AddJob(&jobs[i])
	}
}

func (s *Scheduler) runLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndSchedule()
		}
	}
}

func (s *Scheduler) checkAndSchedule() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	for _, job := range s.jobs {
		if !job.Enabled {
			continue
		}

		if job.NextRun != nil && (job.NextRun.Before(now) || job.NextRun.Equal(now)) {
			schedule := &types.CronSchedule{
				ID:        job.ID,
				JobID:     job.JobID,
				Status:    "pending",
				CreatedAt: now,
			}

			select {
			case s.queue <- schedule:
				job.LastRun = &now
				next := calculateNextRun(job.Schedule, now)
				job.NextRun = &next

				if s.db != nil {
					s.db.UpdateCronJob(job.ID, job.Enabled, now, *job.NextRun)
				}
			default:
			}
		}
	}
}

func calculateNextRun(schedule string, now time.Time) time.Time {
	// Basic cron support: @hourly, @daily, @weekly, @monthly or simple intervals
	switch schedule {
	case "@hourly":
		return now.Add(1 * time.Hour)
	case "@daily":
		return now.Add(24 * time.Hour)
	case "@weekly":
		return now.Add(7 * 24 * time.Hour)
	case "@monthly":
		return now.Add(30 * 24 * time.Hour)
	default:
		// For custom cron expressions, default to hourly
		return now.Add(1 * time.Hour)
	}
}

func (s *Scheduler) GetQueue() <-chan *types.CronSchedule {
	return s.queue
}

type Runner struct {
	executors map[string]CronExecutor
	mu        sync.RWMutex
	wg        sync.WaitGroup
}

type CronExecutor interface {
	Execute(ctx context.Context, schedule *types.CronSchedule, command string) error
}

func NewRunner() *Runner {
	r := &Runner{
		executors: make(map[string]CronExecutor),
	}
	r.RegisterExecutor("shell", &ShellExecutor{})
	r.RegisterExecutor("telegram", &TelegramExecutor{})
	r.RegisterExecutor("discord", &DiscordExecutor{})
	return r
}

func (r *Runner) RegisterExecutor(platform string, exec CronExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[platform] = exec
}

func (r *Runner) Run(ctx context.Context, schedule *types.CronSchedule, platform, command string) error {
	r.mu.RLock()
	exec, ok := r.executors[platform]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no executor for platform: %s", platform)
	}

	return exec.Execute(ctx, schedule, command)
}

func (r *Runner) ProcessQueue(ctx context.Context, queue <-chan *types.CronSchedule) {
	for {
		select {
		case <-ctx.Done():
			return
		case schedule := <-queue:
			r.wg.Add(1)
			go func(s *types.CronSchedule) {
				defer r.wg.Done()
				r.executeJob(ctx, s)
			}(schedule)
		}
	}
}

func (r *Runner) executeJob(ctx context.Context, schedule *types.CronSchedule) {
	schedule.Status = "running"

	err := r.Run(ctx, schedule, "shell", schedule.Output)
	schedule.Status = "completed"
	if err != nil {
		schedule.Status = "failed"
		schedule.Output = err.Error()
	}
	now := time.Now()
	schedule.CompletedAt = &now
}

type ShellExecutor struct{}

func (e *ShellExecutor) Execute(ctx context.Context, schedule *types.CronSchedule, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(os.Environ(), "AIGO_CRON=1")

	out, err := cmd.CombinedOutput()
	schedule.Output = string(out)

	return err
}

type TelegramExecutor struct{}

func (e *TelegramExecutor) Execute(ctx context.Context, schedule *types.CronSchedule, command string) error {
	return fmt.Errorf("telegram executor not implemented - configure bot token first")
}

type DiscordExecutor struct{}

func (e *DiscordExecutor) Execute(ctx context.Context, schedule *types.CronSchedule, command string) error {
	return fmt.Errorf("discord executor not implemented - configure bot token first")
}

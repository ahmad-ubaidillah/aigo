package cron

import (
	"context"
	"testing"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestNewScheduler(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	if s == nil {
		t.Error("expected scheduler")
	}
}

func TestScheduler_AddJob(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	now := time.Now()
	next := now.Add(5 * time.Minute)
	job := &types.CronSchedule{
		ID:       1,
		Name:     "test",
		Schedule: "*/5 * * * *",
		Command:  "echo test",
		Enabled:  true,
		LastRun:  &now,
		NextRun:  &next,
	}
	err := s.AddJob(job)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScheduler_RemoveJob(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	s.AddJob(&types.CronSchedule{ID: 1, Name: "test", Schedule: "*/5 * * * *", Command: "echo test", Enabled: true})
	err := s.RemoveJob(1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScheduler_RemoveJobNotFound(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	err := s.RemoveJob(999)
	if err == nil {
		t.Error("expected error")
	}
}

func TestScheduler_ListJobs(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	s.AddJob(&types.CronSchedule{ID: 1, Name: "a", Schedule: "*/5 * * * *", Command: "echo a", Enabled: true})
	s.AddJob(&types.CronSchedule{ID: 2, Name: "b", Schedule: "0 12 * * *", Command: "echo b", Enabled: true})
	jobs := s.ListJobs()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestScheduler_StartStop(t *testing.T) {
	t.Parallel()
	s := NewScheduler()
	s.AddJob(&types.CronSchedule{ID: 1, Name: "test", Schedule: "*/5 * * * *", Command: "echo test", Enabled: true})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}

func TestNewRunner(t *testing.T) {
	t.Parallel()
	r := NewRunner()
	if r == nil {
		t.Error("expected runner")
	}
}

func TestCronSchedule(t *testing.T) {
	t.Parallel()
	cs := types.CronSchedule{ID: 1, Name: "test", Schedule: "*/5 * * * *", Command: "echo test"}
	if cs.Name != "test" {
		t.Errorf("expected test, got %s", cs.Name)
	}
}

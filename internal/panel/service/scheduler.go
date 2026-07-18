package service

import (
	"context"
	"log/slog"
	"time"
)

// ScheduledTask is a periodic background job.
type ScheduledTask struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

// Scheduler runs periodic background tasks. Register tasks before calling Run.
type Scheduler struct {
	tasks []ScheduledTask
	log   *slog.Logger
}

// NewScheduler creates a new Scheduler with the given logger.
func NewScheduler(log *slog.Logger) *Scheduler {
	if log == nil {
		log = slog.Default()
	}
	return &Scheduler{log: log}
}

// Register adds a task to the scheduler. Must be called before Run.
func (s *Scheduler) Register(task ScheduledTask) {
	s.tasks = append(s.tasks, task)
}

// Run starts all registered tasks in individual goroutines. It blocks until ctx
// is cancelled; each task goroutine also stops on ctx cancellation.
func (s *Scheduler) Run(ctx context.Context) {
	for _, task := range s.tasks {
		go s.runTask(ctx, task)
	}
	<-ctx.Done()
}

func (s *Scheduler) runTask(ctx context.Context, task ScheduledTask) {
	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()
	s.log.Info("scheduler: started task", "name", task.Name, "interval", task.Interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := task.Fn(ctx); err != nil {
				s.log.Error("scheduler: task failed", "name", task.Name, "err", err)
			}
		}
	}
}

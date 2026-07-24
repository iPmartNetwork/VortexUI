package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ScheduledTaskType is the kind of scheduled operation.
type ScheduledTaskType string

const (
	TaskTypeCleanup     ScheduledTaskType = "cleanup"
	TaskTypeBackup      ScheduledTaskType = "backup"
	TaskTypeSNIRotation ScheduledTaskType = "sni_rotation"
	TaskTypeCustom      ScheduledTaskType = "custom"
)

// ScheduledTask represents a task scheduled to run on a cron expression.
type ScheduledTask struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     ScheduledTaskType `json:"type"`
	Cron     string            `json:"cron"` // simplified cron: "minute hour day month weekday"
	Enabled  bool              `json:"enabled"`
	LastRun  *time.Time        `json:"last_run,omitempty"`
	NextRun  *time.Time        `json:"next_run,omitempty"`
	Handler  func(ctx context.Context) error `json:"-"`
}

// SchedulerService manages cron-scheduled tasks.
type SchedulerService struct {
	mu     sync.RWMutex
	tasks  map[string]*ScheduledTask
	stop   chan struct{}
	logger *slog.Logger
}

// NewSchedulerService creates a new scheduler.
func NewSchedulerService(logger *slog.Logger) *SchedulerService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SchedulerService{
		tasks:  make(map[string]*ScheduledTask),
		stop:   make(chan struct{}),
		logger: logger,
	}
}

// Register adds a task to the scheduler.
func (s *SchedulerService) Register(task *ScheduledTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := NextRun(task.Cron, time.Now())
	task.NextRun = &next
	s.tasks[task.ID] = task
}

// Unregister removes a task.
func (s *SchedulerService) Unregister(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}

// ListTasks returns all registered tasks.
func (s *SchedulerService) ListTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// Start begins the scheduler loop. Call Stop() to terminate.
func (s *SchedulerService) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

// Stop halts the scheduler.
func (s *SchedulerService) Stop() {
	close(s.stop)
}

func (s *SchedulerService) tick(ctx context.Context, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if !task.Enabled || task.NextRun == nil {
			continue
		}
		if now.After(*task.NextRun) {
			s.logger.Info("executing scheduled task", "id", task.ID, "name", task.Name)

			go func(t *ScheduledTask) {
				if err := t.Handler(ctx); err != nil {
					s.logger.Error("scheduled task failed", "id", t.ID, "error", err)
				}
			}(task)

			nowCopy := now
			task.LastRun = &nowCopy
			next := NextRun(task.Cron, now)
			task.NextRun = &next
		}
	}
}

// NextRun calculates the next execution time from a simplified cron expression.
// Format: "minute hour day month weekday" where * means any.
// Examples: "0 3 * * *" = daily at 3:00, "*/30 * * * *" = every 30 minutes.
func NextRun(cron string, from time.Time) time.Time {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		// Invalid cron, default to 1 hour from now
		return from.Add(1 * time.Hour)
	}

	minute := parseCronField(parts[0], 0, 59)
	hour := parseCronField(parts[1], 0, 23)

	// Simple implementation: find next matching minute/hour
	candidate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	for i := 0; i < 48*60; i++ { // search up to 48 hours
		candidate = candidate.Add(1 * time.Minute)
		if candidate.Before(from) {
			continue
		}
		if matchesCronField(minute, candidate.Minute()) && matchesCronField(hour, candidate.Hour()) {
			return candidate
		}
	}

	// Fallback: 1 hour from now
	return from.Add(1 * time.Hour)
}

// cronField is either -1 (wildcard *), a specific value, or an interval.
type cronField struct {
	wildcard bool
	value    int
	interval int // for */N patterns
}

func parseCronField(s string, min, max int) cronField {
	if s == "*" {
		return cronField{wildcard: true}
	}
	if strings.HasPrefix(s, "*/") {
		n, err := strconv.Atoi(s[2:])
		if err != nil || n <= 0 {
			return cronField{wildcard: true}
		}
		return cronField{interval: n}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return cronField{wildcard: true}
	}
	if v < min || v > max {
		return cronField{wildcard: true}
	}
	return cronField{value: v}
}

func matchesCronField(f cronField, value int) bool {
	if f.wildcard {
		return true
	}
	if f.interval > 0 {
		return value%f.interval == 0
	}
	return f.value == value
}

// FormatNextRun returns a human-readable string for when a task will next run.
func FormatNextRun(t *time.Time) string {
	if t == nil {
		return "never"
	}
	d := time.Until(*t)
	if d < 0 {
		return "overdue"
	}
	if d < time.Minute {
		return fmt.Sprintf("in %ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("in %dm", int(d.Minutes()))
	}
	return fmt.Sprintf("in %dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

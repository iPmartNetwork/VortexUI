package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// JobStore persists background jobs (backed by PostgreSQL or Redis).
type JobStore interface {
	Enqueue(ctx context.Context, job *domain.BackgroundJob) error
	Dequeue(ctx context.Context) (*domain.BackgroundJob, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus, result map[string]any, errMsg string) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BackgroundJob, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.BackgroundJob, error)
}

// JobHandler processes a specific job type.
type JobHandler func(ctx context.Context, payload map[string]any) (map[string]any, error)

// JobQueue manages background job enqueueing, dequeueing, and processing.
type JobQueue struct {
	store    JobStore
	handlers map[string]JobHandler
	logger   *slog.Logger
}

// NewJobQueue creates a new job queue with the given store.
func NewJobQueue(store JobStore, logger *slog.Logger) *JobQueue {
	if logger == nil {
		logger = slog.Default()
	}
	return &JobQueue{
		store:    store,
		handlers: make(map[string]JobHandler),
		logger:   logger,
	}
}

// RegisterHandler registers a handler for a specific job type.
func (q *JobQueue) RegisterHandler(jobType string, handler JobHandler) {
	q.handlers[jobType] = handler
}

// Enqueue adds a new job to the queue.
func (q *JobQueue) Enqueue(ctx context.Context, jobType string, payload map[string]any) (*domain.BackgroundJob, error) {
	job := &domain.BackgroundJob{
		ID:         uuid.New(),
		JobType:    jobType,
		Payload:    payload,
		Status:     domain.JobStatusPending,
		MaxRetries: 3,
		RunAfter:   time.Now(),
		CreatedAt:  time.Now(),
	}
	if err := q.store.Enqueue(ctx, job); err != nil {
		return nil, fmt.Errorf("enqueue job: %w", err)
	}
	return job, nil
}

// EnqueueDelayed adds a job to the queue with a delay.
func (q *JobQueue) EnqueueDelayed(ctx context.Context, jobType string, payload map[string]any, delay time.Duration) (*domain.BackgroundJob, error) {
	job := &domain.BackgroundJob{
		ID:         uuid.New(),
		JobType:    jobType,
		Payload:    payload,
		Status:     domain.JobStatusPending,
		MaxRetries: 3,
		RunAfter:   time.Now().Add(delay),
		CreatedAt:  time.Now(),
	}
	if err := q.store.Enqueue(ctx, job); err != nil {
		return nil, fmt.Errorf("enqueue delayed job: %w", err)
	}
	return job, nil
}

// GetStatus retrieves the current status of a job.
func (q *JobQueue) GetStatus(ctx context.Context, id uuid.UUID) (*domain.BackgroundJob, error) {
	return q.store.GetByID(ctx, id)
}

// ListRecent returns recent jobs.
func (q *JobQueue) ListRecent(ctx context.Context, limit int) ([]*domain.BackgroundJob, error) {
	if limit <= 0 {
		limit = 50
	}
	return q.store.ListRecent(ctx, limit)
}

// ProcessOne dequeues and processes a single job. Returns false if no jobs available.
func (q *JobQueue) ProcessOne(ctx context.Context) (bool, error) {
	job, err := q.store.Dequeue(ctx)
	if err != nil {
		return false, err
	}
	if job == nil {
		return false, nil
	}

	handler, ok := q.handlers[job.JobType]
	if !ok {
		errMsg := fmt.Sprintf("no handler registered for job type: %s", job.JobType)
		_ = q.store.UpdateStatus(ctx, job.ID, domain.JobStatusFailed, nil, errMsg)
		return true, errors.New(errMsg)
	}

	q.logger.Info("processing job", "id", job.ID, "type", job.JobType, "attempt", job.Attempts+1)

	result, processErr := handler(ctx, job.Payload)
	if processErr != nil {
		if job.Attempts+1 >= job.MaxRetries {
			_ = q.store.UpdateStatus(ctx, job.ID, domain.JobStatusFailed, nil, processErr.Error())
		} else {
			// Re-enqueue with backoff
			job.Attempts++
			job.RunAfter = time.Now().Add(time.Duration(1<<job.Attempts) * time.Second)
			_ = q.store.Enqueue(ctx, job)
		}
		return true, processErr
	}

	_ = q.store.UpdateStatus(ctx, job.ID, domain.JobStatusCompleted, result, "")
	return true, nil
}

// RunWorker starts a worker loop that processes jobs until the context is cancelled.
func (q *JobQueue) RunWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			processed, err := q.ProcessOne(ctx)
			if err != nil {
				q.logger.Error("job processing error", "error", err)
			}
			if !processed {
				// No work available, sleep briefly.
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// MarshalPayload is a helper to convert any struct into a job payload map.
func MarshalPayload(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus is the lifecycle status of a background job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// BackgroundJob represents an async task queued for background processing.
type BackgroundJob struct {
	ID          uuid.UUID      `json:"id"`
	JobType     string         `json:"job_type"`
	Payload     map[string]any `json:"payload"`
	Status      JobStatus      `json:"status"`
	Result      map[string]any `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	Attempts    int            `json:"attempts"`
	MaxRetries  int            `json:"max_retries"`
	RunAfter    time.Time      `json:"run_after"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Common job types.
const (
	JobTypeBackup       = "backup"
	JobTypeBulkOp       = "bulk_operation"
	JobTypeCleanup      = "cleanup"
	JobTypeSNIRotation  = "sni_rotation"
	JobTypeNodeSync     = "node_sync"
	JobTypeCoreUpdate   = "core_update"
	JobTypeNotification = "notification"
)

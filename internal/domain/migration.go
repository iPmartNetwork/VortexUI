package domain

import (
	"time"

	"github.com/google/uuid"
)

// MigrationStatus tracks the state of a user migration event.
type MigrationStatus string

const (
	MigrationPending   MigrationStatus = "pending"
	MigrationCompleted MigrationStatus = "completed"
	MigrationFailed    MigrationStatus = "failed"
)

// MigrationEvent records a user being moved from one node to another due to
// health-based auto-migration.
type MigrationEvent struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	Username   string          `json:"username,omitempty"`
	FromNodeID uuid.UUID       `json:"from_node_id"`
	ToNodeID   uuid.UUID       `json:"to_node_id"`
	Reason     string          `json:"reason"` // "unhealthy" | "high_latency" | "packet_loss"
	Status     MigrationStatus `json:"status"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// MigrationPolicy configures when and how auto-migration triggers.
type MigrationPolicy struct {
	Enabled            bool    `json:"enabled"`
	HealthCheckInterval int    `json:"health_check_interval"` // seconds
	UnhealthyThreshold int    `json:"unhealthy_threshold"`   // consecutive failures before migration
	CPUThreshold       float64 `json:"cpu_threshold"`         // trigger if CPU > this %
	MemThreshold       float64 `json:"mem_threshold"`         // trigger if Mem > this %
	PacketLossMax      float64 `json:"packet_loss_max"`       // trigger if loss > this %
	MigrateBack        bool    `json:"migrate_back"`          // return users when original node recovers
}

// DefaultMigrationPolicy returns sensible defaults.
func DefaultMigrationPolicy() MigrationPolicy {
	return MigrationPolicy{
		Enabled:            false,
		HealthCheckInterval: 30,
		UnhealthyThreshold: 3,
		CPUThreshold:       90,
		MemThreshold:       90,
		PacketLossMax:      10,
		MigrateBack:        true,
	}
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// ConfigVersion stores a snapshot of an inbound's configuration at a point in time.
// Each change increments the version counter for the given inbound, enabling diff
// and rollback operations.
type ConfigVersion struct {
	ID         uuid.UUID      `json:"id"`
	InboundID  uuid.UUID      `json:"inbound_id"`
	Version    int            `json:"version"`
	ConfigData map[string]any `json:"config_data"`
	Comment    string         `json:"comment"`
	AdminID    *uuid.UUID     `json:"admin_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ConfigDiff represents the difference between two configuration versions.
type ConfigDiff struct {
	InboundID  uuid.UUID      `json:"inbound_id"`
	OldVersion int            `json:"old_version"`
	NewVersion int            `json:"new_version"`
	OldConfig  map[string]any `json:"old_config"`
	NewConfig  map[string]any `json:"new_config"`
	Changes    []ConfigChange `json:"changes"`
}

// ConfigChange is a single field-level change between two configurations.
type ConfigChange struct {
	Path     string `json:"path"`      // JSON path, e.g. "streamSettings.network"
	OldValue any    `json:"old_value"` // nil if added
	NewValue any    `json:"new_value"` // nil if removed
	Type     string `json:"type"`      // "added", "removed", "modified"
}

// ConfigValidationError describes a field-level validation failure.
type ConfigValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ConfigDefaults holds a set of valid default settings for a protocol/transport/security combination.
type ConfigDefaults struct {
	Protocol  Protocol       `json:"protocol"`
	Network   string         `json:"network"`
	Security  Security       `json:"security"`
	Config    map[string]any `json:"config"`
}

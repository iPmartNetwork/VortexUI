package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserTemplate is a reusable blueprint for creating users with predefined
// settings. Admins apply a template to quickly provision users with consistent
// data limits, expiry, protocol config, and group membership.
type UserTemplate struct {
	ID               uuid.UUID      `json:"id"`
	Name             string         `json:"name"`
	DataLimit        int64          `json:"data_limit"`
	ExpireDuration   *int64         `json:"expire_duration,omitempty"` // seconds; nil = never expires
	DeviceLimit      int            `json:"device_limit"`
	ResetStrategy    ResetStrategy  `json:"reset_strategy"`
	Note             string         `json:"note,omitempty"`
	ProtocolSettings map[string]any `json:"protocol_settings"`
	Groups           []string       `json:"groups"`
	AllowedAdmins    []uuid.UUID    `json:"allowed_admins,omitempty"` // nil = unrestricted
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

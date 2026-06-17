package domain

import (
	"time"

	"github.com/google/uuid"
)

// FingerprintAction determines what to do when a fingerprint matches.
type FingerprintAction string

const (
	FingerprintAllow FingerprintAction = "allow"
	FingerprintBlock FingerprintAction = "block"
	FingerprintLog   FingerprintAction = "log"
)

// FingerprintRule defines a client TLS fingerprint rule.
type FingerprintRule struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Fingerprint string            `json:"fingerprint"` // e.g. "chrome", "firefox", "ios", "curl", "go"
	JA3Hash     string            `json:"ja3_hash,omitempty"`
	Action      FingerprintAction `json:"action"`
	Priority    int               `json:"priority"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
}

// FingerprintPolicy stores the global fingerprint validation settings.
type FingerprintPolicy struct {
	Enabled       bool              `json:"enabled"`
	DefaultAction FingerprintAction `json:"default_action"` // action for unknown fingerprints
	LogUnknown    bool              `json:"log_unknown"`
}

// DefaultFingerprintPolicy returns sensible defaults.
func DefaultFingerprintPolicy() FingerprintPolicy {
	return FingerprintPolicy{
		Enabled:       false,
		DefaultAction: FingerprintAllow,
		LogUnknown:    true,
	}
}

// FingerprintEvent records a fingerprint detection.
type FingerprintEvent struct {
	ID          uuid.UUID         `json:"id"`
	ClientIP    string            `json:"client_ip"`
	Fingerprint string            `json:"fingerprint"`
	JA3Hash     string            `json:"ja3_hash"`
	UserAgent   string            `json:"user_agent"`
	Action      FingerprintAction `json:"action"`
	NodeID      *uuid.UUID        `json:"node_id,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

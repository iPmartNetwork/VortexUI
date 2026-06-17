package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProbingAction determines what happens when a probe is detected.
type ProbingAction string

const (
	ProbingBlock    ProbingAction = "block"    // drop connection
	ProbingHoneypot ProbingAction = "honeypot" // respond with misleading data
	ProbingLog      ProbingAction = "log"      // log only, don't block
)

// ProbingPolicy configures the active probing protection behavior.
type ProbingPolicy struct {
	Enabled         bool          `json:"enabled"`
	Action          ProbingAction `json:"action"`           // default action
	BlockDuration   int           `json:"block_duration"`   // seconds to block an IP after detection
	MaxProbePerMin  int           `json:"max_probe_per_min"` // threshold before flagging
	WhitelistedIPs  []string      `json:"whitelisted_ips"`
	HoneypotHTML    string        `json:"honeypot_html,omitempty"` // custom HTML for honeypot
	NotifyTelegram  bool          `json:"notify_telegram"`
}

// DefaultProbingPolicy returns sensible defaults.
func DefaultProbingPolicy() ProbingPolicy {
	return ProbingPolicy{
		Enabled:        false,
		Action:         ProbingBlock,
		BlockDuration:  3600,
		MaxProbePerMin: 5,
		NotifyTelegram: true,
	}
}

// ProbeEvent records a detected active probing attempt.
type ProbeEvent struct {
	ID          uuid.UUID     `json:"id"`
	SourceIP    string        `json:"source_ip"`
	Port        int           `json:"port"`
	Method      string        `json:"method"`      // "tls_probe" | "http_probe" | "replay"
	Fingerprint string        `json:"fingerprint"` // detected TLS fingerprint
	Action      ProbingAction `json:"action"`      // what was done
	NodeID      *uuid.UUID    `json:"node_id,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// BlockedIP represents an IP that has been blocked due to probing.
type BlockedIP struct {
	IP        string    `json:"ip"`
	Reason    string    `json:"reason"`
	BlockedAt time.Time `json:"blocked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

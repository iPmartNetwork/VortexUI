package domain

import (
	"time"

	"github.com/google/uuid"
)

// Validation constants for Protocol Groups and Port Hopping.
const (
	MinProbeInterval = 30    // minimum health-check interval in seconds
	MaxProbeInterval = 600   // maximum health-check interval in seconds
	MinRetries       = 1     // minimum retry count
	MaxRetries       = 10    // maximum retry count
	MaxPortRange     = 10000 // maximum ports in a range
)

// AllowedHopIntervals are the only valid hop_interval values (seconds).
// 0 means disabled; 10, 30, 60 are active hop intervals.
var AllowedHopIntervals = []int{0, 10, 30, 60}

// IsValidHopInterval reports whether v is an allowed hop interval.
func IsValidHopInterval(v int) bool {
	for _, allowed := range AllowedHopIntervals {
		if v == allowed {
			return true
		}
	}
	return false
}

// ProtocolGroup logically groups Inbounds on a single Node as switching
// candidates. Priority order determines fallback sequence. The system renders
// these as urltest/fallback groups in subscription configs so clients auto-switch.
type ProtocolGroup struct {
	ID     uuid.UUID `json:"id"`
	NodeID uuid.UUID `json:"node_id"`
	Name   string    `json:"name"`

	// InboundIDs ordered by priority (index 0 = highest priority).
	InboundIDs []uuid.UUID `json:"inbound_ids"`

	// Health probe settings for client-side urltest/fallback.
	ProbeURL      string `json:"probe_url"`
	ProbeInterval int    `json:"probe_interval"` // seconds, minimum 30
	ProbeTimeout  int    `json:"probe_timeout"`  // seconds
	MaxRetries    int    `json:"max_retries"`    // 1-10

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ISPProfile defines per-ISP protocol preference ordering for a ProtocolGroup.
// When a subscription is fetched by a client matching this ISP, the ProtocolGroup
// inbounds are reordered according to PreferredProtocols.
type ISPProfile struct {
	ID            uuid.UUID `json:"id"`
	GroupID       uuid.UUID `json:"group_id"`
	ISPIdentifier string    `json:"isp_identifier"` // e.g. "mci", "irancell"
	CountryCode   string    `json:"country_code"`   // ISO 3166-1 alpha-2

	// PreferredProtocols ordered list of protocol+transport combos.
	// Entries not present in the group's inbounds are skipped at render time.
	// Format: "protocol+transport" e.g. "vless+ws", "vmess+grpc", "hysteria2"
	PreferredProtocols []string `json:"preferred_protocols"`

	CreatedAt time.Time `json:"created_at"`
}

// SwitchEvent records a client-reported protocol switch within a ProtocolGroup.
type SwitchEvent struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	NodeID         uuid.UUID `json:"node_id"`
	GroupID        uuid.UUID `json:"group_id,omitempty"`
	SourceProtocol string    `json:"source_protocol"`
	TargetProtocol string    `json:"target_protocol"`
	ISP            string    `json:"isp,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// SwitchEventFilter parameterizes switch event queries.
type SwitchEventFilter struct {
	NodeID   *uuid.UUID
	UserID   *uuid.UUID
	GroupID  *uuid.UUID
	ISP      string
	FromTime time.Time
	ToTime   time.Time
	Limit    int
}

// SwitchSummary holds aggregated switch event data for a time window.
type SwitchSummary struct {
	TotalSwitches  int                `json:"total_switches"`
	ByProtocol     map[string]int     `json:"by_protocol"`      // target_protocol → count
	ByNode         map[string]int     `json:"by_node"`          // node_id → count
	ByISP          map[string]int     `json:"by_isp"`           // isp → count
	TopSwitchPairs []SwitchPairCount  `json:"top_switch_pairs"` // most common source→target
}

// SwitchPairCount holds a source→target pair with its occurrence count.
type SwitchPairCount struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Count  int    `json:"count"`
}

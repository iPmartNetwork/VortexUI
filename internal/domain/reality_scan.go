package domain

import (
	"time"

	"github.com/google/uuid"
)

// RealityScan stores the result of probing a particular SNI for REALITY usage.
type RealityScan struct {
	ID        uuid.UUID `json:"id"`
	NodeID    uuid.UUID `json:"node_id"`
	SNI       string    `json:"sni"`
	LatencyMS int       `json:"latency_ms"`
	Score     int       `json:"score"`     // 0-100; higher = better
	Valid     bool      `json:"valid"`     // TLS handshake succeeded
	ScannedAt time.Time `json:"scanned_at"`
}

// RealityScanRequest is the input for triggering a scan.
type RealityScanRequest struct {
	NodeID uuid.UUID `json:"node_id"`
	SNIs   []string  `json:"snis"` // list of SNIs to probe
	Port   int       `json:"port"` // default 443
}

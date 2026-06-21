package domain

import (
	"time"

	"github.com/google/uuid"
)

// CleanIPScan stores the result of probing a candidate CDN (e.g. Cloudflare) IP
// for latency and packet loss, scored so the best clean IPs sort first.
type CleanIPScan struct {
	ID        uuid.UUID `json:"id"`
	IP        string    `json:"ip"`
	LatencyMS int       `json:"latency_ms"`
	LossPct   int       `json:"loss_pct"` // 0-100; lower = better
	Score     int       `json:"score"`    // 0-100; higher = better
	Reachable bool      `json:"reachable"`
	ScannedAt time.Time `json:"scanned_at"`
}

// CleanIPScanRequest is the input for triggering a clean-IP scan.
type CleanIPScanRequest struct {
	IPs  []string `json:"ips"`  // candidate IPs to probe
	Port int      `json:"port"` // default 443
}

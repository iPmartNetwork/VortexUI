package domain

import (
	"time"

	"github.com/google/uuid"
)

// TraceSpan represents one segment of a distributed trace.
type TraceSpan struct {
	TraceID   string            `json:"trace_id"`
	SpanID    string            `json:"span_id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Operation string            `json:"operation"` // e.g. "dns_resolve", "tls_handshake", "proxy_connect"
	Service   string            `json:"service"`   // e.g. "panel", "node-de-1", "xray"
	Duration  int64             `json:"duration_ms"`
	Status    string            `json:"status"` // "ok", "error", "timeout"
	StartTime time.Time         `json:"start_time"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Trace is a complete end-to-end trace of a connection.
type Trace struct {
	ID        string      `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	NodeID    uuid.UUID   `json:"node_id"`
	Spans     []TraceSpan `json:"spans"`
	TotalMS   int64       `json:"total_ms"`
	Status    string      `json:"status"` // "success", "failed", "timeout"
	CreatedAt time.Time   `json:"created_at"`
}

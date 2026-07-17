package domain

import (
	"time"

	"github.com/google/uuid"
)

// TrafficDelta is an incremental usage report from a node. Nodes only ever send
// the bytes consumed *since their last report*, so the panel can sum deltas
// idempotently — restarting a node or the panel never double counts or loses
// traffic the way absolute-counter polling does.
type TrafficDelta struct {
	NodeID    uuid.UUID `json:"node_id"`
	UserID    uuid.UUID `json:"user_id"`
	InboundID uuid.UUID `json:"inbound_id,omitempty"`
	Up        int64     `json:"up"`   // bytes uploaded since last report
	Down      int64     `json:"down"` // bytes downloaded since last report
	Timestamp time.Time `json:"timestamp"`
}

// Total is the combined up+down bytes of the delta.
func (d TrafficDelta) Total() int64 { return d.Up + d.Down }

// TrafficPoint is a stored time-series sample (persisted to a TimescaleDB
// hypertable) powering accurate, low-overhead usage charts.
type TrafficPoint struct {
	Time   time.Time `json:"time"`
	UserID uuid.UUID `json:"user_id"`
	NodeID uuid.UUID `json:"node_id"`
	Up     int64     `json:"up"`
	Down   int64     `json:"down"`
}

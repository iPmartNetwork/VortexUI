// Package stats consumes incremental traffic deltas pushed by node agents and
// fans them out to (a) the users table aggregate counter and (b) the time-series
// store. Because nodes report deltas (not absolute counters), aggregation is a
// plain additive fold that is safe across panel/node restarts.
package stats

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// userKey identifies a per-user accumulator bucket between flushes.
type userKey struct{ UserID uuid.UUID }

// Aggregator batches incoming deltas and flushes them periodically to bound DB
// write amplification under heavy traffic.
type Aggregator struct {
	users          port.UserRepository
	traffic        port.TrafficRepository
	inboundTraffic port.InboundTrafficRepository // nil means per-inbound tracking disabled

	in       chan domain.TrafficDelta
	flushDur time.Duration
	maxBatch int

	// AfterFlush runs after each successful user-traffic flush (optional). Used to
	// cut access promptly when a user crosses their data limit.
	AfterFlush func(ctx context.Context)
}

// New builds an Aggregator. flushDur and maxBatch trade latency for write load.
// inboundTraffic may be nil to disable per-inbound accumulation.
func New(users port.UserRepository, traffic port.TrafficRepository, inboundTraffic port.InboundTrafficRepository) *Aggregator {
	return &Aggregator{
		users:          users,
		traffic:        traffic,
		inboundTraffic: inboundTraffic,
		in:             make(chan domain.TrafficDelta, 4096),
		flushDur:       5 * time.Second,
		maxBatch:       1000,
	}
}

// Ingest hands a delta to the aggregator. Non-blocking: if the buffer is full
// the delta is dropped from the *time-series* path only after the caller decides
// — here we block briefly to avoid losing billing data. Callers on the gRPC hot
// path should select with ctx.Done().
func (a *Aggregator) Ingest(d domain.TrafficDelta) { a.in <- d }

// Run is the single consumer goroutine. Running exactly one consumer keeps the
// additive fold lock-free and ordered. Cancel ctx to drain and stop.
func (a *Aggregator) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.flushDur)
	defer ticker.Stop()

	// pending accumulates per-user byte sums between flushes to collapse many
	// tiny deltas into one UPDATE per user.
	pending := make(map[userKey]int64)
	// pendingInbound accumulates per-inbound upload/download between flushes.
	pendingInbound := make(map[uuid.UUID][2]int64) // [0]=up, [1]=down
	var points []domain.TrafficPoint

	flush := func() {
		flushedUsers := false
		if len(pending) > 0 {
			// One round-trip for every user's accumulated delta.
			byUser := make(map[uuid.UUID]int64, len(pending))
			for k, total := range pending {
				byUser[k.UserID] = total
				delete(pending, k)
			}
			if err := a.users.AddUsedTrafficBatch(ctx, byUser); err != nil {
				slog.Default().Error("failed to update user traffic", "err", err)
			} else {
				flushedUsers = true
			}
		}
		if len(points) > 0 {
			if err := a.traffic.WriteBatch(ctx, points); err != nil {
				slog.Default().Error("failed to write traffic points", "err", err)
			}
			points = points[:0]
		}
		// Flush per-inbound accumulated traffic.
		if a.inboundTraffic != nil && len(pendingInbound) > 0 {
			for ibID, ud := range pendingInbound {
				_ = a.inboundTraffic.AddTraffic(ctx, ibID, ud[0], ud[1])
			}
			clear(pendingInbound)
		}
		if flushedUsers && a.AfterFlush != nil {
			a.AfterFlush(ctx)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return ctx.Err()
		case <-ticker.C:
			flush()
		case d := <-a.in:
			k := userKey{UserID: d.UserID}
			pending[k] += d.Total()
			points = append(points, domain.TrafficPoint{
				Time: d.Timestamp, UserID: d.UserID, NodeID: d.NodeID,
				Up: d.Up, Down: d.Down,
			})
			// Accumulate per-inbound traffic when InboundID is set.
			if d.InboundID != (uuid.UUID{}) {
				v := pendingInbound[d.InboundID]
				v[0] += d.Up
				v[1] += d.Down
				pendingInbound[d.InboundID] = v
			}
			if len(points) >= a.maxBatch {
				flush()
			}
		}
	}
}

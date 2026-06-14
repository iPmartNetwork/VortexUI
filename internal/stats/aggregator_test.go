package stats

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

type fakeUsers struct {
	mu      sync.Mutex
	batches []map[uuid.UUID]int64
}

func (f *fakeUsers) AddUsedTrafficBatch(_ context.Context, d map[uuid.UUID]int64) error {
	f.mu.Lock()
	cp := make(map[uuid.UUID]int64, len(d))
	for k, v := range d {
		cp[k] = v
	}
	f.batches = append(f.batches, cp)
	f.mu.Unlock()
	return nil
}

// total sums every flushed delta for a user across all batches.
func (f *fakeUsers) total(id uuid.UUID) int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	var t int64
	for _, b := range f.batches {
		t += b[id]
	}
	return t
}

// Unused port methods (the aggregator only calls AddUsedTrafficBatch).
func (f *fakeUsers) Create(context.Context, *domain.User) error                       { return nil }
func (f *fakeUsers) GetByID(context.Context, uuid.UUID) (*domain.User, error)         { return nil, nil }
func (f *fakeUsers) GetBySubToken(context.Context, string) (*domain.User, error)      { return nil, nil }
func (f *fakeUsers) Update(context.Context, *domain.User) error                       { return nil }
func (f *fakeUsers) Delete(context.Context, uuid.UUID) error                          { return nil }
func (f *fakeUsers) List(context.Context, port.UserFilter) ([]*domain.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUsers) AddUsedTraffic(context.Context, uuid.UUID, int64) error      { return nil }
func (f *fakeUsers) SetInbounds(context.Context, uuid.UUID, []uuid.UUID) error   { return nil }
func (f *fakeUsers) InboundsFor(context.Context, uuid.UUID) ([]domain.Inbound, error) {
	return nil, nil
}

type fakeTraffic struct {
	mu     sync.Mutex
	points int
}

func (f *fakeTraffic) WriteBatch(_ context.Context, p []domain.TrafficPoint) error {
	f.mu.Lock()
	f.points += len(p)
	f.mu.Unlock()
	return nil
}
func (f *fakeTraffic) UsageSeries(context.Context, uuid.UUID, port.SeriesQuery) ([]domain.TrafficPoint, error) {
	return nil, nil
}

func TestAggregatorFoldsDeltasPerUser(t *testing.T) {
	users := &fakeUsers{}
	traffic := &fakeTraffic{}
	agg := New(users, traffic)
	agg.flushDur = 20 * time.Millisecond // flush quickly for the test

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = agg.Run(ctx); close(done) }()

	uid := uuid.New()
	// Many small deltas for one user must collapse into a summed counter.
	for i := 0; i < 5; i++ {
		agg.Ingest(domain.TrafficDelta{UserID: uid, Up: 10, Down: 5, Timestamp: time.Now()})
	}

	// Wait for at least one flush.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && users.total(uid) < 75 {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done

	if got := users.total(uid); got != 75 {
		t.Errorf("summed delta = %d, want 75 (5 × (10+5))", got)
	}
	if traffic.points != 5 {
		t.Errorf("time-series points = %d, want 5", traffic.points)
	}
}

func TestAggregatorDrainsOnShutdown(t *testing.T) {
	users := &fakeUsers{}
	agg := New(users, &fakeTraffic{})
	agg.flushDur = time.Hour // ensure the only flush is the shutdown drain

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = agg.Run(ctx); close(done) }()

	uid := uuid.New()
	agg.Ingest(domain.TrafficDelta{UserID: uid, Up: 100, Timestamp: time.Now()})
	time.Sleep(20 * time.Millisecond) // let the consumer pick it up

	cancel() // triggers a final flush
	<-done

	if got := users.total(uid); got != 100 {
		t.Errorf("shutdown drain total = %d, want 100", got)
	}
}

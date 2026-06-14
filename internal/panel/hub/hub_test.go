package hub

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- fakes ---

type fakeConn struct {
	mu         sync.Mutex
	deltas     []domain.TrafficDelta
	healthSeq  []domain.NodeHealth
	healthIdx  int
	healthFn   func(call int) (domain.NodeHealth, error) // overrides healthSeq when set
	healthCall int
	consumeErr error
}

func (f *fakeConn) Sync(context.Context, *core.GeneratedConfig, domain.CoreType) error { return nil }
func (f *fakeConn) AddUser(context.Context, string, *domain.User) error                { return nil }
func (f *fakeConn) RemoveUser(context.Context, string, uuid.UUID) error                { return nil }
func (f *fakeConn) Close() error                                                       { return nil }

func (f *fakeConn) Health(context.Context) (domain.NodeHealth, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.healthFn != nil {
		call := f.healthCall
		f.healthCall++
		return f.healthFn(call)
	}
	if len(f.healthSeq) == 0 {
		return domain.NodeHealth{CoreRunning: true}, nil
	}
	h := f.healthSeq[f.healthIdx]
	if f.healthIdx < len(f.healthSeq)-1 {
		f.healthIdx++
	}
	return h, nil
}

func (f *fakeConn) ConsumeTraffic(ctx context.Context, ingest func(domain.TrafficDelta)) error {
	for _, d := range f.deltas {
		ingest(d)
	}
	if f.consumeErr != nil {
		return f.consumeErr
	}
	<-ctx.Done()
	return ctx.Err()
}

// nopNodeRepo satisfies port.NodeRepository; only UpdateHealth is exercised.
type nopNodeRepo struct{ mu sync.Mutex; healthCalls int }

func (r *nopNodeRepo) Create(context.Context, *domain.Node) error          { return nil }
func (r *nopNodeRepo) GetByID(context.Context, uuid.UUID) (*domain.Node, error) { return nil, nil }
func (r *nopNodeRepo) Update(context.Context, *domain.Node) error          { return nil }
func (r *nopNodeRepo) Delete(context.Context, uuid.UUID) error             { return nil }
func (r *nopNodeRepo) List(context.Context) ([]*domain.Node, error)        { return nil, nil }
func (r *nopNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	r.mu.Lock()
	r.healthCalls++
	r.mu.Unlock()
	return nil
}

var _ port.NodeRepository = (*nopNodeRepo)(nil)

// --- tests ---

func TestSelectFailoverTarget(t *testing.T) {
	failed := uuid.New()
	a := &domain.Node{ID: uuid.New(), Name: "a", Status: domain.NodeConnected, UsageRatio: 1, Health: domain.NodeHealth{CoreRunning: true, CPUPercent: 80}}
	b := &domain.Node{ID: uuid.New(), Name: "b", Status: domain.NodeConnected, UsageRatio: 2, Health: domain.NodeHealth{CoreRunning: true, CPUPercent: 50}}
	down := &domain.Node{ID: uuid.New(), Name: "down", Status: domain.NodeDisconnected, UsageRatio: 9}

	got, ok := selectFailoverTarget([]*domain.Node{a, b, down}, failed)
	if !ok || got.Name != "b" {
		t.Fatalf("want b (highest usage ratio, healthy), got %v ok=%v", got, ok)
	}

	// Exclude the highest and the unhealthy: only a remains.
	got, ok = selectFailoverTarget([]*domain.Node{a, b, down}, b.ID)
	if !ok || got.Name != "a" {
		t.Fatalf("want a after excluding b, got %v", got)
	}

	// No healthy candidates.
	if _, ok := selectFailoverTarget([]*domain.Node{down}, failed); ok {
		t.Fatal("expected no target when none healthy")
	}
}

func TestHub_DrainsTrafficIntoIngest(t *testing.T) {
	uid := uuid.New()
	conn := &fakeConn{deltas: []domain.TrafficDelta{
		{UserID: uid, Up: 10, Down: 20},
		{UserID: uid, Up: 5, Down: 5},
	}}

	got := make(chan domain.TrafficDelta, 4)
	h := New(Options{
		Dialer: func(*domain.Node) (NodeConn, error) { return conn, nil },
		Ingest: func(d domain.TrafficDelta) { got <- d },
	})
	defer h.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	node := &domain.Node{ID: uuid.New(), Name: "n1", Core: domain.CoreXray}
	if err := h.Register(ctx, node); err != nil {
		t.Fatalf("Register: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case d := <-got:
			if d.UserID != uid {
				t.Errorf("unexpected user %s", d.UserID)
			}
		case <-ctx.Done():
			t.Fatalf("timed out waiting for delta %d", i)
		}
	}
}

func TestHub_ResyncOnConnectAndReconnect(t *testing.T) {
	healthy := domain.NodeHealth{CoreRunning: true}
	conn := &fakeConn{healthFn: func(call int) (domain.NodeHealth, error) {
		if call == 1 {
			return domain.NodeHealth{}, errors.New("agent down") // drop -> reachable=false
		}
		return healthy, nil // calls 0 (connect) and 2+ (reconnect)
	}}

	connected := make(chan string, 4)
	h := New(Options{
		Dialer:         func(*domain.Node) (NodeConn, error) { return conn, nil },
		HealthInterval: 5 * time.Millisecond,
		OnConnect:      func(_ context.Context, n *domain.Node) { connected <- n.Name },
	})
	defer h.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.Register(ctx, &domain.Node{ID: uuid.New(), Name: "n1", Core: domain.CoreXray}); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Expect two resyncs: one on initial connect, one after the reconnect edge.
	for i := 0; i < 2; i++ {
		select {
		case name := <-connected:
			if name != "n1" {
				t.Errorf("resync for wrong node: %s", name)
			}
		case <-ctx.Done():
			t.Fatalf("expected resync #%d (connect/reconnect), timed out", i+1)
		}
	}
}

func TestHub_FailoverOnHealthTransition(t *testing.T) {
	conn := &fakeConn{healthSeq: []domain.NodeHealth{
		{CoreRunning: true},  // first poll: becomes healthy
		{CoreRunning: false}, // second poll: unhealthy -> failover
	}}
	repo := &nopNodeRepo{}

	failed := make(chan string, 1)
	h := New(Options{
		Dialer:         func(*domain.Node) (NodeConn, error) { return conn, nil },
		Nodes:          repo,
		HealthInterval: 5 * time.Millisecond,
		OnFailover: func(_ context.Context, f *domain.Node, _ *domain.Node) {
			select {
			case failed <- f.Name:
			default:
			}
		},
	})
	defer h.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.Register(ctx, &domain.Node{ID: uuid.New(), Name: "flaky", Core: domain.CoreXray}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	select {
	case name := <-failed:
		if name != "flaky" {
			t.Errorf("failover for wrong node: %s", name)
		}
	case <-ctx.Done():
		t.Fatal("expected failover to fire on health transition")
	}
}

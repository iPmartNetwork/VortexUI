package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// --- fakes for outbound/routing/balancer repos ---

type fakeOutboundRepo struct {
	items   []*domain.Outbound
	created *domain.Outbound
	deleted uuid.UUID
}

func (f *fakeOutboundRepo) Create(_ context.Context, o *domain.Outbound) error {
	f.created = o
	f.items = append(f.items, o)
	return nil
}
func (f *fakeOutboundRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Outbound, error) {
	for _, o := range f.items {
		if o.ID == id {
			return o, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (f *fakeOutboundRepo) Update(_ context.Context, o *domain.Outbound) error {
	f.created = o
	return nil
}
func (f *fakeOutboundRepo) Delete(_ context.Context, id uuid.UUID) error { f.deleted = id; return nil }
func (f *fakeOutboundRepo) ListByNode(context.Context, uuid.UUID) ([]*domain.Outbound, error) {
	return f.items, nil
}

type fakeRoutingRepo struct {
	items   []*domain.RoutingRule
	created *domain.RoutingRule
}

func (f *fakeRoutingRepo) Create(_ context.Context, r *domain.RoutingRule) error {
	f.created = r
	f.items = append(f.items, r)
	return nil
}
func (f *fakeRoutingRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.RoutingRule, error) {
	for _, r := range f.items {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (f *fakeRoutingRepo) Update(_ context.Context, r *domain.RoutingRule) error {
	f.created = r
	return nil
}
func (f *fakeRoutingRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (f *fakeRoutingRepo) ListByNode(context.Context, uuid.UUID) ([]*domain.RoutingRule, error) {
	return f.items, nil
}

type fakeBalancerRepo struct {
	items   []*domain.Balancer
	created *domain.Balancer
}

func (f *fakeBalancerRepo) Create(_ context.Context, b *domain.Balancer) error {
	f.created = b
	f.items = append(f.items, b)
	return nil
}
func (f *fakeBalancerRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Balancer, error) {
	for _, b := range f.items {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (f *fakeBalancerRepo) Update(_ context.Context, b *domain.Balancer) error {
	f.created = b
	return nil
}
func (f *fakeBalancerRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (f *fakeBalancerRepo) ListByNode(context.Context, uuid.UUID) ([]*domain.Balancer, error) {
	return f.items, nil
}

// newSync builds a SyncService whose syncer captures the pushed config.
func newSyncWith(t *testing.T, outs *fakeOutboundRepo, routes *fakeRoutingRepo, bals *fakeBalancerRepo) (*SyncService, *fakeSyncer) {
	t.Helper()
	syncer := &fakeSyncer{}
	svc := NewSyncService(&fakeInboundLister{}, &fakeUsersByNoder{m: map[string][]*domain.User{}}, syncer, outs, routes, bals)
	return svc, syncer
}

// --- tests ---

func TestOutboundServiceCreateValidatesAndResyncs(t *testing.T) {
	repo := &fakeOutboundRepo{}
	sync, syncer := newSyncWith(t, repo, &fakeRoutingRepo{}, &fakeBalancerRepo{})
	svc := NewOutboundService(repo, sync)
	nodeID := uuid.New()

	o, err := svc.Create(context.Background(), CreateOutboundInput{
		NodeID: nodeID, Tag: "proxy-de", Protocol: domain.OutVLESS,
		Address: "de.example.com", Port: 443, UUID: "11111111-1111-1111-1111-111111111111",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if repo.created == nil || !o.Enabled {
		t.Error("outbound not persisted / not enabled by default")
	}
	// Resync must have pushed a config that includes the new outbound.
	if syncer.cfg == nil || len(syncer.cfg.Outbounds) != 1 || syncer.cfg.Outbounds[0].Tag != "proxy-de" {
		t.Errorf("resync did not include the outbound: %+v", syncer.cfg)
	}

	// A proxy outbound without an endpoint must be rejected (validation).
	if _, err := svc.Create(context.Background(), CreateOutboundInput{
		NodeID: nodeID, Tag: "bad", Protocol: domain.OutTrojan,
	}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("expected ErrInvalid for endpointless proxy outbound, got %v", err)
	}
}

func TestRoutingServiceCreateValidates(t *testing.T) {
	repo := &fakeRoutingRepo{}
	sync, syncer := newSyncWith(t, &fakeOutboundRepo{}, repo, &fakeBalancerRepo{})
	svc := NewRoutingService(repo, sync)
	nodeID := uuid.New()

	rule, err := svc.Create(context.Background(), nodeID, RoutingRuleInput{
		InboundTags: []string{"vless-ws"}, OutboundTag: "direct",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !rule.Enabled || repo.created == nil {
		t.Error("rule not persisted / not enabled by default")
	}
	if syncer.cfg == nil || len(syncer.cfg.Routing) != 1 {
		t.Errorf("resync did not include the rule: %+v", syncer.cfg)
	}

	// No target -> invalid.
	if _, err := svc.Create(context.Background(), nodeID, RoutingRuleInput{InboundTags: []string{"x"}}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("expected ErrInvalid for rule without target, got %v", err)
	}
	// Two targets -> invalid.
	if _, err := svc.Create(context.Background(), nodeID, RoutingRuleInput{InboundTags: []string{"x"}, OutboundTag: "a", BalancerTag: "b"}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("expected ErrInvalid for rule with two targets, got %v", err)
	}
	// No matcher -> invalid.
	if _, err := svc.Create(context.Background(), nodeID, RoutingRuleInput{OutboundTag: "direct"}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("expected ErrInvalid for matcherless rule, got %v", err)
	}
}

func TestBalancerServiceCreateDefaultsAndValidates(t *testing.T) {
	repo := &fakeBalancerRepo{}
	sync, syncer := newSyncWith(t, &fakeOutboundRepo{}, &fakeRoutingRepo{}, repo)
	svc := NewBalancerService(repo, sync)
	nodeID := uuid.New()

	b, err := svc.Create(context.Background(), nodeID, BalancerInput{
		Tag: "lb", Selectors: []string{"proxy-"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Strategy defaults to random.
	if b.Strategy != domain.BalancerRandom {
		t.Errorf("strategy = %q, want random default", b.Strategy)
	}
	if syncer.cfg == nil || len(syncer.cfg.Balancers) != 1 {
		t.Errorf("resync did not include the balancer: %+v", syncer.cfg)
	}

	// No selectors -> invalid.
	if _, err := svc.Create(context.Background(), nodeID, BalancerInput{Tag: "empty"}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("expected ErrInvalid for balancer without selectors, got %v", err)
	}
	// Unknown strategy -> invalid.
	if _, err := svc.Create(context.Background(), nodeID, BalancerInput{Tag: "x", Selectors: []string{"p"}, Strategy: "bogus"}); !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("expected ErrInvalid for unknown strategy, got %v", err)
	}
}

func TestSyncServiceAssemblesAllPolicySections(t *testing.T) {
	nodeID := uuid.New()
	outs := &fakeOutboundRepo{items: []*domain.Outbound{{ID: uuid.New(), NodeID: nodeID, Tag: "direct", Protocol: domain.OutFreedom, Enabled: true}}}
	routes := &fakeRoutingRepo{items: []*domain.RoutingRule{{ID: uuid.New(), NodeID: nodeID, InboundTags: []string{"in"}, OutboundTag: "direct", Enabled: true}}}
	bals := &fakeBalancerRepo{items: []*domain.Balancer{{ID: uuid.New(), NodeID: nodeID, Tag: "lb", Selectors: []string{"proxy-"}, Strategy: domain.BalancerRandom, Enabled: true}}}

	syncer := &fakeSyncer{}
	svc := NewSyncService(&fakeInboundLister{}, &fakeUsersByNoder{m: map[string][]*domain.User{}}, syncer, outs, routes, bals)
	if err := svc.Resync(context.Background(), nodeID); err != nil {
		t.Fatalf("resync: %v", err)
	}
	if len(syncer.cfg.Outbounds) != 1 || len(syncer.cfg.Routing) != 1 || len(syncer.cfg.Balancers) != 1 {
		t.Errorf("resync missing sections: outs=%d routes=%d bals=%d",
			len(syncer.cfg.Outbounds), len(syncer.cfg.Routing), len(syncer.cfg.Balancers))
	}
}

type fakeUserStatsRepo struct{ stats domain.UserStats }

func (f *fakeUserStatsRepo) Stats(context.Context) (domain.UserStats, error) {
	return f.stats, nil
}

func TestOverviewBuildAggregatesUsersAndNodeConnectivity(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	stale := now.Add(-30 * time.Minute)

	onID, offID := uuid.New(), uuid.New()
	nodes := &mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
		onID:  {ID: onID, Name: "on", Core: domain.CoreXray, LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true, Connections: 3}},
		offID: {ID: offID, Name: "off", Core: domain.CoreSingbox, LastSeen: &stale, Health: domain.NodeHealth{CoreRunning: true}},
	}}
	// mapNodeRepo.List returns nil; override via a list-capable fake below.
	stats := &fakeUserStatsRepo{stats: domain.UserStats{
		Total: 5, TotalUsed: 999,
		ByStatus: map[domain.UserStatus]int{domain.UserStatusActive: 4, domain.UserStatusDisabled: 1},
	}}

	svc := NewOverviewService(stats, &listNodeRepo{nodes: []*domain.Node{nodes.nodes[onID], nodes.nodes[offID]}})
	svc.now = func() time.Time { return now }

	ov, err := svc.Build(context.Background())
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if ov.Users.Total != 5 || ov.Users.TotalUsed != 999 || ov.Users.ByStatus[domain.UserStatusActive] != 4 {
		t.Errorf("user stats wrong: %+v", ov.Users)
	}
	if ov.Nodes.Total != 2 || ov.Nodes.Online != 1 {
		t.Errorf("node summary wrong: total=%d online=%d", ov.Nodes.Total, ov.Nodes.Online)
	}
}

// listNodeRepo returns a fixed node list (mapNodeRepo.List returns nil).
type listNodeRepo struct{ nodes []*domain.Node }

func (l *listNodeRepo) Create(context.Context, *domain.Node) error { return nil }
func (l *listNodeRepo) GetByID(context.Context, uuid.UUID) (*domain.Node, error) {
	return nil, domain.ErrNotFound
}
func (l *listNodeRepo) Update(context.Context, *domain.Node) error   { return nil }
func (l *listNodeRepo) Delete(context.Context, uuid.UUID) error      { return nil }
func (l *listNodeRepo) List(context.Context) ([]*domain.Node, error) { return l.nodes, nil }
func (l *listNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	return nil
}

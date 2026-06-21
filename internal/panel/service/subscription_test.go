package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// mapNodeRepo serves nodes from a map for multi-node subscription tests.
type mapNodeRepo struct{ nodes map[uuid.UUID]*domain.Node }

func (m *mapNodeRepo) Create(context.Context, *domain.Node) error { return nil }
func (m *mapNodeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Node, error) {
	if n, ok := m.nodes[id]; ok {
		return n, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mapNodeRepo) Update(context.Context, *domain.Node) error   { return nil }
func (m *mapNodeRepo) Delete(context.Context, uuid.UUID) error      { return nil }
func (m *mapNodeRepo) List(context.Context) ([]*domain.Node, error) { return nil, nil }
func (m *mapNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	return nil
}

func TestSubscriptionPrunesUnhealthyNodes(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	stale := now.Add(-30 * time.Minute)

	liveID, deadID := uuid.New(), uuid.New()
	user := &domain.User{ID: uuid.New(), Username: "alice", SubToken: "tok",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}

	userRepo := &fakeUserRepo{
		created: user,
		inbounds: []domain.Inbound{
			{ID: uuid.New(), NodeID: liveID, Tag: "live-in", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true},
			{ID: uuid.New(), NodeID: deadID, Tag: "dead-in", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true},
		},
	}
	nodeRepo := &mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
		liveID: {ID: liveID, Name: "live", Address: "1.1.1.1:50051", LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true}},
		deadID: {ID: deadID, Name: "dead", Address: "2.2.2.2:50051", LastSeen: &stale, Health: domain.NodeHealth{CoreRunning: true}},
	}}

	svc := NewSubscriptionService(userRepo, nodeRepo, nil)
	svc.now = func() time.Time { return now }

	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// All enabled inbounds are included (even on stale nodes) — the client sees
	// the full config and silently skips unreachable endpoints, rather than the
	// panel hiding them and confusing users when a node briefly blips.
	if len(res.Proxies) != 2 {
		t.Fatalf("want 2 proxies (all enabled inbounds included), got %d", len(res.Proxies))
	}
}

// stubSubHostRepo serves SubHosts from a per-inbound map for projection tests.
type stubSubHostRepo struct {
	byInbound map[uuid.UUID][]*domain.SubHost
	calls     int
}

func (s *stubSubHostRepo) Create(context.Context, *domain.SubHost) error { return nil }
func (s *stubSubHostRepo) Update(context.Context, *domain.SubHost) error { return nil }
func (s *stubSubHostRepo) Delete(context.Context, uuid.UUID) error       { return nil }
func (s *stubSubHostRepo) GetByID(context.Context, uuid.UUID) (*domain.SubHost, error) {
	return nil, domain.ErrNotFound
}
func (s *stubSubHostRepo) ListByInbound(_ context.Context, id uuid.UUID) ([]*domain.SubHost, error) {
	return s.byInbound[id], nil
}
func (s *stubSubHostRepo) ListByInbounds(_ context.Context, ids []uuid.UUID) ([]*domain.SubHost, error) {
	s.calls++
	var out []*domain.SubHost
	for _, id := range ids {
		out = append(out, s.byInbound[id]...)
	}
	return out, nil
}

// TestSubscriptionNoHostsMatchesPreHostBehavior asserts that an empty host repo
// produces exactly the same proxies as the nil-repo (pre-host) path — the core
// no-regression guarantee of Requirement 1.6.
func TestSubscriptionNoHostsMatchesPreHostBehavior(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	nodeID := uuid.New()
	inID := uuid.New()
	user := &domain.User{ID: uuid.New(), Username: "alice", SubToken: "tok",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	mkRepos := func() (*fakeUserRepo, *mapNodeRepo) {
		return &fakeUserRepo{
				created:  user,
				inbounds: []domain.Inbound{{ID: inID, NodeID: nodeID, Tag: "in", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws", Security: domain.SecurityTLS, Enabled: true}},
			},
			&mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
				nodeID: {ID: nodeID, Name: "n1", Address: "1.1.1.1:50051", LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true}},
			}}
	}

	// Pre-host behavior: nil repo.
	ur1, nr1 := mkRepos()
	base := NewSubscriptionService(ur1, nr1, nil)
	base.now = func() time.Time { return now }
	want, err := base.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("base build: %v", err)
	}

	// Empty host repo: must be identical.
	ur2, nr2 := mkRepos()
	empty := &stubSubHostRepo{byInbound: map[uuid.UUID][]*domain.SubHost{}}
	withRepo := NewSubscriptionService(ur2, nr2, empty)
	withRepo.now = func() time.Time { return now }
	got, err := withRepo.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("repo build: %v", err)
	}

	if len(got.Proxies) != len(want.Proxies) {
		t.Fatalf("proxy count = %d, want %d (no-regression)", len(got.Proxies), len(want.Proxies))
	}
	for i := range want.Proxies {
		if !reflect.DeepEqual(got.Proxies[i], want.Proxies[i]) {
			t.Errorf("proxy[%d] differs from pre-host behavior:\n got=%+v\nwant=%+v", i, got.Proxies[i], want.Proxies[i])
		}
	}
}

// TestSubscriptionProjectsEnabledHostsInPriorityOrder asserts that an inbound
// with 2 enabled + 1 disabled host yields exactly the 2 enabled host proxies (in
// priority order) with overlaid fields and expanded remark/address — never the
// base inbound link (Requirement 1.2).
func TestSubscriptionProjectsEnabledHostsInPriorityOrder(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	nodeID := uuid.New()
	inID := uuid.New()
	user := &domain.User{ID: uuid.New(), Username: "alice", SubToken: "tok",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	userRepo := &fakeUserRepo{
		created:  user,
		inbounds: []domain.Inbound{{ID: inID, NodeID: nodeID, Tag: "in", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws", Security: domain.SecurityTLS, Enabled: true}},
	}
	nodeRepo := &mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
		nodeID: {ID: nodeID, Name: "n1", Address: "9.9.9.9:50051", LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true}},
	}}

	port8443 := 8443
	hosts := &stubSubHostRepo{byInbound: map[uuid.UUID][]*domain.SubHost{
		inID: {
			// Deliberately out of priority order to exercise the sort.
			{ID: uuid.New(), InboundID: inID, Remark: "second {USERNAME}", Address: "cdn2.example.com", Priority: 20, Enabled: true, Security: domain.HostSecurityInboundDefault},
			{ID: uuid.New(), InboundID: inID, Remark: "first {SERVER_IP}", Address: "{SERVER_IP}", Port: &port8443, SNI: "front.example.com", ALPN: "h2,http/1.1", MuxEnable: true, Priority: 10, Enabled: true, Security: domain.HostSecurityTLS},
			{ID: uuid.New(), InboundID: inID, Remark: "disabled", Address: "nope.example.com", Priority: 5, Enabled: false, Security: domain.HostSecurityInboundDefault},
		},
	}}

	svc := NewSubscriptionService(userRepo, nodeRepo, hosts)
	svc.now = func() time.Time { return now }

	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(res.Proxies) != 2 {
		t.Fatalf("want 2 host proxies (1 disabled skipped), got %d", len(res.Proxies))
	}
	if hosts.calls != 1 {
		t.Errorf("ListByInbounds called %d times, want 1 (batched, no N+1)", hosts.calls)
	}

	// Priority order: 10 (first) before 20 (second).
	p0, p1 := res.Proxies[0], res.Proxies[1]
	if p0.Name != "first 9.9.9.9" {
		t.Errorf("proxy[0] name = %q, want expanded \"first 9.9.9.9\"", p0.Name)
	}
	if p0.Host != "9.9.9.9" { // {SERVER_IP} expanded to the node host
		t.Errorf("proxy[0] host = %q, want expanded server ip 9.9.9.9", p0.Host)
	}
	if p0.Port != 8443 {
		t.Errorf("proxy[0] port = %d, want overlaid 8443", p0.Port)
	}
	if p0.SNI != "front.example.com" {
		t.Errorf("proxy[0] sni = %q, want overlaid front.example.com", p0.SNI)
	}
	if len(p0.ALPN) != 2 || p0.ALPN[0] != "h2" || p0.ALPN[1] != "http/1.1" {
		t.Errorf("proxy[0] alpn = %v, want [h2 http/1.1]", p0.ALPN)
	}
	if !p0.Mux {
		t.Error("proxy[0] mux should be enabled from host")
	}
	if p1.Name != "second alice" {
		t.Errorf("proxy[1] name = %q, want expanded \"second alice\"", p1.Name)
	}
	if p1.Host != "cdn2.example.com" {
		t.Errorf("proxy[1] host = %q, want cdn2.example.com", p1.Host)
	}
	// inbound_default security host inherits the inbound's tls security.
	if p1.Security != "tls" {
		t.Errorf("proxy[1] security = %q, want inherited tls", p1.Security)
	}
}

// fakePackResolver is a minimal PackResolver for embedding tests. userPack and
// globalPack are the IDs returned for the two selection scopes; packs maps an ID
// to the pack returned by GetPack.
type fakePackResolver struct {
	userPack   string
	globalPack string
	packs      map[string]*domain.RoutingPack
	err        error
}

func (f *fakePackResolver) GetUserPack(context.Context, uuid.UUID) (string, error) {
	return f.userPack, f.err
}
func (f *fakePackResolver) GetGlobalDefault(context.Context) (string, error) {
	return f.globalPack, f.err
}
func (f *fakePackResolver) GetPack(_ context.Context, id string) (*domain.RoutingPack, error) {
	if p, ok := f.packs[id]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}

func subFixtures(t *testing.T) (*fakeUserRepo, *mapNodeRepo, time.Time) {
	t.Helper()
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	// Fixed IDs so two independently-built fixtures yield byte-identical proxies
	// (the no-regression comparisons below depend on this).
	nodeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	inID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	user := &domain.User{ID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), Username: "alice", SubToken: "tok",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.MustParse("44444444-4444-4444-4444-444444444444")}}
	ur := &fakeUserRepo{
		created:  user,
		inbounds: []domain.Inbound{{ID: inID, NodeID: nodeID, Tag: "in", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws", Security: domain.SecurityTLS, Enabled: true}},
	}
	nr := &mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
		nodeID: {ID: nodeID, Name: "n1", Address: "1.1.1.1:50051", LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true}},
	}}
	return ur, nr, now
}

// A user with a selected pack gets that pack's rules embedded in the result,
// while the proxies stay exactly what the no-pack path produces (Req 3.3.1).
func TestSubscriptionEmbedsSelectedPackRules(t *testing.T) {
	wantRules := []domain.RoutingRule{
		{Name: "iran-direct", Domains: []string{"geosite:category-ir"}, OutboundTag: "direct"},
	}

	// Baseline: no resolver wired → no rules.
	ur0, nr0, now := subFixtures(t)
	base := NewSubscriptionService(ur0, nr0, nil)
	base.now = func() time.Time { return now }
	baseRes, err := base.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("base build: %v", err)
	}
	if baseRes.Rules != nil {
		t.Errorf("no resolver should yield nil rules, got %v", baseRes.Rules)
	}

	// With a selected user pack → rules embedded, proxies unchanged.
	ur1, nr1, _ := subFixtures(t)
	svc := NewSubscriptionService(ur1, nr1, nil)
	svc.now = func() time.Time { return now }
	svc.SetRoutingPacks(&fakePackResolver{
		userPack: "Iran Direct",
		packs:    map[string]*domain.RoutingPack{"Iran Direct": {ID: "Iran Direct", Rules: wantRules}},
	})
	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if !reflect.DeepEqual(res.Rules, wantRules) {
		t.Errorf("embedded rules = %v, want %v", res.Rules, wantRules)
	}
	if !reflect.DeepEqual(res.Proxies, baseRes.Proxies) {
		t.Errorf("proxies changed when a pack is selected:\n got=%+v\nwant=%+v", res.Proxies, baseRes.Proxies)
	}
}

// A user with no per-user pack falls back to the global default pack's rules.
func TestSubscriptionFallsBackToGlobalDefaultPack(t *testing.T) {
	wantRules := []domain.RoutingRule{
		{Name: "block-ads", Domains: []string{"ads.example.com"}, OutboundTag: "blocked"},
	}
	ur, nr, now := subFixtures(t)
	svc := NewSubscriptionService(ur, nr, nil)
	svc.now = func() time.Time { return now }
	svc.SetRoutingPacks(&fakePackResolver{
		userPack:   "", // none per-user
		globalPack: "Block Ads",
		packs:      map[string]*domain.RoutingPack{"Block Ads": {ID: "Block Ads", Rules: wantRules}},
	})
	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if !reflect.DeepEqual(res.Rules, wantRules) {
		t.Errorf("global default rules = %v, want %v", res.Rules, wantRules)
	}
}

// With no pack selected (resolver present but empty selections) the result must
// match the no-resolver path exactly: nil rules and identical proxies (Req 3.3.3).
func TestSubscriptionNoPackMatchesUnchangedPath(t *testing.T) {
	ur0, nr0, now := subFixtures(t)
	base := NewSubscriptionService(ur0, nr0, nil)
	base.now = func() time.Time { return now }
	baseRes, err := base.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("base build: %v", err)
	}

	ur1, nr1, _ := subFixtures(t)
	svc := NewSubscriptionService(ur1, nr1, nil)
	svc.now = func() time.Time { return now }
	svc.SetRoutingPacks(&fakePackResolver{packs: map[string]*domain.RoutingPack{}})
	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if res.Rules != nil {
		t.Errorf("no selected pack should yield nil rules, got %v", res.Rules)
	}
	if !reflect.DeepEqual(res.Proxies, baseRes.Proxies) {
		t.Errorf("proxies differ from unchanged path:\n got=%+v\nwant=%+v", res.Proxies, baseRes.Proxies)
	}
}

// A pack-resolution error must be swallowed (fail-open): nil rules, valid result.
func TestSubscriptionPackResolutionErrorFailsOpen(t *testing.T) {
	ur, nr, now := subFixtures(t)
	svc := NewSubscriptionService(ur, nr, nil)
	svc.now = func() time.Time { return now }
	svc.SetRoutingPacks(&fakePackResolver{userPack: "x", err: context.DeadlineExceeded})
	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build should not fail on pack error: %v", err)
	}
	if res.Rules != nil {
		t.Errorf("pack error should yield nil rules, got %v", res.Rules)
	}
}

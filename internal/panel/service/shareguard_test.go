package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- fakes specific to ShareGuard / IP-limit enforcement ---

// sgUserRepo is a UserRepository tailored for ShareGuard tests: List returns a
// fixed active+device-limited user set, InboundsFor maps per user, and Update
// records the latest status so restore/disable transitions can be asserted.
type sgUserRepo struct {
	fakeUserRepo
	list     []*domain.User
	inbounds map[uuid.UUID][]domain.Inbound
	mu       sync.Mutex
	statuses map[uuid.UUID]domain.UserStatus
}

func (r *sgUserRepo) List(_ context.Context, f port.UserFilter) ([]*domain.User, int, error) {
	if f.Offset > 0 {
		return nil, len(r.list), nil
	}
	return r.list, len(r.list), nil
}

func (r *sgUserRepo) InboundsFor(_ context.Context, id uuid.UUID) ([]domain.Inbound, error) {
	return r.inbounds[id], nil
}

// UsesToLimit is unused by ShareGuard but required to satisfy EnforceRepo.
func (r *sgUserRepo) UsersToLimit(context.Context) ([]*domain.User, error) { return nil, nil }

func (r *sgUserRepo) Update(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.statuses == nil {
		r.statuses = map[uuid.UUID]domain.UserStatus{}
	}
	r.statuses[u.ID] = u.Status
	return nil
}

func (r *sgUserRepo) status(id uuid.UUID) domain.UserStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.statuses[id]
}

// sgOnlineQuerier reports a fixed set of online IPs per node, or an error to
// exercise the fail-open path.
type sgOnlineQuerier struct {
	ips map[uuid.UUID]map[string]int64 // nodeID -> ip:lastSeen
	err bool
}

func (q *sgOnlineQuerier) OnlineStats(context.Context, uuid.UUID) (map[string]int, error) {
	return nil, nil
}

func (q *sgOnlineQuerier) OnlineIPs(_ context.Context, nodeID uuid.UUID, _ string) (map[string]int64, error) {
	if q.err {
		return nil, errors.New("tracker down")
	}
	return q.ips[nodeID], nil
}

// fakeIPLimitStore captures policy reads and inserted events.
type fakeIPLimitStore struct {
	pol      *domain.IPLimitPolicy
	polErr   bool
	mu       sync.Mutex
	inserted []*domain.IPLimitEvent
}

func (s *fakeIPLimitStore) GetPolicy(context.Context) (*domain.IPLimitPolicy, error) {
	if s.polErr {
		return nil, errors.New("policy read failed")
	}
	return s.pol, nil
}

func (s *fakeIPLimitStore) InsertEvent(_ context.Context, e *domain.IPLimitEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inserted = append(s.inserted, e)
	return nil
}

func (s *fakeIPLimitStore) events() []*domain.IPLimitEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*domain.IPLimitEvent(nil), s.inserted...)
}

// sgCoreLookup resolves node cores for the kill-connections branch.
type sgCoreLookup struct{ cores map[uuid.UUID]domain.CoreType }

func (l *sgCoreLookup) GetByID(_ context.Context, id uuid.UUID) (*domain.Node, error) {
	return &domain.Node{ID: id, Core: l.cores[id]}, nil
}

// shareGuardHarness builds a ShareGuard whose UserService is backed by the
// supplied repo + online querier, with the restore timer made synchronous and
// captured so tests can trigger the restore deterministically.
func shareGuardHarness(t *testing.T, repo *sgUserRepo, online *sgOnlineQuerier, store *fakeIPLimitStore, cores *sgCoreLookup, ops *fakeNodeOps) (*ShareGuard, func()) {
	t.Helper()
	us := NewUserService(repo, ops)
	us.SetOnlineQuerier(online)
	g := NewShareGuard(us, repo, ops, time.Minute, nil)
	g.now = func() time.Time { return time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC) }

	var pending []func()
	g.afterFunc = func(_ time.Duration, f func()) *time.Timer {
		pending = append(pending, f)
		return nil
	}
	if store != nil {
		g.SetIPLimit(store, cores)
	}
	runRestores := func() {
		fns := pending
		pending = nil
		for _, f := range fns {
			f()
		}
	}
	return g, runRestores
}

// sharer builds a user that is online from more IPs than its device limit.
func sharer(deviceLimit int) *domain.User {
	return &domain.User{ID: uuid.New(), Username: "sharer", Status: domain.UserStatusActive, DeviceLimit: deviceLimit}
}

// ips builds an online-IP map for a node with n distinct IPs.
func ips(n int) map[string]int64 {
	m := map[string]int64{}
	for i := 0; i < n; i++ {
		m[uuid.NewString()] = int64(1000 + i)
	}
	return m
}

func TestShareGuardWarnEmitsEventOnly(t *testing.T) {
	u := sharer(1)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{ips: map[uuid.UUID]map[string]int64{nodeID: ips(3)}}
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: true, Action: domain.IPLimitActionWarn}}
	ops := &fakeNodeOps{}
	g, _ := shareGuardHarness(t, repo, online, store, &sgCoreLookup{}, ops)

	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if len(ops.removed) != 0 || len(ops.added) != 0 {
		t.Errorf("warn must not touch cores: removed=%v added=%v", ops.removed, ops.added)
	}
	if repo.status(u.ID) != "" {
		t.Errorf("warn must not change status, got %q", repo.status(u.ID))
	}
	evs := store.events()
	if len(evs) != 1 || evs[0].Action != string(domain.IPLimitActionWarn) {
		t.Fatalf("expected 1 warn event, got %+v", evs)
	}
	if evs[0].OnlineIPs != 3 || evs[0].Limit != 1 {
		t.Errorf("event detail wrong: %+v", evs[0])
	}
}

func TestShareGuardDisableTemporarilyAndRestore(t *testing.T) {
	u := sharer(1)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{ips: map[uuid.UUID]map[string]int64{nodeID: ips(4)}}
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: true, Action: domain.IPLimitActionDisable, RestoreAfter: 900}}
	ops := &fakeNodeOps{}
	g, runRestores := shareGuardHarness(t, repo, online, store, &sgCoreLookup{}, ops)

	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if repo.status(u.ID) != domain.UserStatusLimited {
		t.Errorf("status = %q, want limited", repo.status(u.ID))
	}
	if len(ops.removed) != 1 {
		t.Errorf("expected 1 deprovision, got %v", ops.removed)
	}
	if len(store.events()) != 1 || store.events()[0].Action != string(domain.IPLimitActionDisable) {
		t.Errorf("expected disable_temporarily event, got %+v", store.events())
	}

	// Trigger the scheduled restore: user goes active and is re-provisioned.
	runRestores()
	if repo.status(u.ID) != domain.UserStatusActive {
		t.Errorf("after restore status = %q, want active", repo.status(u.ID))
	}
	if len(ops.added) != 1 {
		t.Errorf("restore must re-provision, added=%v", ops.added)
	}
}

func TestShareGuardKillConnectionsXray(t *testing.T) {
	u := sharer(2)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{ips: map[uuid.UUID]map[string]int64{nodeID: ips(5)}}
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: true, Action: domain.IPLimitActionKill}}
	cores := &sgCoreLookup{cores: map[uuid.UUID]domain.CoreType{nodeID: domain.CoreXray}}
	ops := &fakeNodeOps{}
	g, _ := shareGuardHarness(t, repo, online, store, cores, ops)

	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	// xray kill = remove then immediately re-add (re-provision the valid account).
	if len(ops.removed) != 1 || len(ops.added) != 1 {
		t.Errorf("xray kill should remove+add once each: removed=%v added=%v", ops.removed, ops.added)
	}
	if repo.status(u.ID) != "" {
		t.Errorf("xray kill must not flip status, got %q", repo.status(u.ID))
	}
	if len(store.events()) != 1 || store.events()[0].Action != string(domain.IPLimitActionKill) {
		t.Errorf("expected kill_connections event, got %+v", store.events())
	}
}

func TestShareGuardKillConnectionsSingboxDegrades(t *testing.T) {
	u := sharer(1)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{ips: map[uuid.UUID]map[string]int64{nodeID: ips(3)}}
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: true, Action: domain.IPLimitActionKill, RestoreAfter: 600}}
	cores := &sgCoreLookup{cores: map[uuid.UUID]domain.CoreType{nodeID: domain.CoreSingbox}}
	ops := &fakeNodeOps{}
	g, _ := shareGuardHarness(t, repo, online, store, cores, ops)

	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	// sing-box has no live-kill: degrade to disable_temporarily.
	if repo.status(u.ID) != domain.UserStatusLimited {
		t.Errorf("sing-box kill should degrade to disable: status=%q", repo.status(u.ID))
	}
	if len(ops.removed) != 1 {
		t.Errorf("degrade should deprovision, removed=%v", ops.removed)
	}
	evs := store.events()
	if len(evs) != 1 || evs[0].Action != string(domain.IPLimitActionDisable) {
		t.Errorf("expected effective disable_temporarily event, got %+v", evs)
	}
}

func TestShareGuardFailsOpenOnTrackerError(t *testing.T) {
	u := sharer(1)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{err: true}
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: true, Action: domain.IPLimitActionKill}}
	ops := &fakeNodeOps{}
	g, _ := shareGuardHarness(t, repo, online, store, &sgCoreLookup{}, ops)

	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if len(ops.removed) != 0 || len(ops.added) != 0 || repo.status(u.ID) != "" {
		t.Errorf("tracker error must fail open (no action): removed=%v added=%v status=%q", ops.removed, ops.added, repo.status(u.ID))
	}
	if len(store.events()) != 0 {
		t.Errorf("tracker error must not record events, got %+v", store.events())
	}
}

func TestShareGuardEnforcementOffIsUnchanged(t *testing.T) {
	u := sharer(1)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{ips: map[uuid.UUID]map[string]int64{nodeID: ips(3)}}
	// Policy present but disabled: legacy behavior must hold and record nothing.
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: false, Action: domain.IPLimitActionKill}}
	ops := &fakeNodeOps{}
	g, _ := shareGuardHarness(t, repo, online, store, &sgCoreLookup{}, ops)

	pub := &capPublisher{}
	g.SetPublisher(pub)
	// auto-limit defaults to false (legacy): detection-only, no core changes.
	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if len(ops.removed) != 0 || len(ops.added) != 0 || repo.status(u.ID) != "" {
		t.Errorf("enforcement-off legacy path must take no action: removed=%v added=%v status=%q", ops.removed, ops.added, repo.status(u.ID))
	}
	if len(store.events()) != 0 {
		t.Errorf("enforcement-off must not write ip_limit_events, got %+v", store.events())
	}
	// Legacy detection still alerts via the event bus.
	if len(pub.events) != 1 {
		t.Errorf("expected 1 legacy alert event, got %d", len(pub.events))
	}
	if _, hasLimited := pub.events[0].Data["limited"]; !hasLimited {
		t.Errorf("legacy event payload changed; want legacy 'limited' key, got %+v", pub.events[0].Data)
	}
}

func TestShareGuardAlertCooldownDedups(t *testing.T) {
	u := sharer(1)
	nodeID := uuid.New()
	repo := &sgUserRepo{
		list:     []*domain.User{u},
		inbounds: map[uuid.UUID][]domain.Inbound{u.ID: {{NodeID: nodeID, Tag: "in"}}},
	}
	online := &sgOnlineQuerier{ips: map[uuid.UUID]map[string]int64{nodeID: ips(3)}}
	store := &fakeIPLimitStore{pol: &domain.IPLimitPolicy{Enabled: true, Action: domain.IPLimitActionWarn, AlertCooldown: 900}}
	ops := &fakeNodeOps{}
	g, _ := shareGuardHarness(t, repo, online, store, &sgCoreLookup{}, ops)

	// Two passes within the cooldown window (clock is fixed): only one event.
	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick 1: %v", err)
	}
	if err := g.Tick(context.Background()); err != nil {
		t.Fatalf("tick 2: %v", err)
	}
	if len(store.events()) != 1 {
		t.Errorf("AlertCooldown should dedup repeat violations, got %d events", len(store.events()))
	}
}

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// --- fakes for inbound enhancement tests ---

// fakeInboundRepo is a minimal in-memory repository for testing the InboundService.
type fakeInboundRepo struct {
	inbounds map[uuid.UUID]*domain.Inbound
	byNode   map[uuid.UUID][]*domain.Inbound
}

func newFakeInboundRepo() *fakeInboundRepo {
	return &fakeInboundRepo{
		inbounds: make(map[uuid.UUID]*domain.Inbound),
		byNode:   make(map[uuid.UUID][]*domain.Inbound),
	}
}

func (f *fakeInboundRepo) Create(_ context.Context, in *domain.Inbound) error {
	f.inbounds[in.ID] = in
	f.byNode[in.NodeID] = append(f.byNode[in.NodeID], in)
	return nil
}

func (f *fakeInboundRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Inbound, error) {
	in, ok := f.inbounds[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return in, nil
}

func (f *fakeInboundRepo) Update(_ context.Context, in *domain.Inbound) error {
	f.inbounds[in.ID] = in
	// Update in byNode slice too.
	list := f.byNode[in.NodeID]
	for i, existing := range list {
		if existing.ID == in.ID {
			list[i] = in
			break
		}
	}
	return nil
}

func (f *fakeInboundRepo) Delete(_ context.Context, id uuid.UUID) error {
	in, ok := f.inbounds[id]
	if !ok {
		return domain.ErrNotFound
	}
	delete(f.inbounds, id)
	list := f.byNode[in.NodeID]
	for i, existing := range list {
		if existing.ID == id {
			f.byNode[in.NodeID] = append(list[:i], list[i+1:]...)
			break
		}
	}
	return nil
}

func (f *fakeInboundRepo) ListByNode(_ context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error) {
	return f.byNode[nodeID], nil
}

func (f *fakeInboundRepo) ListFleet(context.Context) ([]domain.InboundListItem, error) {
	return nil, nil
}

func (f *fakeInboundRepo) ListByNodePort(context.Context, uuid.UUID, int, int) ([]*domain.Inbound, error) {
	return nil, nil
}

// fakeNodeRepo implements port.NodeRepository for inbound tests.
type fakeInboundNodeRepo struct {
	nodes map[uuid.UUID]*domain.Node
}

func (f *fakeInboundNodeRepo) Create(context.Context, *domain.Node) error               { return nil }
func (f *fakeInboundNodeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Node, error) {
	n, ok := f.nodes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return n, nil
}
func (f *fakeInboundNodeRepo) Update(context.Context, *domain.Node) error               { return nil }
func (f *fakeInboundNodeRepo) Delete(context.Context, uuid.UUID) error                  { return nil }
func (f *fakeInboundNodeRepo) List(context.Context) ([]*domain.Node, error)             { return nil, nil }
func (f *fakeInboundNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	return nil
}

// --- portsOverlap tests ---

func TestPortsOverlap(t *testing.T) {
	tests := []struct {
		name               string
		aStart, aEnd       int
		bStart, bEnd       int
		want               bool
	}{
		{"single same", 443, 0, 443, 0, true},
		{"single different", 443, 0, 8080, 0, false},
		{"range overlap left", 2000, 3000, 2500, 3500, true},
		{"range overlap right", 2500, 3500, 2000, 3000, true},
		{"range contains", 1000, 5000, 2000, 3000, true},
		{"range contained", 2000, 3000, 1000, 5000, true},
		{"range adjacent no overlap", 1000, 2000, 2001, 3000, false},
		{"range adjacent overlap at boundary", 1000, 2000, 2000, 3000, true},
		{"single in range", 2500, 0, 2000, 3000, true},
		{"single outside range", 4000, 0, 2000, 3000, false},
		{"degenerate range", 443, 443, 443, 0, true},
		{"both single both zero end", 80, 0, 80, 0, true},
		{"single vs range no overlap", 1999, 0, 2000, 3000, false},
		{"range touching at single point", 5000, 0, 4000, 5000, true},
		{"fully disjoint ranges", 100, 200, 300, 400, false},
		{"range with end equal to start", 3000, 3000, 3000, 3000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := portsOverlap(tt.aStart, tt.aEnd, tt.bStart, tt.bEnd)
			if got != tt.want {
				t.Errorf("portsOverlap(%d, %d, %d, %d) = %v, want %v",
					tt.aStart, tt.aEnd, tt.bStart, tt.bEnd, got, tt.want)
			}
		})
	}
}

// --- listenOverlap tests ---

func TestListenOverlap(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"both empty", "", "", true},
		{"one wildcard", "0.0.0.0", "10.0.0.1", true},
		{"both wildcard", "0.0.0.0", "0.0.0.0", true},
		{"same specific", "10.0.0.1", "10.0.0.1", true},
		{"different specific", "10.0.0.1", "10.0.0.2", false},
		{"empty and specific", "", "192.168.1.1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listenOverlap(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("listenOverlap(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// --- helper to build a minimal InboundService for tests ---

func newTestInboundService(t *testing.T) (*InboundService, *fakeInboundRepo, uuid.UUID) {
	t.Helper()
	nodeID := uuid.New()
	inboundRepo := newFakeInboundRepo()
	nodeRepo := &fakeInboundNodeRepo{
		nodes: map[uuid.UUID]*domain.Node{
			nodeID: {
				ID:           nodeID,
				Name:         "test-node",
				Core:         domain.CoreXray,
				EnabledCores: []domain.CoreType{domain.CoreXray},
			},
		},
	}
	// Build a real SyncService with a no-op syncer so Resync doesn't panic.
	syncSvc := NewSyncService(inboundRepo, &noopUsersByNoder{}, &noopCoreSyncer{}, nil, nil, nil)
	svc := NewInboundService(inboundRepo, nodeRepo, syncSvc)
	return svc, inboundRepo, nodeID
}

// noopCoreSyncer implements the Syncer interface for test purposes.
type noopCoreSyncer struct{}

func (noopCoreSyncer) Sync(context.Context, uuid.UUID, *core.GeneratedConfig) error { return nil }

// noopUsersByNoder implements UsersByNoder for tests (returns empty user maps).
type noopUsersByNoder struct{}

func (noopUsersByNoder) UsersByNode(context.Context, uuid.UUID) (map[string][]*domain.User, error) {
	return map[string][]*domain.User{}, nil
}

// --- Clone tests ---

func TestCloneCreatesNewInboundWithCopySuffix(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	// Seed an inbound.
	original := &domain.Inbound{
		ID:       uuid.New(),
		NodeID:   nodeID,
		Tag:      "vless-ws",
		Protocol: domain.ProtoVLESS,
		Port:     443,
		Network:  "ws",
		Security: domain.SecurityNone,
		Enabled:  true,
		Notes:    "production endpoint",
	}
	repo.inbounds[original.ID] = original
	repo.byNode[nodeID] = append(repo.byNode[nodeID], original)

	// Clone with a specific port.
	cloned, err := svc.Clone(ctx, original.ID, 8443)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if cloned.ID == original.ID {
		t.Error("cloned inbound has the same ID as original")
	}
	if cloned.Tag != "vless-ws-copy" {
		t.Errorf("tag = %q, want %q", cloned.Tag, "vless-ws-copy")
	}
	if cloned.Port != 8443 {
		t.Errorf("port = %d, want 8443", cloned.Port)
	}
	if cloned.Notes != "production endpoint" {
		t.Errorf("notes not preserved: %q", cloned.Notes)
	}
	if cloned.Protocol != domain.ProtoVLESS {
		t.Errorf("protocol = %q, want vless", cloned.Protocol)
	}
}

func TestCloneWithZeroPortAssignsRandom(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	original := &domain.Inbound{
		ID:       uuid.New(),
		NodeID:   nodeID,
		Tag:      "trojan-grpc",
		Protocol: domain.ProtoTrojan,
		Port:     2053,
		Network:  "grpc",
		Security: domain.SecurityNone,
		Enabled:  false, // disabled so no port conflict check fires
	}
	repo.inbounds[original.ID] = original
	repo.byNode[nodeID] = append(repo.byNode[nodeID], original)

	cloned, err := svc.Clone(ctx, original.ID, 0)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if cloned.Port < 10000 || cloned.Port >= 60000 {
		t.Errorf("random port = %d, want in [10000, 60000)", cloned.Port)
	}
}

// --- BulkAction tests ---

func TestBulkActionEnable(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	id1, id2 := uuid.New(), uuid.New()
	repo.inbounds[id1] = &domain.Inbound{ID: id1, NodeID: nodeID, Tag: "a", Port: 1000, Enabled: false, Network: "tcp", Security: domain.SecurityNone}
	repo.inbounds[id2] = &domain.Inbound{ID: id2, NodeID: nodeID, Tag: "b", Port: 2000, Enabled: false, Network: "tcp", Security: domain.SecurityNone}
	repo.byNode[nodeID] = []*domain.Inbound{repo.inbounds[id1], repo.inbounds[id2]}

	affected, err := svc.BulkAction(ctx, []uuid.UUID{id1, id2}, "enable")
	if err != nil {
		t.Fatalf("BulkAction: %v", err)
	}
	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}
	if !repo.inbounds[id1].Enabled || !repo.inbounds[id2].Enabled {
		t.Error("expected both inbounds to be enabled")
	}
}

func TestBulkActionDisable(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	id1 := uuid.New()
	repo.inbounds[id1] = &domain.Inbound{ID: id1, NodeID: nodeID, Tag: "a", Port: 1000, Enabled: true, Network: "tcp", Security: domain.SecurityNone}
	repo.byNode[nodeID] = []*domain.Inbound{repo.inbounds[id1]}

	affected, err := svc.BulkAction(ctx, []uuid.UUID{id1}, "disable")
	if err != nil {
		t.Fatalf("BulkAction: %v", err)
	}
	if affected != 1 {
		t.Errorf("affected = %d, want 1", affected)
	}
	if repo.inbounds[id1].Enabled {
		t.Error("expected inbound to be disabled")
	}
}

func TestBulkActionDelete(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	id1, id2 := uuid.New(), uuid.New()
	repo.inbounds[id1] = &domain.Inbound{ID: id1, NodeID: nodeID, Tag: "a", Port: 1000, Enabled: true, Network: "tcp", Security: domain.SecurityNone}
	repo.inbounds[id2] = &domain.Inbound{ID: id2, NodeID: nodeID, Tag: "b", Port: 2000, Enabled: true, Network: "tcp", Security: domain.SecurityNone}
	repo.byNode[nodeID] = []*domain.Inbound{repo.inbounds[id1], repo.inbounds[id2]}

	affected, err := svc.BulkAction(ctx, []uuid.UUID{id1, id2}, "delete")
	if err != nil {
		t.Fatalf("BulkAction: %v", err)
	}
	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}
	if _, ok := repo.inbounds[id1]; ok {
		t.Error("inbound id1 should have been deleted")
	}
	if _, ok := repo.inbounds[id2]; ok {
		t.Error("inbound id2 should have been deleted")
	}
}

func TestBulkActionInvalidAction(t *testing.T) {
	svc, _, _ := newTestInboundService(t)
	ctx := context.Background()

	_, err := svc.BulkAction(ctx, []uuid.UUID{uuid.New()}, "restart")
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

// --- CheckPort tests ---

func TestCheckPortAvailable(t *testing.T) {
	svc, _, nodeID := newTestInboundService(t)
	ctx := context.Background()

	available, conflictTag, err := svc.CheckPort(ctx, nodeID, 8080)
	if err != nil {
		t.Fatalf("CheckPort: %v", err)
	}
	if !available {
		t.Error("expected port to be available")
	}
	if conflictTag != "" {
		t.Errorf("conflict tag = %q, want empty", conflictTag)
	}
}

func TestCheckPortConflict(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	// Add an existing enabled inbound on port 443.
	existing := &domain.Inbound{
		ID:       uuid.New(),
		NodeID:   nodeID,
		Tag:      "vless-443",
		Port:     443,
		Enabled:  true,
		Network:  "tcp",
		Security: domain.SecurityNone,
	}
	repo.inbounds[existing.ID] = existing
	repo.byNode[nodeID] = append(repo.byNode[nodeID], existing)

	available, conflictTag, err := svc.CheckPort(ctx, nodeID, 443)
	if err != nil {
		t.Fatalf("CheckPort: %v", err)
	}
	if available {
		t.Error("expected port to be unavailable")
	}
	if conflictTag != "vless-443" {
		t.Errorf("conflict tag = %q, want %q", conflictTag, "vless-443")
	}
}

func TestCheckPortConflictWithRange(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	// Add an inbound with port range 2000-3000.
	existing := &domain.Inbound{
		ID:      uuid.New(),
		NodeID:  nodeID,
		Tag:     "hysteria-range",
		Port:    2000,
		PortEnd: 3000,
		Enabled: true,
		Network: "tcp",
		Security: domain.SecurityNone,
	}
	repo.inbounds[existing.ID] = existing
	repo.byNode[nodeID] = append(repo.byNode[nodeID], existing)

	// Port 2500 falls within the range, should conflict.
	available, conflictTag, err := svc.CheckPort(ctx, nodeID, 2500)
	if err != nil {
		t.Fatalf("CheckPort: %v", err)
	}
	if available {
		t.Error("expected port 2500 to conflict with range 2000-3000")
	}
	if conflictTag != "hysteria-range" {
		t.Errorf("conflict tag = %q, want %q", conflictTag, "hysteria-range")
	}

	// Port 3001 is outside the range, should be available.
	available, _, err = svc.CheckPort(ctx, nodeID, 3001)
	if err != nil {
		t.Fatalf("CheckPort: %v", err)
	}
	if !available {
		t.Error("expected port 3001 to be available (outside range 2000-3000)")
	}
}

// --- Notes/PortEnd pass-through in Create/Update ---

func TestCreatePreservesNotesAndPortEnd(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	inbound, err := svc.Create(ctx, CreateInboundInput{
		NodeID:   nodeID,
		Tag:      "with-notes",
		Protocol: domain.ProtoVLESS,
		Port:     9000,
		PortEnd:  9100,
		Network:  "tcp",
		Security: domain.SecurityNone,
		Notes:    "internal testing endpoint",
		Enabled:  false, // disabled to avoid port conflict checks requiring node syncer
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if inbound.Notes != "internal testing endpoint" {
		t.Errorf("notes = %q, want %q", inbound.Notes, "internal testing endpoint")
	}
	if inbound.PortEnd != 9100 {
		t.Errorf("port_end = %d, want 9100", inbound.PortEnd)
	}
	// Verify it's persisted in the repo.
	stored := repo.inbounds[inbound.ID]
	if stored.Notes != "internal testing endpoint" || stored.PortEnd != 9100 {
		t.Errorf("repo stored: notes=%q port_end=%d", stored.Notes, stored.PortEnd)
	}
}

func TestUpdatePreservesNotesAndPortEnd(t *testing.T) {
	svc, repo, nodeID := newTestInboundService(t)
	ctx := context.Background()

	// Seed an inbound.
	original := &domain.Inbound{
		ID:       uuid.New(),
		NodeID:   nodeID,
		Tag:      "update-test",
		Protocol: domain.ProtoTrojan,
		Port:     7000,
		Network:  "tcp",
		Security: domain.SecurityNone,
		Enabled:  false,
	}
	repo.inbounds[original.ID] = original
	repo.byNode[nodeID] = append(repo.byNode[nodeID], original)

	updated, err := svc.Update(ctx, original.ID, UpdateInboundInput{
		Port:    7000,
		PortEnd: 7500,
		Network: "tcp",
		Notes:   "updated notes value",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Notes != "updated notes value" {
		t.Errorf("notes = %q, want %q", updated.Notes, "updated notes value")
	}
	if updated.PortEnd != 7500 {
		t.Errorf("port_end = %d, want 7500", updated.PortEnd)
	}
}

package main

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/hub"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

// fakeNodes is a minimal in-memory port.NodeRepository for wiring tests.
type fakeNodes struct {
	items   []*domain.Node
	created *domain.Node
	updated *domain.Node
}

func (f *fakeNodes) Create(_ context.Context, n *domain.Node) error {
	f.created = n
	f.items = append(f.items, n)
	return nil
}
func (f *fakeNodes) GetByID(context.Context, uuid.UUID) (*domain.Node, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeNodes) Update(_ context.Context, n *domain.Node) error { f.updated = n; return nil }
func (f *fakeNodes) Delete(context.Context, uuid.UUID) error        { return nil }
func (f *fakeNodes) List(context.Context) ([]*domain.Node, error)   { return f.items, nil }
func (f *fakeNodes) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	return nil
}

func TestEnsureLocalNodeCreatesThenReuses(t *testing.T) {
	repo := &fakeNodes{}
	cfg := &config.Panel{LocalNode: true, LocalNodeName: "local", LocalNodeHost: "1.2.3.4", Core: "xray"}
	ctx := context.Background()

	n1, err := ensureLocalNode(ctx, repo, cfg)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if repo.created == nil || n1.Name != "local" || n1.Address != "1.2.3.4" || n1.Core != domain.CoreXray {
		t.Fatalf("local node not created correctly: %+v", n1)
	}

	// Second call finds the existing node (no duplicate create).
	repo.created = nil
	n2, err := ensureLocalNode(ctx, repo, cfg)
	if err != nil {
		t.Fatalf("reuse: %v", err)
	}
	if repo.created != nil {
		t.Error("should not create a second local node")
	}
	if n2.ID != n1.ID {
		t.Errorf("local node id changed: %s -> %s", n1.ID, n2.ID)
	}
}

func TestEnsureLocalNodeUpdatesOnConfigChange(t *testing.T) {
	existing := &domain.Node{ID: uuid.New(), Name: "local", Address: "old", Core: domain.CoreXray, UsageRatio: 1}
	repo := &fakeNodes{items: []*domain.Node{existing}}
	cfg := &config.Panel{LocalNodeName: "local", LocalNodeHost: "new-host", Core: "singbox"}

	n, err := ensureLocalNode(context.Background(), repo, cfg)
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if repo.updated == nil || n.Address != "new-host" || n.Core != domain.CoreSingbox {
		t.Errorf("local node not updated on config change: %+v", n)
	}
}

func TestLocalAwareDialerRoutesLocalInProcess(t *testing.T) {
	localID := uuid.New()
	drv := &fakeLocalDriverWiring{}
	dial := localAwareDialer(vgrpc.TLSFiles{}, localID, drv, nil)

	conn, err := dial(&domain.Node{ID: localID, Name: "local"})
	if err != nil {
		t.Fatalf("dial local: %v", err)
	}
	if _, ok := conn.(*hub.LocalConn); !ok {
		t.Errorf("local node should get a *hub.LocalConn, got %T", conn)
	}
}

// fakeLocalDriverWiring is a no-op core.CoreDriver used only to prove the dialer
// routes the local node in-process (its methods are never invoked here).
type fakeLocalDriverWiring struct{}

func (fakeLocalDriverWiring) Type() domain.CoreType                               { return domain.CoreXray }
func (fakeLocalDriverWiring) Start(context.Context, *core.GeneratedConfig) error  { return nil }
func (fakeLocalDriverWiring) Stop(context.Context) error                          { return nil }
func (fakeLocalDriverWiring) Reload(context.Context, *core.GeneratedConfig) error { return nil }
func (fakeLocalDriverWiring) AddUser(context.Context, string, *domain.User) error { return nil }
func (fakeLocalDriverWiring) RemoveUser(context.Context, string, string) error    { return nil }
func (fakeLocalDriverWiring) StreamTraffic(context.Context) (<-chan domain.TrafficDelta, error) {
	return nil, nil
}
func (fakeLocalDriverWiring) Health(context.Context) (domain.NodeHealth, error) {
	return domain.NodeHealth{}, nil
}
func (fakeLocalDriverWiring) Version(context.Context) (string, error) { return "", nil }
func (fakeLocalDriverWiring) OnlineStats(context.Context) (map[string]int, error) {
	return nil, nil
}
func (fakeLocalDriverWiring) OnlineIPList(context.Context, string) (map[string]int64, error) {
	return nil, nil
}
func (fakeLocalDriverWiring) UpdateGeoAssets(context.Context, string, string) (int64, int64, error) {
	return 0, 0, nil
}
func (fakeLocalDriverWiring) Logs(context.Context, int) ([]string, error) { return nil, nil }

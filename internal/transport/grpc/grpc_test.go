package grpc

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	genv1 "github.com/vortexui/vortexui/internal/transport/genv1"
)

// fakeDriver records calls and emits a scripted set of traffic deltas, letting
// us exercise the full gRPC contract without a real proxy engine.
type fakeDriver struct {
	mu      sync.Mutex
	added   []string // "tag/userid"
	removed []string
	started bool
	deltas  []domain.TrafficDelta
	online  map[string]int
	logs    []string
}

func (f *fakeDriver) Type() domain.CoreType { return domain.CoreXray }
func (f *fakeDriver) Start(context.Context, *core.GeneratedConfig) error {
	f.mu.Lock()
	f.started = true
	f.mu.Unlock()
	return nil
}
func (f *fakeDriver) Stop(context.Context) error                          { return nil }
func (f *fakeDriver) Reload(context.Context, *core.GeneratedConfig) error { return nil }
func (f *fakeDriver) AddUser(_ context.Context, tag string, u *domain.User) error {
	f.mu.Lock()
	f.added = append(f.added, tag+"/"+u.ID.String())
	f.mu.Unlock()
	return nil
}
func (f *fakeDriver) RemoveUser(_ context.Context, tag, userID string) error {
	f.mu.Lock()
	f.removed = append(f.removed, tag+"/"+userID)
	f.mu.Unlock()
	return nil
}
func (f *fakeDriver) StreamTraffic(ctx context.Context) (<-chan domain.TrafficDelta, error) {
	ch := make(chan domain.TrafficDelta)
	go func() {
		defer close(ch)
		for _, d := range f.deltas {
			select {
			case <-ctx.Done():
				return
			case ch <- d:
			}
		}
	}()
	return ch, nil
}
func (f *fakeDriver) Health(context.Context) (domain.NodeHealth, error) {
	return domain.NodeHealth{CoreRunning: true, Connections: 7, CPUPercent: 12.5}, nil
}
func (f *fakeDriver) Version(context.Context) (string, error) { return "xray-1.8.0", nil }
func (f *fakeDriver) OnlineStats(context.Context) (map[string]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.online, nil
}
func (f *fakeDriver) OnlineIPList(context.Context, string) (map[string]int64, error) {
	return nil, nil
}
func (f *fakeDriver) Logs(context.Context, int) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.logs, nil
}

// newTestLink spins the NodeServer on an in-memory listener and returns a
// NodeClient wired to it over an insecure (test-only) transport.
func newTestLink(t *testing.T, drv core.CoreDriver, nodeID uuid.UUID) *NodeClient {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	genv1.RegisterNodeServiceServer(srv, NewNodeServer(drv, "test-agent"))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &NodeClient{nodeID: nodeID, conn: conn, rpc: genv1.NewNodeServiceClient(conn)}
}

func TestNodeContract_UnaryRPCs(t *testing.T) {
	drv := &fakeDriver{}
	c := newTestLink(t, drv, uuid.New())
	ctx := context.Background()

	u := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VMessUUID: uuid.New()}}

	if err := c.Sync(ctx, &core.GeneratedConfig{LogLevel: "warn"}, domain.CoreXray); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if err := c.AddUser(ctx, "in-1", u); err != nil {
		t.Fatalf("AddUser: %v", err)
	}
	if err := c.RemoveUser(ctx, "in-1", u.ID); err != nil {
		t.Fatalf("RemoveUser: %v", err)
	}

	h, err := c.Health(ctx)
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !h.CoreRunning || h.Connections != 7 {
		t.Fatalf("unexpected health: %+v", h)
	}

	drv.mu.Lock()
	defer drv.mu.Unlock()
	if !drv.started {
		t.Error("Sync did not start the driver")
	}
	if len(drv.added) != 1 || drv.added[0] != "in-1/"+u.ID.String() {
		t.Errorf("AddUser not recorded: %v", drv.added)
	}
	if len(drv.removed) != 1 {
		t.Errorf("RemoveUser not recorded: %v", drv.removed)
	}
}

func TestNodeContract_OnlineStatsAndLogs(t *testing.T) {
	uid := uuid.New()
	drv := &fakeDriver{
		online: map[string]int{uid.String(): 3},
		logs:   []string{"line-1", "line-2"},
	}
	c := newTestLink(t, drv, uuid.New())
	ctx := context.Background()

	online, err := c.OnlineStats(ctx)
	if err != nil {
		t.Fatalf("OnlineStats: %v", err)
	}
	if online[uid.String()] != 3 {
		t.Errorf("online = %v, want %s:3", online, uid)
	}

	lines, err := c.Logs(ctx, 100)
	if err != nil {
		t.Fatalf("Logs: %v", err)
	}
	if len(lines) != 2 || lines[0] != "line-1" || lines[1] != "line-2" {
		t.Errorf("logs = %v, want [line-1 line-2]", lines)
	}
}

func TestNodeContract_TrafficStreamIsDeltaAttributed(t *testing.T) {
	nodeID := uuid.New()
	uid := uuid.New()
	drv := &fakeDriver{deltas: []domain.TrafficDelta{
		{UserID: uid, Up: 100, Down: 200, Timestamp: time.Now()},
		{UserID: uid, Up: 50, Down: 25, Timestamp: time.Now()},
	}}
	c := newTestLink(t, drv, nodeID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var got []domain.TrafficDelta
	if err := c.ConsumeTraffic(ctx, func(d domain.TrafficDelta) { got = append(got, d) }); err != nil {
		t.Fatalf("ConsumeTraffic: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("want 2 deltas, got %d", len(got))
	}
	// The panel side must stamp every delta with the originating node id.
	for _, d := range got {
		if d.NodeID != nodeID {
			t.Errorf("delta not attributed to node: got %s want %s", d.NodeID, nodeID)
		}
		if d.UserID != uid {
			t.Errorf("wrong user: %s", d.UserID)
		}
	}
	if total := got[0].Total() + got[1].Total(); total != 375 {
		t.Errorf("summed deltas = %d, want 375", total)
	}
}

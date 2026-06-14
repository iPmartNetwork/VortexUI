package hub

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// fakeLocalDriver records calls and emits scripted deltas, standing in for a
// real engine so LocalConn is testable without a process.
type fakeLocalDriver struct {
	started bool
	added   []string
	removed []string
	stopped bool
	deltas  []domain.TrafficDelta
	online  map[string]int
	logs    []string
}

func (f *fakeLocalDriver) Type() domain.CoreType { return domain.CoreXray }
func (f *fakeLocalDriver) Start(context.Context, *core.GeneratedConfig) error {
	f.started = true
	return nil
}
func (f *fakeLocalDriver) Stop(context.Context) error                          { f.stopped = true; return nil }
func (f *fakeLocalDriver) Reload(context.Context, *core.GeneratedConfig) error { return nil }
func (f *fakeLocalDriver) AddUser(_ context.Context, tag string, u *domain.User) error {
	f.added = append(f.added, tag+"/"+u.ID.String())
	return nil
}
func (f *fakeLocalDriver) RemoveUser(_ context.Context, tag, userID string) error {
	f.removed = append(f.removed, tag+"/"+userID)
	return nil
}
func (f *fakeLocalDriver) StreamTraffic(ctx context.Context) (<-chan domain.TrafficDelta, error) {
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
func (f *fakeLocalDriver) Health(context.Context) (domain.NodeHealth, error) {
	return domain.NodeHealth{CoreRunning: f.started}, nil
}
func (f *fakeLocalDriver) Version(context.Context) (string, error) { return "fake", nil }
func (f *fakeLocalDriver) OnlineStats(context.Context) (map[string]int, error) {
	return f.online, nil
}
func (f *fakeLocalDriver) OnlineIPList(context.Context, string) (map[string]int64, error) {
	return nil, nil
}
func (f *fakeLocalDriver) Logs(context.Context, int) ([]string, error) { return f.logs, nil }

func TestLocalConnDelegatesToDriver(t *testing.T) {
	drv := &fakeLocalDriver{}
	nodeID := uuid.New()
	c := NewLocalConn(nodeID, drv)
	ctx := context.Background()

	if err := c.Sync(ctx, &core.GeneratedConfig{}, domain.CoreXray); err != nil || !drv.started {
		t.Fatalf("Sync should start the driver: started=%v err=%v", drv.started, err)
	}
	u := &domain.User{ID: uuid.New()}
	if err := c.AddUser(ctx, "in-1", u); err != nil || len(drv.added) != 1 {
		t.Fatalf("AddUser not delegated: %v err=%v", drv.added, err)
	}
	if err := c.RemoveUser(ctx, "in-1", u.ID); err != nil || len(drv.removed) != 1 {
		t.Fatalf("RemoveUser not delegated: %v err=%v", drv.removed, err)
	}
	if h, err := c.Health(ctx); err != nil || !h.CoreRunning {
		t.Fatalf("Health: %+v err=%v", h, err)
	}
}

func TestLocalConnStampsNodeIDOnTraffic(t *testing.T) {
	nodeID := uuid.New()
	uid := uuid.New()
	drv := &fakeLocalDriver{deltas: []domain.TrafficDelta{{UserID: uid, Up: 10, Down: 20}}}
	c := NewLocalConn(nodeID, drv)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got := make(chan domain.TrafficDelta, 1)
	go func() { _ = c.ConsumeTraffic(ctx, func(d domain.TrafficDelta) { got <- d }) }()

	select {
	case d := <-got:
		if d.NodeID != nodeID {
			t.Errorf("delta node id = %s, want %s", d.NodeID, nodeID)
		}
		if d.UserID != uid || d.Up != 10 || d.Down != 20 {
			t.Errorf("delta payload wrong: %+v", d)
		}
	case <-ctx.Done():
		t.Fatal("no delta received")
	}
}

func TestLocalConnCloseDoesNotStopDriver(t *testing.T) {
	drv := &fakeLocalDriver{}
	c := NewLocalConn(uuid.New(), drv)
	if err := c.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if drv.stopped {
		t.Error("Close must NOT stop the local driver (caller owns its lifecycle)")
	}
}

func TestLocalConnOnlineStatsAndLogs(t *testing.T) {
	uid := uuid.New()
	drv := &fakeLocalDriver{
		online: map[string]int{uid.String(): 2},
		logs:   []string{"x", "y"},
	}
	c := NewLocalConn(uuid.New(), drv)
	ctx := context.Background()

	online, err := c.OnlineStats(ctx)
	if err != nil || online[uid.String()] != 2 {
		t.Errorf("OnlineStats = %v err=%v", online, err)
	}
	lines, err := c.Logs(ctx, 0)
	if err != nil || len(lines) != 2 {
		t.Errorf("Logs = %v err=%v", lines, err)
	}
}

package xray

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// fakeAPI scripts QueryTraffic results and records mutations, standing in for a
// live Xray gRPC API so the driver's orchestration/delta logic is testable.
type fakeAPI struct {
	mu      sync.Mutex
	samples [][]UserTraffic
	call    int
	added   []string
	removed []string
	online  map[string]int
}

func (f *fakeAPI) AddUser(_ context.Context, in domain.Inbound, u *domain.User) error {
	f.mu.Lock()
	f.added = append(f.added, in.Tag+"/"+u.ID.String())
	f.mu.Unlock()
	return nil
}
func (f *fakeAPI) RemoveUser(_ context.Context, tag, email string) error {
	f.mu.Lock()
	f.removed = append(f.removed, tag+"/"+email)
	f.mu.Unlock()
	return nil
}
func (f *fakeAPI) QueryTraffic(context.Context, bool) ([]UserTraffic, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.call >= len(f.samples) {
		return nil, nil
	}
	s := f.samples[f.call]
	f.call++
	return s, nil
}
func (f *fakeAPI) Close() error { return nil }

func (f *fakeAPI) OnlineUsers(context.Context) (map[string]int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.online, nil
}

func (f *fakeAPI) OnlineIPs(context.Context, string) (map[string]int64, error) {
	return nil, nil
}

// newDriverWithAPI builds a Driver wired to a fake API without starting any
// process, by injecting the dialer and pre-connecting.
func newDriverWithAPI(t *testing.T, api xrayAPI) *Driver {
	t.Helper()
	d := New(Options{
		BinPath:       "unused",
		ConfigPath:    t.TempDir() + "/config.json",
		StatsInterval: 10 * time.Millisecond,
		APIDialer:     func(string) (xrayAPI, error) { return api, nil },
	})
	if err := d.connectAPI(); err != nil {
		t.Fatalf("connectAPI: %v", err)
	}
	return d
}

func TestDriver_StreamTrafficEmitsDeltasAndSkipsNoise(t *testing.T) {
	uid := uuid.New()
	api := &fakeAPI{samples: [][]UserTraffic{
		{
			{Email: uid.String(), Up: 100, Down: 200},               // real user
			{Email: uid.String(), Up: 0, Down: 0},                   // zero -> skipped
			{Email: "inbound>>>vless-ws>>>traffic", Up: 9, Down: 9}, // non-uuid -> skipped
		},
	}}
	d := newDriverWithAPI(t, api)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch, err := d.StreamTraffic(ctx)
	if err != nil {
		t.Fatalf("StreamTraffic: %v", err)
	}

	select {
	case got := <-ch:
		if got.UserID != uid {
			t.Errorf("user = %s, want %s", got.UserID, uid)
		}
		if got.Up != 100 || got.Down != 200 {
			t.Errorf("delta = %d/%d, want 100/200", got.Up, got.Down)
		}
	case <-ctx.Done():
		t.Fatal("expected a delta, timed out")
	}
}

func TestDriver_AddRemoveUserDelegatesToAPI(t *testing.T) {
	api := &fakeAPI{}
	d := newDriverWithAPI(t, api)
	// Seed the inbound cache as a real Start would, so AddUser can resolve it.
	d.inbounds = map[string]domain.Inbound{"vless-ws": {Tag: "vless-ws", Protocol: domain.ProtoVLESS}}
	ctx := context.Background()
	u := &domain.User{ID: uuid.New()}

	if err := d.AddUser(ctx, "vless-ws", u); err != nil {
		t.Fatalf("AddUser: %v", err)
	}
	if err := d.RemoveUser(ctx, "vless-ws", u.ID.String()); err != nil {
		t.Fatalf("RemoveUser: %v", err)
	}

	api.mu.Lock()
	defer api.mu.Unlock()
	if len(api.added) != 1 || api.added[0] != "vless-ws/"+u.ID.String() {
		t.Errorf("added = %v", api.added)
	}
	if len(api.removed) != 1 {
		t.Errorf("removed = %v", api.removed)
	}
}

func TestDriver_AddUserWithoutAPIErrors(t *testing.T) {
	// Default dialer is not wired, and we never connect: AddUser must fail clearly.
	d := New(Options{ConfigPath: t.TempDir() + "/c.json"})
	if err := d.AddUser(context.Background(), "t", &domain.User{ID: uuid.New()}); err == nil {
		t.Fatal("expected error when api not connected")
	}
}

func TestDriver_OnlineStatsDelegatesToAPI(t *testing.T) {
	uid := uuid.New()
	api := &fakeAPI{online: map[string]int{uid.String(): 5}}
	d := newDriverWithAPI(t, api)

	got, err := d.OnlineStats(context.Background())
	if err != nil {
		t.Fatalf("OnlineStats: %v", err)
	}
	if got[uid.String()] != 5 {
		t.Errorf("online = %v, want %s:5", got, uid)
	}
}

func TestDriver_OnlineStatsErrorsWithoutAPI(t *testing.T) {
	d := New(Options{BinPath: "x", ConfigPath: t.TempDir() + "/c.json"})
	if _, err := d.OnlineStats(context.Background()); err == nil {
		t.Error("expected error when API is not connected")
	}
}

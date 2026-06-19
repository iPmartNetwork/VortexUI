package singbox

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// fakeRunner records every applied config instead of spawning a process.
type fakeRunner struct {
	mu      sync.Mutex
	applied [][]byte
	running bool
}

func (f *fakeRunner) Apply(_ context.Context, raw []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(raw))
	copy(cp, raw)
	f.applied = append(f.applied, cp)
	f.running = true
	return nil
}
func (f *fakeRunner) Stop()             { f.mu.Lock(); f.running = false; f.mu.Unlock() }
func (f *fakeRunner) Running() bool     { f.mu.Lock(); defer f.mu.Unlock(); return f.running }
func (f *fakeRunner) Logs(int) []string { return nil }

func (f *fakeRunner) lastUserCount(t *testing.T) int {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.applied) == 0 {
		t.Fatal("no config applied")
	}
	var p parsedConfig
	if err := json.Unmarshal(f.applied[len(f.applied)-1], &p); err != nil {
		t.Fatalf("applied config invalid: %v", err)
	}
	return len(p.Experimental.V2RayAPI.Stats.Users)
}

type fakeStats struct {
	samples []UserTraffic
	called  int
}

func (f *fakeStats) QueryTraffic(context.Context, bool) ([]UserTraffic, error) {
	f.called++
	if f.called == 1 {
		return f.samples, nil
	}
	return nil, nil
}
func (f *fakeStats) Close() error { return nil }

func newTestDriver(t *testing.T, stats statsClient) (*Driver, *fakeRunner) {
	t.Helper()
	d := New(Options{
		ConfigPath:    t.TempDir() + "/sb.json",
		StatsInterval: 10 * time.Millisecond,
		StatsDialer:   func(string) (statsClient, error) { return stats, nil },
	})
	fr := &fakeRunner{}
	d.run = fr // inject (same package)
	return d, fr
}

func TestDriverRebuildsConfigOnMembershipChange(t *testing.T) {
	d, fr := newTestDriver(t, &fakeStats{})
	ctx := context.Background()

	u1 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	in := domain.Inbound{Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443}
	if err := d.Start(ctx, &core.GeneratedConfig{Inbounds: []domain.Inbound{in}, UsersByInbound: map[string][]*domain.User{"vless-ws": {u1}}}); err != nil {
		t.Fatalf("start: %v", err)
	}
	if n := fr.lastUserCount(t); n != 1 {
		t.Fatalf("after start want 1 user, got %d", n)
	}

	// Adding a user rebuilds + reapplies with both users present.
	u2 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	if err := d.AddUser(ctx, "vless-ws", u2); err != nil {
		t.Fatalf("add user: %v", err)
	}
	if n := fr.lastUserCount(t); n != 2 {
		t.Errorf("after add want 2 users, got %d", n)
	}

	// Removing one brings it back to a single-user config.
	if err := d.RemoveUser(ctx, "vless-ws", u1.ID.String()); err != nil {
		t.Fatalf("remove user: %v", err)
	}
	if n := fr.lastUserCount(t); n != 1 {
		t.Errorf("after remove want 1 user, got %d", n)
	}

	// Unknown inbound is rejected.
	if err := d.AddUser(ctx, "ghost", u2); err == nil {
		t.Error("expected error adding to unknown inbound")
	}
}

func TestDriverStartPropagatesWireGuardPeers(t *testing.T) {
	d, fr := newTestDriver(t, &fakeStats{})
	inID := uuid.New()
	in := domain.Inbound{
		ID: inID, Tag: "wg0", Protocol: domain.ProtoWireGuard, Port: 51820,
		Raw: map[string]any{"wireguard": map[string]any{
			"private_key": "SRV_PRIV",
			"public_key":  "SRV_PUB",
			"listen_port": 51820,
			"subnet":      "10.7.0.0/24",
		}},
	}
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{in},
		WireGuardPeers: map[string][]domain.WireGuardPeer{
			"wg0": {{InboundID: inID, UserID: uuid.New(), PublicKey: "PEER1_PUB", Address: "10.7.0.2"}},
		},
	}
	if err := d.Start(context.Background(), cfg); err != nil {
		t.Fatalf("start: %v", err)
	}

	// The driver must cache the peers from the full sync.
	if got := len(d.wgPeers["wg0"]); got != 1 {
		t.Fatalf("driver wgPeers want 1, got %d", got)
	}

	// And the applied (rendered) config must include the peer.
	fr.mu.Lock()
	defer fr.mu.Unlock()
	if len(fr.applied) == 0 {
		t.Fatal("no config applied")
	}
	var p struct {
		Endpoints []struct {
			Tag   string `json:"tag"`
			Peers []struct {
				PublicKey string `json:"public_key"`
			} `json:"peers"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(fr.applied[len(fr.applied)-1], &p); err != nil {
		t.Fatalf("applied config invalid: %v", err)
	}
	if len(p.Endpoints) != 1 || len(p.Endpoints[0].Peers) != 1 || p.Endpoints[0].Peers[0].PublicKey != "PEER1_PUB" {
		t.Fatalf("rendered config missing wireguard peer: %+v", p.Endpoints)
	}
}

func TestDriverStreamTrafficEmitsDeltas(t *testing.T) {
	uid := uuid.New()
	stats := &fakeStats{samples: []UserTraffic{{Email: uid.String(), Up: 7, Down: 8}}}
	d, _ := newTestDriver(t, stats)

	if err := d.Start(context.Background(), &core.GeneratedConfig{}); err != nil {
		t.Fatalf("start: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch, err := d.StreamTraffic(ctx)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	select {
	case d := <-ch:
		if d.UserID != uid || d.Up != 7 || d.Down != 8 {
			t.Errorf("delta wrong: %+v", d)
		}
	case <-ctx.Done():
		t.Fatal("no delta emitted")
	}
}

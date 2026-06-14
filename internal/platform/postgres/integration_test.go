package postgres

import (
	"context"
	_ "embed"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

//go:embed schema.sql
var schemaSQL string

// openTestStore connects to the database named by VORTEX_TEST_DB and resets the
// schema. Without that env var the integration tests skip, so `go test ./...`
// stays green on a machine with no database while CI (which sets it) runs them
// against a real Postgres/TimescaleDB.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := os.Getenv("VORTEX_TEST_DB")
	if dsn == "" {
		t.Skip("VORTEX_TEST_DB not set; skipping Postgres integration test")
	}
	ctx := context.Background()
	st, err := Open(ctx, dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if _, err := st.pool.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if _, err := st.pool.Exec(ctx, schemaSQL); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

func TestIntegration_UserLifecycleAndTraffic(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	nodes, inbounds, users, traffic := st.Nodes(), st.Inbounds(), st.Users(), st.Traffic()

	// Node + inbound the user will be bound to.
	node := &domain.Node{ID: uuid.New(), Name: "n1", Address: "1.2.3.4:50051", Core: domain.CoreXray, CreatedAt: time.Now()}
	if err := nodes.Create(ctx, node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	in := &domain.Inbound{ID: uuid.New(), NodeID: node.ID, Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws", Security: domain.SecurityTLS, SNI: []string{"a.com"}, Enabled: true}
	if err := inbounds.Create(ctx, in); err != nil {
		t.Fatalf("create inbound: %v", err)
	}

	// Create a user and bind it to the inbound (user-centric model).
	u := &domain.User{
		ID: uuid.New(), Username: "alice", Status: domain.UserStatusActive,
		DataLimit:     1 << 30,
		ResetStrategy: domain.ResetNone,
		Proxies:       domain.UserCredentials{VMessUUID: uuid.New(), VLESSUUID: uuid.New(), SSMethod: "aes-128-gcm"},
		SubToken: "tok-alice", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := users.SetInbounds(ctx, u.ID, []uuid.UUID{in.ID}); err != nil {
		t.Fatalf("set inbounds: %v", err)
	}

	// Atomic traffic accumulation, twice, must sum (no double-count, no lost write).
	if err := users.AddUsedTraffic(ctx, u.ID, 100); err != nil {
		t.Fatalf("add traffic 1: %v", err)
	}
	if err := users.AddUsedTraffic(ctx, u.ID, 250); err != nil {
		t.Fatalf("add traffic 2: %v", err)
	}

	got, err := users.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.UsedTraffic != 350 {
		t.Errorf("used_traffic = %d, want 350", got.UsedTraffic)
	}
	if got.Username != "alice" || got.Proxies.VLESSUUID != u.Proxies.VLESSUUID {
		t.Errorf("round-trip mismatch: %+v", got)
	}

	// The binding must be readable back.
	bound, err := users.InboundsFor(ctx, u.ID)
	if err != nil {
		t.Fatalf("inbounds for: %v", err)
	}
	if len(bound) != 1 || bound[0].Tag != "vless-ws" {
		t.Errorf("bindings = %+v, want [vless-ws]", bound)
	}

	// Time-series write + bucketed read.
	now := time.Now().UTC()
	if err := traffic.WriteBatch(ctx, []domain.TrafficPoint{
		{Time: now, UserID: u.ID, NodeID: node.ID, Up: 10, Down: 20},
		{Time: now.Add(time.Second), UserID: u.ID, NodeID: node.ID, Up: 5, Down: 5},
	}); err != nil {
		t.Fatalf("write traffic: %v", err)
	}
	series, err := traffic.UsageSeries(ctx, u.ID, port.SeriesQuery{
		FromUnix: now.Add(-time.Hour).Unix(), ToUnix: now.Add(time.Hour).Unix(), Bucket: "1h",
	})
	if err != nil {
		t.Fatalf("usage series: %v", err)
	}
	var up, down int64
	for _, p := range series {
		up += p.Up
		down += p.Down
	}
	if up != 15 || down != 25 {
		t.Errorf("series totals up/down = %d/%d, want 15/25", up, down)
	}

	// Not-found path normalizes to the domain sentinel.
	if _, err := users.GetByID(ctx, uuid.New()); err != domain.ErrNotFound {
		t.Errorf("missing user err = %v, want ErrNotFound", err)
	}
}

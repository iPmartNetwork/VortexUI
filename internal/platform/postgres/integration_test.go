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

func TestIntegration_OutboundRoutingBalancerRoundTrip(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	nodes := st.Nodes()
	outbounds, routing, balancers := st.Outbounds(), st.Routing(), st.Balancers()

	node := &domain.Node{ID: uuid.New(), Name: "n1", Address: "1.2.3.4:50051", Core: domain.CoreXray, CreatedAt: time.Now()}
	if err := nodes.Create(ctx, node); err != nil {
		t.Fatalf("create node: %v", err)
	}

	// Outbound round-trip with credentials + transport.
	o := &domain.Outbound{
		ID: uuid.New(), NodeID: node.ID, Tag: "proxy-de", Protocol: domain.OutVLESS,
		Address: "de.example.com", Port: 443, UUID: uuid.NewString(), Flow: "xtls-rprx-vision",
		Network: "tcp", Security: domain.SecurityTLS, SNI: "de.example.com",
		Raw: map[string]any{"note": "primary"}, Enabled: true,
	}
	if err := outbounds.Create(ctx, o); err != nil {
		t.Fatalf("create outbound: %v", err)
	}
	gotOut, err := outbounds.GetByID(ctx, o.ID)
	if err != nil {
		t.Fatalf("get outbound: %v", err)
	}
	if gotOut.Tag != "proxy-de" || gotOut.Protocol != domain.OutVLESS || gotOut.SNI != "de.example.com" || gotOut.Raw["note"] != "primary" {
		t.Errorf("outbound round-trip mismatch: %+v", gotOut)
	}

	// Routing rule round-trip with slice matchers.
	rule := &domain.RoutingRule{
		ID: uuid.New(), NodeID: node.ID, Priority: 5, Name: "ads",
		Domains: []string{"geosite:category-ads"}, Port: "80,443",
		OutboundTag: "blocked", Enabled: true,
	}
	if err := routing.Create(ctx, rule); err != nil {
		t.Fatalf("create rule: %v", err)
	}
	rules, err := routing.ListByNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	if len(rules) != 1 || rules[0].OutboundTag != "blocked" || len(rules[0].Domains) != 1 || rules[0].Port != "80,443" {
		t.Errorf("routing round-trip mismatch: %+v", rules)
	}

	// Balancer round-trip.
	b := &domain.Balancer{
		ID: uuid.New(), NodeID: node.ID, Tag: "lb", Selectors: []string{"proxy-"},
		Strategy: domain.BalancerLeastPing, Observe: true, ProbeInterval: "5s", Enabled: true,
	}
	if err := balancers.Create(ctx, b); err != nil {
		t.Fatalf("create balancer: %v", err)
	}
	gotBal, err := balancers.GetByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("get balancer: %v", err)
	}
	if gotBal.Tag != "lb" || gotBal.Strategy != domain.BalancerLeastPing || len(gotBal.Selectors) != 1 || !gotBal.Observe {
		t.Errorf("balancer round-trip mismatch: %+v", gotBal)
	}

	// Deleting the node cascades to its outbounds/routing/balancers.
	if err := nodes.Delete(ctx, node.ID); err != nil {
		t.Fatalf("delete node: %v", err)
	}
	if outs, _ := outbounds.ListByNode(ctx, node.ID); len(outs) != 0 {
		t.Errorf("outbounds not cascaded on node delete: %d", len(outs))
	}
	if rs, _ := routing.ListByNode(ctx, node.ID); len(rs) != 0 {
		t.Errorf("routing not cascaded on node delete: %d", len(rs))
	}
	if bs, _ := balancers.ListByNode(ctx, node.ID); len(bs) != 0 {
		t.Errorf("balancers not cascaded on node delete: %d", len(bs))
	}
}

func TestIntegration_BackupRestoreReplacesConfig(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()
	nodes, inbounds, users := st.Nodes(), st.Inbounds(), st.Users()

	// Pre-existing config that restore must wipe.
	oldNode := &domain.Node{ID: uuid.New(), Name: "old", Address: "9.9.9.9:50051", Core: domain.CoreXray, CreatedAt: time.Now()}
	if err := nodes.Create(ctx, oldNode); err != nil {
		t.Fatalf("create old node: %v", err)
	}
	oldIn := &domain.Inbound{ID: uuid.New(), NodeID: oldNode.ID, Tag: "old-in", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true}
	if err := inbounds.Create(ctx, oldIn); err != nil {
		t.Fatalf("create old inbound: %v", err)
	}

	// Snapshot describing a completely different fleet.
	newNodeID, newInID, newUID := uuid.New(), uuid.New(), uuid.New()
	snap := &domain.Backup{
		Version: domain.BackupVersion,
		Nodes:   []*domain.Node{{ID: newNodeID, Name: "new", Address: "1.1.1.1:50051", Core: domain.CoreSingbox, CreatedAt: time.Now()}},
		Inbounds: []*domain.Inbound{
			{ID: newInID, NodeID: newNodeID, Tag: "new-in", Protocol: domain.ProtoVLESS, Port: 8443, Network: "tcp", Security: domain.SecurityReality, Enabled: true},
		},
		Outbounds: []*domain.Outbound{{ID: uuid.New(), NodeID: newNodeID, Tag: "direct", Protocol: domain.OutFreedom, Enabled: true}},
		Routing:   []*domain.RoutingRule{{ID: uuid.New(), NodeID: newNodeID, InboundTags: []string{"new-in"}, OutboundTag: "direct", Enabled: true}},
		Balancers: []*domain.Balancer{{ID: uuid.New(), NodeID: newNodeID, Tag: "lb", Selectors: []string{"p"}, Strategy: domain.BalancerRandom, Enabled: true}},
		Users: []*domain.User{{
			ID: newUID, Username: "carol", Status: domain.UserStatusActive, ResetStrategy: domain.ResetNone,
			Proxies: domain.UserCredentials{VMessUUID: uuid.New(), VLESSUUID: uuid.New(), SSMethod: "aes-128-gcm"},
			SubToken: "tok-carol", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}},
		Bindings: []domain.UserProxy{{UserID: newUID, InboundID: newInID}},
	}

	if err := st.Backup().Restore(ctx, snap); err != nil {
		t.Fatalf("restore: %v", err)
	}

	// Old node (and its cascaded inbound) must be gone.
	if _, err := nodes.GetByID(ctx, oldNode.ID); err != domain.ErrNotFound {
		t.Errorf("old node should be wiped, got err=%v", err)
	}
	// New fleet must be present.
	list, err := nodes.List(ctx)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(list) != 1 || list[0].Name != "new" {
		t.Errorf("nodes after restore = %+v, want only 'new'", list)
	}
	carol, err := users.GetByID(ctx, newUID)
	if err != nil || carol.Username != "carol" {
		t.Fatalf("restored user missing: %v", err)
	}
	bound, err := users.InboundsFor(ctx, newUID)
	if err != nil || len(bound) != 1 || bound[0].Tag != "new-in" {
		t.Errorf("restored binding wrong: %+v err=%v", bound, err)
	}
}

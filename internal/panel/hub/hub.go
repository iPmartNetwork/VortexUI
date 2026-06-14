// Package hub is the panel's control plane over the fleet of nodes. It owns one
// long-lived connection per node, continuously drains each node's traffic stream
// into the stats aggregator, polls health for failover, and fans configuration
// and user mutations out to nodes. Everything below the NodeConn interface is
// transport detail, so the hub is fully testable with in-memory fakes.
package hub

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NodeConn is the hub's view of a single node link. *grpc.NodeClient satisfies
// it; tests supply a fake. Keeping this narrow lets the hub stay transport-agnostic.
type NodeConn interface {
	Sync(ctx context.Context, cfg *core.GeneratedConfig, coreType domain.CoreType) error
	AddUser(ctx context.Context, inboundTag string, u *domain.User) error
	RemoveUser(ctx context.Context, inboundTag string, userID uuid.UUID) error
	Health(ctx context.Context) (domain.NodeHealth, error)
	ConsumeTraffic(ctx context.Context, ingest func(domain.TrafficDelta)) error
	OnlineStats(ctx context.Context) (map[string]int, error)
	Logs(ctx context.Context, limit int) ([]string, error)
	Close() error
}

// Dialer establishes a connection to a node (mTLS gRPC in production).
type Dialer func(node *domain.Node) (NodeConn, error)

// FailoverFunc is invoked when a node is declared unhealthy, with a healthy
// target chosen by the hub (or nil if none is available). The higher layer (user
// service) performs the actual user migration; the hub only detects and routes.
type FailoverFunc func(ctx context.Context, failed *domain.Node, target *domain.Node)

// ReadyFunc is invoked when a node's agent becomes reachable (initial connect or
// after a reconnect). The higher layer uses it to push the node's full desired
// config, so a restarted node is repopulated automatically.
type ReadyFunc func(ctx context.Context, node *domain.Node)

// Options configures a Hub.
type Options struct {
	Dialer         Dialer
	Nodes          port.NodeRepository
	Ingest         func(domain.TrafficDelta) // usually stats.Aggregator.Ingest
	OnFailover     FailoverFunc
	OnConnect      ReadyFunc // called when a node (re)connects; usually a resync
	HealthInterval time.Duration
	Logger         *slog.Logger
}

// Hub manages the node fleet.
type Hub struct {
	opts  Options
	log   *slog.Logger
	mu    sync.RWMutex
	conns map[uuid.UUID]*managedNode

	failoverFn FailoverFunc // guarded by mu; settable post-construction
	onConnect  ReadyFunc    // guarded by mu; settable post-construction
}

// New builds a Hub, applying defaults.
func New(opts Options) *Hub {
	if opts.HealthInterval == 0 {
		opts.HealthInterval = 10 * time.Second
	}
	if opts.Ingest == nil {
		opts.Ingest = func(domain.TrafficDelta) {}
	}
	if opts.OnFailover == nil {
		opts.OnFailover = func(context.Context, *domain.Node, *domain.Node) {}
	}
	if opts.OnConnect == nil {
		opts.OnConnect = func(context.Context, *domain.Node) {}
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Hub{
		opts: opts, log: opts.Logger, conns: make(map[uuid.UUID]*managedNode),
		failoverFn: opts.OnFailover, onConnect: opts.OnConnect,
	}
}

// SetOnFailover replaces the failover handler after construction. This breaks the
// chicken-and-egg between the hub and the migration service (which needs the hub
// as its provisioner): build the hub, build the service with it, then wire this.
func (h *Hub) SetOnFailover(fn FailoverFunc) {
	if fn == nil {
		return
	}
	h.mu.Lock()
	h.failoverFn = fn
	h.mu.Unlock()
}

func (h *Hub) failover() FailoverFunc {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.failoverFn
}

// SetOnConnect replaces the (re)connect handler after construction, mirroring
// SetOnFailover (the resync handler needs the hub as its Syncer).
func (h *Hub) SetOnConnect(fn ReadyFunc) {
	if fn == nil {
		return
	}
	h.mu.Lock()
	h.onConnect = fn
	h.mu.Unlock()
}

func (h *Hub) connectHook() ReadyFunc {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.onConnect
}

// Register brings a node under management: it dials, then starts the traffic and
// health loops. Re-registering an already-managed node is a no-op.
func (h *Hub) Register(ctx context.Context, node *domain.Node) error {
	h.mu.Lock()
	if _, exists := h.conns[node.ID]; exists {
		h.mu.Unlock()
		return nil
	}
	mn := &managedNode{node: node, hub: h, status: domain.NodeDisconnected}
	h.conns[node.ID] = mn
	h.mu.Unlock()

	loopCtx, cancel := context.WithCancel(ctx)
	mn.cancel = cancel
	go mn.runTraffic(loopCtx)
	go mn.runHealth(loopCtx)
	return nil
}

// Deregister stops managing a node and closes its connection.
func (h *Hub) Deregister(nodeID uuid.UUID) {
	h.mu.Lock()
	mn := h.conns[nodeID]
	delete(h.conns, nodeID)
	h.mu.Unlock()
	if mn != nil {
		mn.stop()
	}
}

// Sync pushes desired state to one node.
func (h *Hub) Sync(ctx context.Context, nodeID uuid.UUID, cfg *core.GeneratedConfig) error {
	mn, err := h.managed(nodeID)
	if err != nil {
		return err
	}
	conn, err := mn.connection()
	if err != nil {
		return err
	}
	return conn.Sync(ctx, cfg, mn.node.Core)
}

// AddUser provisions a user on one node's inbound at runtime.
func (h *Hub) AddUser(ctx context.Context, nodeID uuid.UUID, inboundTag string, u *domain.User) error {
	mn, err := h.managed(nodeID)
	if err != nil {
		return err
	}
	conn, err := mn.connection()
	if err != nil {
		return err
	}
	return conn.AddUser(ctx, inboundTag, u)
}

// RemoveUser deprovisions a user from one node's inbound at runtime.
func (h *Hub) RemoveUser(ctx context.Context, nodeID uuid.UUID, inboundTag string, userID uuid.UUID) error {
	mn, err := h.managed(nodeID)
	if err != nil {
		return err
	}
	conn, err := mn.connection()
	if err != nil {
		return err
	}
	return conn.RemoveUser(ctx, inboundTag, userID)
}

// Status returns the live status/health snapshot of a managed node.
func (h *Hub) Status(nodeID uuid.UUID) (domain.NodeStatus, domain.NodeHealth, error) {
	mn, err := h.managed(nodeID)
	if err != nil {
		return "", domain.NodeHealth{}, err
	}
	mn.mu.Lock()
	defer mn.mu.Unlock()
	return mn.status, mn.health, nil
}

// OnlineStats returns one node's live per-user connection counts.
func (h *Hub) OnlineStats(ctx context.Context, nodeID uuid.UUID) (map[string]int, error) {
	mn, err := h.managed(nodeID)
	if err != nil {
		return nil, err
	}
	conn, err := mn.connection()
	if err != nil {
		return nil, err
	}
	return conn.OnlineStats(ctx)
}

// Logs returns up to limit recent core log lines from one node.
func (h *Hub) Logs(ctx context.Context, nodeID uuid.UUID, limit int) ([]string, error) {
	mn, err := h.managed(nodeID)
	if err != nil {
		return nil, err
	}
	conn, err := mn.connection()
	if err != nil {
		return nil, err
	}
	return conn.Logs(ctx, limit)
}

// HealthyNodes returns a snapshot of currently-healthy managed nodes, used for
// failover target selection.
func (h *Hub) HealthyNodes() []*domain.Node {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var out []*domain.Node
	for _, mn := range h.conns {
		mn.mu.Lock()
		healthy := mn.status == domain.NodeConnected && mn.health.CoreRunning
		n := *mn.node
		n.Health = mn.health
		n.Status = mn.status
		mn.mu.Unlock()
		if healthy {
			nn := n
			out = append(out, &nn)
		}
	}
	return out
}

// Close tears down all managed connections.
func (h *Hub) Close() {
	h.mu.Lock()
	conns := h.conns
	h.conns = make(map[uuid.UUID]*managedNode)
	h.mu.Unlock()
	for _, mn := range conns {
		mn.stop()
	}
}

func (h *Hub) managed(nodeID uuid.UUID) (*managedNode, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	mn, ok := h.conns[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %s not managed", nodeID)
	}
	return mn, nil
}

var errNotConnected = errors.New("node not connected")

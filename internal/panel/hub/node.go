package hub

import (
	"context"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// managedNode owns one node's connection and the two background loops that keep
// it healthy: a traffic drain and a health poll. Its mutex guards the live
// conn/status/health snapshot read by hub callers.
type managedNode struct {
	node *domain.Node
	hub  *Hub

	mu        sync.Mutex
	conn      NodeConn
	status    domain.NodeStatus
	health    domain.NodeHealth
	reachable bool // agent gRPC reachable; tracks the (re)connect edge

	cancel context.CancelFunc
}

func (m *managedNode) stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.mu.Lock()
	conn := m.conn
	m.conn = nil
	m.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

// connection returns the live connection or an error if the node is currently
// disconnected, so callers fail fast instead of nil-panicking.
func (m *managedNode) connection() (NodeConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.conn == nil {
		return nil, errNotConnected
	}
	return m.conn, nil
}

// ensureConn lazily (re)dials, caching the connection. Callers hold no lock.
func (m *managedNode) ensureConn() (NodeConn, error) {
	m.mu.Lock()
	if m.conn != nil {
		c := m.conn
		m.mu.Unlock()
		return c, nil
	}
	m.mu.Unlock()

	conn, err := m.hub.opts.Dialer(m.node)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.conn = conn
	m.mu.Unlock()
	return conn, nil
}

// dropConn discards the current connection so the next loop iteration redials.
func (m *managedNode) dropConn() {
	m.mu.Lock()
	conn := m.conn
	m.conn = nil
	m.status = domain.NodeDisconnected
	m.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

// runTraffic keeps the node's traffic stream draining into the aggregator,
// redialing with capped backoff whenever it drops, until ctx is cancelled.
func (m *managedNode) runTraffic(ctx context.Context) {
	bo := backoff{min: time.Second, max: 30 * time.Second}
	for {
		if ctx.Err() != nil {
			return
		}
		conn, err := m.ensureConn()
		if err != nil {
			m.hub.log.Warn("dial node failed", "node", m.node.Name, "err", err)
			if !bo.sleep(ctx) {
				return
			}
			continue
		}
		bo.reset()
		// Blocks until the stream ends or errors.
		err = conn.ConsumeTraffic(ctx, m.hub.opts.Ingest)
		if ctx.Err() != nil {
			return
		}
		m.hub.log.Warn("traffic stream ended, reconnecting", "node", m.node.Name, "err", err)
		m.dropConn()
		if !bo.sleep(ctx) {
			return
		}
	}
}

// runHealth polls node health on an interval, persists it, and triggers failover
// the moment a node transitions to unhealthy.
func (m *managedNode) runHealth(ctx context.Context) {
	ticker := time.NewTicker(m.hub.opts.HealthInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.pollOnce(ctx)
		}
	}
}

func (m *managedNode) pollOnce(ctx context.Context) {
	conn, err := m.ensureConn()
	if err != nil {
		m.markUnhealthy(ctx)
		return
	}
	h, err := conn.Health(ctx)
	if err != nil {
		m.hub.log.Warn("health check failed", "node", m.node.Name, "err", err)
		m.dropConn()
		m.markUnhealthy(ctx)
		return
	}

	m.mu.Lock()
	wasHealthy := m.status == domain.NodeConnected && m.health.CoreRunning
	wasReachable := m.reachable
	m.health = h
	m.status = domain.NodeConnected
	m.reachable = true
	nowHealthy := h.CoreRunning
	if !nowHealthy {
		m.status = domain.NodeError
	}
	m.mu.Unlock()

	if m.hub.opts.Nodes != nil {
		_ = m.hub.opts.Nodes.UpdateHealth(ctx, m.node.ID, h)
	}
	// Unreachable -> reachable edge: (re)push the node's full config so a freshly
	// started or recovered agent is repopulated without manual intervention.
	if !wasReachable {
		m.hub.connectHook()(ctx, m.node)
	}
	// Healthy -> unhealthy edge triggers failover exactly once per transition.
	if wasHealthy && !nowHealthy {
		m.triggerFailover(ctx)
	}
}

func (m *managedNode) markUnhealthy(ctx context.Context) {
	m.mu.Lock()
	wasHealthy := m.status == domain.NodeConnected && m.health.CoreRunning
	m.status = domain.NodeDisconnected
	m.health.CoreRunning = false
	m.reachable = false // next successful poll will re-trigger a resync
	m.mu.Unlock()
	if wasHealthy {
		m.triggerFailover(ctx)
	}
}

func (m *managedNode) triggerFailover(ctx context.Context) {
	target, ok := selectFailoverTarget(m.hub.HealthyNodes(), m.node.ID)
	var t *domain.Node
	if ok {
		t = target
	}
	m.hub.log.Warn("node unhealthy, triggering failover",
		"failed", m.node.Name, "target", targetName(t))
	m.hub.failover()(ctx, m.node, t)
}

func targetName(n *domain.Node) string {
	if n == nil {
		return "<none available>"
	}
	return n.Name
}

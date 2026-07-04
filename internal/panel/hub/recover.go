package hub

import (
	"context"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

func (m *managedNode) maybeAutoRecoverCore(ctx context.Context) {
	if !m.hub.opts.AutoRecoverCore {
		return
	}
	m.mu.Lock()
	if m.reachable && !m.health.CoreRunning && !m.coreDownSince.IsZero() {
		since := time.Since(m.coreDownSince)
		cooldownOK := m.lastAutoRecover.IsZero() || time.Since(m.lastAutoRecover) >= m.hub.opts.AutoRecoverCoreCooldown
		if since >= m.hub.opts.AutoRecoverCoreAfter && cooldownOK {
			node := m.node
			downFor := since
			m.lastAutoRecover = time.Now()
			m.mu.Unlock()
			m.autoRestartCore(ctx, node, downFor)
			return
		}
	}
	m.mu.Unlock()
}

func (m *managedNode) autoRestartCore(ctx context.Context, node *domain.Node, downFor time.Duration) {
	m.hub.log.Warn("auto-recovering node core",
		"node", node.Name,
		"down_for", downFor.Round(time.Second).String(),
	)
	rctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	conn, err := m.ensureConn()
	if err != nil {
		m.hub.log.Error("auto-recover dial failed", "node", node.Name, "err", err)
		return
	}
	if err := conn.RestartCore(rctx); err != nil {
		m.hub.log.Error("auto-recover restart core failed", "node", node.Name, "err", err)
		return
	}
	m.hub.log.Info("auto-recover restart core succeeded", "node", node.Name)
	if fn := m.hub.autoRecoverHook(); fn != nil {
		go fn(ctx, node, "restart_core")
	}
}

func (h *Hub) runHubWatchdog(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.maybeRecoverAllConnections(ctx)
		}
	}
}

func (h *Hub) maybeRecoverAllConnections(ctx context.Context) {
	h.mu.RLock()
	if !h.lastHubRecover.IsZero() && time.Since(h.lastHubRecover) < h.opts.AutoRecoverHubCooldown {
		h.mu.RUnlock()
		return
	}
	nodes := make([]*managedNode, 0, len(h.conns))
	for _, mn := range h.conns {
		nodes = append(nodes, mn)
	}
	h.mu.RUnlock()
	if len(nodes) == 0 {
		return
	}

	now := time.Now()
	for _, mn := range nodes {
		mn.mu.Lock()
		healthy := mn.status == domain.NodeConnected && mn.health.CoreRunning
		stuckSince := mn.disconnectedAt
		mn.mu.Unlock()
		if healthy {
			return
		}
		if stuckSince.IsZero() || now.Sub(stuckSince) < h.opts.AutoRecoverHubAfter {
			return
		}
	}

	h.mu.Lock()
	if !h.lastHubRecover.IsZero() && time.Since(h.lastHubRecover) < h.opts.AutoRecoverHubCooldown {
		h.mu.Unlock()
		return
	}
	h.lastHubRecover = now
	h.mu.Unlock()

	h.log.Warn("all nodes unhealthy; resetting hub connections",
		"nodes", len(nodes),
		"threshold", h.opts.AutoRecoverHubAfter.String(),
	)
	for _, mn := range nodes {
		mn.mu.Lock()
		mn.reachable = false
		mn.mu.Unlock()
		mn.dropConn()
	}
	if fn := h.autoRecoverHook(); fn != nil {
		for _, mn := range nodes {
			node := mn.node
			go fn(ctx, node, "reset_connection")
		}
	}
}

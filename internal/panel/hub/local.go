package hub

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/decoy"
	"github.com/vortexui/vortexui/internal/domain"
)

// LocalConn adapts an in-process core.CoreDriver to NodeConn, letting the hub
// manage a proxy core running on the panel's own host with no gRPC agent. The
// "connection" never really drops, so Close is a no-op: the driver's lifecycle
// is owned by the caller (stopped on panel shutdown), not by the hub's reconnect
// logic — otherwise a transient stream error would tear the local core down.
type LocalConn struct {
	nodeID uuid.UUID
	driver core.CoreDriver
	decoy  *decoy.Server

	mu      sync.Mutex
	lastCfg *core.GeneratedConfig // remembered so Restart/Start can re-apply it
}

var _ NodeConn = (*LocalConn)(nil)

// NewLocalConn wraps driver as the in-process connection for nodeID.
func NewLocalConn(nodeID uuid.UUID, driver core.CoreDriver, decoySrv *decoy.Server) *LocalConn {
	return &LocalConn{nodeID: nodeID, driver: driver, decoy: decoySrv}
}

// Sync applies the full config by (re)starting the local core, remembering the
// config so Restart/Start can re-apply it. coreType is ignored: the driver is
// already a concrete engine.
func (c *LocalConn) Sync(ctx context.Context, cfg *core.GeneratedConfig, _ domain.CoreType) error {
	c.mu.Lock()
	c.lastCfg = cfg
	c.mu.Unlock()
	if err := c.driver.Start(ctx, cfg); err != nil {
		return err
	}
	if c.decoy != nil {
		_ = c.decoy.ReloadFromInbounds(cfg.Inbounds)
	}
	return nil
}

// AddUser provisions a user on the local core at runtime.
func (c *LocalConn) AddUser(ctx context.Context, inboundTag string, u *domain.User) error {
	return c.driver.AddUser(ctx, inboundTag, u)
}

// RemoveUser deprovisions a user from the local core at runtime.
func (c *LocalConn) RemoveUser(ctx context.Context, inboundTag string, userID uuid.UUID) error {
	return c.driver.RemoveUser(ctx, inboundTag, userID.String())
}

// Health returns the local core's liveness snapshot including version.
func (c *LocalConn) Health(ctx context.Context) (domain.NodeHealth, error) {
	h, err := c.driver.Health(ctx)
	if err != nil {
		return h, err
	}
	ver, _ := c.driver.Version(ctx)
	h.CoreVersion = ver
	return h, nil
}

// ConsumeTraffic drains the driver's delta stream into ingest, stamping each
// delta with the local node's id (the driver only knows per-user counters).
func (c *LocalConn) ConsumeTraffic(ctx context.Context, ingest func(domain.TrafficDelta)) error {
	ch, err := c.driver.StreamTraffic(ctx)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-ch:
			if !ok {
				return nil
			}
			d.NodeID = c.nodeID
			ingest(d)
		}
	}
}

// Close is a no-op: the local driver is owned and stopped by the caller, not by
// the hub's per-connection lifecycle.
func (c *LocalConn) Close() error { return nil }

// OnlineStats reports the local core's live per-user connection counts.
func (c *LocalConn) OnlineStats(ctx context.Context) (map[string]int, error) {
	return c.driver.OnlineStats(ctx)
}

// OnlineIPs reports the local core's distinct online source IPs for one user.
func (c *LocalConn) OnlineIPs(ctx context.Context, userID string) (map[string]int64, error) {
	return c.driver.OnlineIPList(ctx, userID)
}

// UpdateGeo refreshes the local core's geo routing databases.
func (c *LocalConn) UpdateGeo(ctx context.Context, geoipURL, geositeURL string) (int64, int64, error) {
	return c.driver.UpdateGeoAssets(ctx, geoipURL, geositeURL)
}

// Logs returns recent log lines from the local core.
func (c *LocalConn) Logs(ctx context.Context, limit int) ([]string, error) {
	return c.driver.Logs(ctx, limit)
}

// RestartCore re-applies the last synced config — restarting a running core or
// starting a stopped one. Passing nil (the old behaviour) crashed the builder.
func (c *LocalConn) RestartCore(ctx context.Context) error {
	c.mu.Lock()
	cfg := c.lastCfg
	c.mu.Unlock()
	if cfg == nil {
		return errors.New("local core has no config yet; create an inbound first")
	}
	return c.driver.Start(ctx, cfg)
}

// StopCore stops the local core.
func (c *LocalConn) StopCore(ctx context.Context) error {
	return c.driver.Stop(ctx)
}

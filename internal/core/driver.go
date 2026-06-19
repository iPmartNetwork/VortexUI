// Package core defines the engine-agnostic abstraction over proxy cores
// (Xray-core and sing-box). Every node talks to its local core exclusively
// through CoreDriver, so the rest of VortexUI never knows or cares which engine
// is running. Adding a new engine = implementing this one interface.
package core

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// CoreDriver is the single seam between VortexUI and a concrete proxy engine.
//
// Implementations must be safe for concurrent use: the node agent calls AddUser/
// RemoveUser from API-driven goroutines while StreamTraffic runs continuously.
type CoreDriver interface {
	// Type identifies the concrete engine (xray | singbox).
	Type() domain.CoreType

	// Start applies the full generated config and boots the engine process.
	// It must be idempotent: calling Start on a running core hot-reloads it.
	Start(ctx context.Context, cfg *GeneratedConfig) error

	// Stop gracefully shuts the engine down.
	Stop(ctx context.Context) error

	// Reload hot-applies a new config without dropping live connections where
	// the engine supports it, falling back to restart otherwise.
	Reload(ctx context.Context, cfg *GeneratedConfig) error

	// AddUser provisions a user on a specific inbound tag at runtime (no restart).
	AddUser(ctx context.Context, inboundTag string, u *domain.User) error

	// RemoveUser deprovisions a user from an inbound tag at runtime.
	RemoveUser(ctx context.Context, inboundTag string, userID string) error

	// StreamTraffic emits incremental usage deltas until ctx is cancelled. The
	// engine's absolute counters are read-and-reset, and the driver converts
	// them into deltas so the panel can sum them idempotently.
	StreamTraffic(ctx context.Context) (<-chan domain.TrafficDelta, error)

	// OnlineStats returns the number of live connections (online IPs) per user,
	// keyed by the user's stats email (the user UUID string). Engines that cannot
	// report this return an empty map rather than an error.
	OnlineStats(ctx context.Context) (map[string]int, error)

	// OnlineIPList returns the distinct source IPs currently online for one user
	// (by stats email == UUID), mapped to each IP's last-seen unix time. Used to
	// detect account sharing. Engines that cannot report this return an empty map.
	OnlineIPList(ctx context.Context, email string) (map[string]int64, error)

	// UpdateGeoAssets downloads geoip/geosite routing databases into the engine's
	// asset directory and reloads so they take effect. Empty URLs skip that file.
	// Returns the bytes written for each.
	UpdateGeoAssets(ctx context.Context, geoipURL, geositeURL string) (geoip, geosite int64, err error)

	// Logs returns up to limit of the most recent core log lines (oldest first);
	// limit <= 0 returns all retained lines.
	Logs(ctx context.Context, limit int) ([]string, error)

	// Health returns a live resource/liveness snapshot of the engine.
	Health(ctx context.Context) (domain.NodeHealth, error)

	// Version reports the running engine version string.
	Version(ctx context.Context) (string, error)
}

// GeneratedConfig is the engine-neutral, fully-resolved configuration produced
// by the config builder from domain objects. Each driver renders it into its
// engine's native JSON.
type GeneratedConfig struct {
	Inbounds []domain.Inbound
	// Users keyed by inbound tag -> the users bound to that inbound.
	UsersByInbound map[string][]*domain.User
	// Outbounds, Routing, and Balancers describe egress and traffic steering. All
	// optional: when empty the builder emits a safe direct-egress default so a
	// node with no explicit policy still works.
	Outbounds []domain.Outbound
	Routing   []domain.RoutingRule
	Balancers []domain.Balancer
	// LogLevel and other engine-wide knobs.
	LogLevel string
	// WireGuardPeers maps a WireGuard inbound's tag to its peers (one per bound
	// user). Populated by the sync layer; consumed by the sing-box builder.
	WireGuardPeers map[string][]domain.WireGuardPeer
}

// Builder renders engine-native configuration. Each engine ships its own Builder
// so protocol/transport quirks stay isolated from domain logic.
type Builder interface {
	Build(cfg *GeneratedConfig) ([]byte, error)
}

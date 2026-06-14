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
	// LogLevel and other engine-wide knobs.
	LogLevel string
}

// Builder renders engine-native configuration. Each engine ships its own Builder
// so protocol/transport quirks stay isolated from domain logic.
type Builder interface {
	Build(cfg *GeneratedConfig) ([]byte, error)
}

package xray

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// xrayAPI is the runtime control surface VortexUI uses on a live Xray process,
// implemented over Xray's local gRPC API (HandlerService + StatsService). It is
// an interface so the driver's orchestration and delta logic are unit-testable
// with a fake, and the one piece that binds to xray-core's generated stubs is
// isolated to a single implementation.
type xrayAPI interface {
	// AddUser provisions a user on an inbound via HandlerService.AlterInbound
	// (AddUserOperation) — no process restart. The inbound is passed so the
	// driver can build the protocol-specific account (vless flow, ss cipher, …).
	AddUser(ctx context.Context, in domain.Inbound, u *domain.User) error

	// RemoveUser removes a user (by its stats email == user UUID) from an inbound.
	RemoveUser(ctx context.Context, inboundTag, email string) error

	// QueryTraffic reads per-user counters via StatsService.QueryStats. When
	// reset is true Xray zeroes the counters as it returns them, so each call
	// yields exactly the traffic since the previous call — i.e. a delta. This is
	// what makes our accounting naturally incremental and restart-safe.
	QueryTraffic(ctx context.Context, reset bool) ([]UserTraffic, error)

	Close() error
}

// UserTraffic is one user's up/down counter sample. Email is the Xray stats key,
// which the builder sets to the user's UUID string.
type UserTraffic struct {
	Email string
	Up    int64
	Down  int64
}

// APIDialer opens an xrayAPI against the local API inbound at addr (host:port).
// Injecting the dialer keeps the driver decoupled from the concrete gRPC client.
type APIDialer func(addr string) (xrayAPI, error)

// defaultDialer connects to Xray's local gRPC API (HandlerService +
// StatsService) over a plaintext loopback link — see api_grpc.go.
func defaultDialer(addr string) (xrayAPI, error) {
	return dialGRPC(addr)
}

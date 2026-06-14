// Package port declares the interfaces (hexagonal "ports") that the panel
// service layer depends on. Concrete adapters (Postgres, Redis) live under
// internal/platform and are wired in at startup, keeping business logic free of
// infrastructure detail and trivially mockable in tests.
package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// UserRepository persists service users and their bindings.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetBySubToken(ctx context.Context, token string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f UserFilter) ([]*domain.User, int, error)

	// AddUsedTraffic atomically increments used_traffic by delta in a single
	// SQL statement (UPDATE ... SET used = used + $1) so concurrent delta
	// reports never race or lose writes.
	AddUsedTraffic(ctx context.Context, id uuid.UUID, delta int64) error

	// AddUsedTrafficBatch applies many per-user deltas in one statement — the
	// stats aggregator's flush path, collapsing N updates into a single query.
	AddUsedTrafficBatch(ctx context.Context, deltas map[uuid.UUID]int64) error

	// Bindings management for the user-centric model.
	SetInbounds(ctx context.Context, userID uuid.UUID, inboundIDs []uuid.UUID) error
	InboundsFor(ctx context.Context, userID uuid.UUID) ([]domain.Inbound, error)
}

// UserFilter parameterizes paginated listing.
type UserFilter struct {
	Search string
	Status domain.UserStatus
	Limit  int
	Offset int
}

// NodeRepository persists nodes and their live health.
type NodeRepository interface {
	Create(ctx context.Context, n *domain.Node) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Node, error)
	Update(ctx context.Context, n *domain.Node) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.Node, error)
	UpdateHealth(ctx context.Context, id uuid.UUID, h domain.NodeHealth) error
}

// InboundRepository persists inbounds.
type InboundRepository interface {
	Create(ctx context.Context, in *domain.Inbound) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Inbound, error)
	Update(ctx context.Context, in *domain.Inbound) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error)
}

// OutboundRepository persists per-node egress handlers.
type OutboundRepository interface {
	Create(ctx context.Context, o *domain.Outbound) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Outbound, error)
	Update(ctx context.Context, o *domain.Outbound) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Outbound, error)
}

// RoutingRepository persists per-node routing rules.
type RoutingRepository interface {
	Create(ctx context.Context, rule *domain.RoutingRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RoutingRule, error)
	Update(ctx context.Context, rule *domain.RoutingRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.RoutingRule, error)
}

// BalancerRepository persists per-node balancers.
type BalancerRepository interface {
	Create(ctx context.Context, b *domain.Balancer) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Balancer, error)
	Update(ctx context.Context, b *domain.Balancer) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Balancer, error)
}

// AdminRepository persists panel operators and roles.
type AdminRepository interface {
	Create(ctx context.Context, a *domain.Admin) error
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
	Update(ctx context.Context, a *domain.Admin) error
	GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error)
}

// TrafficRepository writes time-series samples (TimescaleDB hypertable).
type TrafficRepository interface {
	WriteBatch(ctx context.Context, points []domain.TrafficPoint) error
	UsageSeries(ctx context.Context, userID uuid.UUID, q SeriesQuery) ([]domain.TrafficPoint, error)
}

// SeriesQuery bounds a time-series read.
type SeriesQuery struct {
	FromUnix int64
	ToUnix   int64
	Bucket   string // e.g. "1h", "1d" — passed to time_bucket()
}

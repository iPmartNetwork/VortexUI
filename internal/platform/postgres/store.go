// Package postgres implements the panel's repository ports on top of PostgreSQL
// (with optional TimescaleDB) using sqlc-generated, type-safe queries. Each
// repository is a thin adapter that maps between domain types and db rows; all
// SQL lives in internal/platform/postgres/queries.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// Store owns the connection pool and constructs repositories that share it.
type Store struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// Open connects to Postgres and verifies the connection. The caller owns Close.
func Open(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Store{pool: pool, q: db.New(pool)}, nil
}

// Close releases the pool.
func (s *Store) Close() { s.pool.Close() }

// Users returns the user repository.
func (s *Store) Users() *UserRepo { return &UserRepo{pool: s.pool, q: s.q} }

// Nodes returns the node repository.
func (s *Store) Nodes() *NodeRepo { return &NodeRepo{q: s.q} }

// Inbounds returns the inbound repository.
func (s *Store) Inbounds() *InboundRepo { return &InboundRepo{q: s.q} }

// Outbounds returns the outbound repository.
func (s *Store) Outbounds() *OutboundRepo { return &OutboundRepo{q: s.q} }

// Routing returns the routing-rule repository.
func (s *Store) Routing() *RoutingRepo { return &RoutingRepo{q: s.q} }

// Balancers returns the balancer repository.
func (s *Store) Balancers() *BalancerRepo { return &BalancerRepo{q: s.q} }

// Backup returns the configuration backup/restore repository.
func (s *Store) Backup() *BackupRepo { return &BackupRepo{pool: s.pool} }

// APITokens returns the API token repository.
func (s *Store) APITokens() *APITokenRepo { return &APITokenRepo{q: s.q} }

// Traffic returns the traffic repository.
func (s *Store) Traffic() *TrafficRepo { return &TrafficRepo{q: s.q} }

// Admins returns the admin repository.
func (s *Store) Admins() *AdminRepo { return &AdminRepo{q: s.q} }

// Audit returns the audit-log repository.
func (s *Store) Audit() *AuditRepo { return &AuditRepo{q: s.q} }

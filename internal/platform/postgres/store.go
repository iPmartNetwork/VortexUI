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

// Migration returns the migration-event repository.
func (s *Store) Migration() *MigrationRepo { return &MigrationRepo{pool: s.pool} }

// Monitor returns the live-monitor repository (traffic-derived online users).
func (s *Store) Monitor() *MonitorRepo { return &MonitorRepo{pool: s.pool} }

// Tickets returns the ticket repository.
func (s *Store) Tickets() *TicketRepo { return &TicketRepo{pool: s.pool} }

// Plans returns the plan/order repository.
func (s *Store) Plans() *PlanRepo { return &PlanRepo{pool: s.pool} }

// Analytics returns the analytics repository.
func (s *Store) Analytics() *AnalyticsRepo { return &AnalyticsRepo{pool: s.pool} }

// QuotaPolicies returns the quota policy repository.
func (s *Store) QuotaPolicies() *QuotaPolicyRepo { return &QuotaPolicyRepo{pool: s.pool} }

// RelayChains returns the relay chain repository.
func (s *Store) RelayChains() *RelayChainRepo { return &RelayChainRepo{pool: s.pool} }

// DecoySites returns the decoy site repository.
func (s *Store) DecoySites() *DecoySiteRepo { return &DecoySiteRepo{pool: s.pool} }

// RealityScans returns the reality scan repository.
func (s *Store) RealityScans() *RealityScanRepo { return &RealityScanRepo{pool: s.pool} }

// Probing returns the probing repository.
func (s *Store) Probing() *ProbingRepo { return &ProbingRepo{pool: s.pool} }

// Families returns the family repository.
func (s *Store) Families() *FamilyRepo { return &FamilyRepo{pool: s.pool} }

// Referrals returns the referral repository.
func (s *Store) Referrals() *ReferralRepo { return &ReferralRepo{pool: s.pool} }

// DoH returns the DNS-over-HTTPS repository.
func (s *Store) DoH() *DoHRepo { return &DoHRepo{pool: s.pool} }

// SNIDomains returns the SNI domain repository.
func (s *Store) SNIDomains() *SNIDomainRepo { return &SNIDomainRepo{pool: s.pool} }

// TLSTricks returns the TLS tricks repository.
func (s *Store) TLSTricks() *TLSTricksRepo { return &TLSTricksRepo{pool: s.pool} }

// Fingerprints returns the fingerprint repository.
func (s *Store) Fingerprints() *FingerprintRepo { return &FingerprintRepo{pool: s.pool} }

// Federation returns the federation repository.
func (s *Store) Federation() *FederationRepo { return &FederationRepo{pool: s.pool} }

// DeepLinks returns the deep link repository.
func (s *Store) DeepLinks() *DeepLinkRepo { return &DeepLinkRepo{pool: s.pool} }

// QuotaNotify returns the quota notification repository.
func (s *Store) QuotaNotify() *QuotaNotifyRepo { return &QuotaNotifyRepo{pool: s.pool} }

// SubSettings returns the subscription-settings repository.
func (s *Store) SubSettings() *SubSettingsRepo { return &SubSettingsRepo{pool: s.pool} }

// UserGeo returns the per-user geo (country) repository.
func (s *Store) UserGeo() *UserGeoRepo { return &UserGeoRepo{pool: s.pool} }

// WireGuardPeers returns the WireGuard per-user peer repository.
func (s *Store) WireGuardPeers() *WireGuardPeerRepo { return &WireGuardPeerRepo{pool: s.pool} }

package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// InboundLister yields a node's inbounds (port.InboundRepository satisfies it).
type InboundLister interface {
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error)
}

// OutboundLister yields a node's outbounds (port.OutboundRepository satisfies it).
type OutboundLister interface {
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Outbound, error)
}

// RoutingLister yields a node's routing rules (port.RoutingRepository satisfies it).
type RoutingLister interface {
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.RoutingRule, error)
}

// BalancerLister yields a node's balancers (port.BalancerRepository satisfies it).
type BalancerLister interface {
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Balancer, error)
}

// UsersByNoder yields the per-inbound user map for a node (postgres.UserRepo
// satisfies it).
type UsersByNoder interface {
	UsersByNode(ctx context.Context, nodeID uuid.UUID) (map[string][]*domain.User, error)
}

// Syncer pushes a fully-assembled config to one node (*hub.Hub satisfies it).
type Syncer interface {
	Sync(ctx context.Context, nodeID uuid.UUID, cfg *core.GeneratedConfig) error
}

// SyncService rebuilds a node's complete desired configuration from the database
// and pushes it to the live core. It is invoked whenever a node's inbounds,
// outbounds, routing rules, or balancers change; per-user edits use the lighter
// AddUser/RemoveUser path instead.
type SyncService struct {
	inbounds  InboundLister
	users     UsersByNoder
	outbounds OutboundLister
	routing   RoutingLister
	balancers BalancerLister
	syncer    Syncer
	wireguard *WireGuardService
	security  *SecuritySyncHelper
}

// NewSyncService wires the service. The outbound/routing/balancer listers are
// optional (nil disables that section), which keeps tests that only exercise
// inbounds simple and lets the panel run before those features are configured.
func NewSyncService(inbounds InboundLister, users UsersByNoder, syncer Syncer, outbounds OutboundLister, routing RoutingLister, balancers BalancerLister) *SyncService {
	return &SyncService{
		inbounds:  inbounds,
		users:     users,
		outbounds: outbounds,
		routing:   routing,
		balancers: balancers,
		syncer:    syncer,
	}
}

// SetWireGuard attaches the WireGuard service so the sync layer provisions a
// per-user peer for each WireGuard inbound and feeds them to the builder. It is
// optional: a nil WireGuardService leaves WireGuard peers unpopulated.
func (s *SyncService) SetWireGuard(wg *WireGuardService) { s.wireguard = wg }

// SetSecurity attaches probing/SNI/decoy runtime enrichment for resync.
func (s *SyncService) SetSecurity(h *SecuritySyncHelper) { s.security = h }

// Resync assembles the node's enabled inbounds plus their bound users, and its
// outbounds, routing rules, and balancers, into a core.GeneratedConfig and
// pushes it. Disabled inbounds are excluded so they stop listening after the
// sync; disabled outbounds/routing/balancers are passed through and skipped by
// the builder, keeping the enabled/disabled decision in one place per concern.
func (s *SyncService) Resync(ctx context.Context, nodeID uuid.UUID) error {
	inbounds, err := s.inbounds.ListByNode(ctx, nodeID)
	if err != nil {
		return err
	}
	usersByInbound, err := s.users.UsersByNode(ctx, nodeID)
	if err != nil {
		return err
	}
	// Defense in depth: never push limited/expired/disabled users even if the
	// DB read model is stale or unfiltered.
	usersByInbound = filterProvisionableUsers(usersByInbound, time.Now())

	var securityRules []domain.RoutingRule
	if s.security != nil {
		securityRules, err = s.security.EnrichInbounds(ctx, nodeID, inbounds)
		if err != nil {
			return err
		}
	}

	cfg := &core.GeneratedConfig{
		LogLevel:       "warning",
		UsersByInbound: usersByInbound,
	}
	for _, in := range inbounds {
		if in.Enabled {
			cfg.Inbounds = append(cfg.Inbounds, *in)
			// Enforce the inbound's geo policy by appending its geo-blocking
			// routing rules. GeoBlockRules returns nil when the inbound has no
			// GeoPolicy, so this is a no-op for unrestricted inbounds. These rules
			// land before the operator's own rules (added below), but renderers
			// sort by Priority and geo rules use -1/0, so ordering is preserved.
			cfg.Routing = append(cfg.Routing, core.GeoBlockRules(*in)...)
		}
	}

	if s.wireguard != nil {
		cfg.WireGuardPeers = map[string][]domain.WireGuardPeer{}
		for _, in := range inbounds {
			if in.Enabled && in.Protocol == domain.ProtoWireGuard {
				peers, err := s.wireguard.EnsurePeers(ctx, *in, usersByInbound[in.Tag])
				if err != nil {
					return err
				}
				cfg.WireGuardPeers[in.Tag] = peers
			}
		}
	}

	if s.outbounds != nil {
		outs, err := s.outbounds.ListByNode(ctx, nodeID)
		if err != nil {
			return err
		}
		for _, o := range outs {
			cfg.Outbounds = append(cfg.Outbounds, *o)
		}
	}
	if s.routing != nil {
		rules, err := s.routing.ListByNode(ctx, nodeID)
		if err != nil {
			return err
		}
		for _, r := range rules {
			cfg.Routing = append(cfg.Routing, *r)
		}
	}
	cfg.Routing = append(cfg.Routing, securityRules...)

	if s.balancers != nil {
		bals, err := s.balancers.ListByNode(ctx, nodeID)
		if err != nil {
			return err
		}
		for _, b := range bals {
			cfg.Balancers = append(cfg.Balancers, *b)
		}
	}

	return s.syncer.Sync(ctx, nodeID, cfg)
}

// filterProvisionableUsers drops accounts that must not be present on a live
// core: disabled/limited/expired, or active rows that already crossed data/expiry.
func filterProvisionableUsers(in map[string][]*domain.User, now time.Time) map[string][]*domain.User {
	if len(in) == 0 {
		return in
	}
	out := make(map[string][]*domain.User, len(in))
	for tag, users := range in {
		kept := make([]*domain.User, 0, len(users))
		for _, u := range users {
			if u == nil {
				continue
			}
			switch u.Status {
			case domain.UserStatusDisabled, domain.UserStatusLimited, domain.UserStatusExpired:
				continue
			}
			if !u.IsActive(now) {
				continue
			}
			kept = append(kept, u)
		}
		if len(kept) > 0 {
			out[tag] = kept
		}
	}
	return out
}

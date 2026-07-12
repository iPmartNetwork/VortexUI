package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SecuritySyncHelper enriches node configs with probing blocks, SNI certs, decoy
// fallbacks, and SNI routing rules at sync time.
type SecuritySyncHelper struct {
	probing port.ProbingRepository
	sni     port.SNIDomainRepository
	decoy   port.DecoySiteRepository
	now     func() time.Time
}

// NewSecuritySyncHelper wires security runtime dependencies for sync.
func NewSecuritySyncHelper(probing port.ProbingRepository, sni port.SNIDomainRepository, decoy port.DecoySiteRepository) *SecuritySyncHelper {
	return &SecuritySyncHelper{
		probing: probing,
		sni:     sni,
		decoy:   decoy,
		now:     time.Now,
	}
}

// EnrichInbounds mutates enabled inbounds in-place and returns extra routing rules.
func (h *SecuritySyncHelper) EnrichInbounds(ctx context.Context, nodeID uuid.UUID, inbounds []*domain.Inbound) ([]domain.RoutingRule, error) {
	if h == nil {
		return nil, nil
	}
	var rules []domain.RoutingRule
	var inboundTags []string

	decoy, _ := h.loadDecoy(ctx, nodeID)

	for _, in := range inbounds {
		if in == nil || !in.Enabled {
			continue
		}
		inboundTags = append(inboundTags, in.Tag)

		if h.sni != nil {
			domains, err := h.sni.ListByInbound(ctx, in.ID)
			if err != nil {
				return nil, err
			}
			core.ApplyManagedTLS(in, domains, func(domainName string) (*domain.SSLCertificate, error) {
				return h.sni.GetCertByDomain(ctx, domainName)
			})
			routes, err := h.sni.ListRoutesByInbound(ctx, in.ID)
			if err != nil {
				return nil, err
			}
			rules = append(rules, core.SNIRouteRules(nodeID, in.Tag, routes)...)
		}

		if decoy != nil {
			core.ApplyDecoy(in, decoy)
		} else if h.probing != nil {
			if policy, err := h.probing.GetPolicy(ctx); err == nil && policy != nil &&
				policy.Enabled && policy.Action == domain.ProbingHoneypot && policy.HoneypotHTML != "" {
				core.ApplyHoneypotDecoy(in, policy.HoneypotHTML)
			}
		}
	}

	if h.probing != nil {
		blocked, err := h.probing.ListBlockedIPs(ctx)
		if err != nil {
			return nil, err
		}
		ips := core.ActiveBlockedIPs(blocked, h.now())
		rules = append(rules, core.ProbeBlockRules(nodeID, ips, inboundTags)...)
	}

	return rules, nil
}

func (h *SecuritySyncHelper) loadDecoy(ctx context.Context, nodeID uuid.UUID) (*domain.DecoySite, error) {
	if h.decoy == nil {
		return nil, nil
	}
	if d, err := h.decoy.GetByNode(ctx, nodeID); err == nil && d != nil && d.Enabled {
		return d, nil
	}
	return h.decoy.GetGlobal(ctx)
}

// FleetResync triggers node config rebuilds after security policy changes.
type FleetResync struct {
	nodes port.NodeRepository
	sync  *SyncService
}

// NewFleetResync wires fleet-wide resync.
func NewFleetResync(nodes port.NodeRepository, sync *SyncService) *FleetResync {
	return &FleetResync{nodes: nodes, sync: sync}
}

// All resyncs every node.
func (f *FleetResync) All(ctx context.Context) error {
	if f == nil || f.nodes == nil || f.sync == nil {
		return nil
	}
	list, err := f.nodes.List(ctx)
	if err != nil {
		return err
	}
	for _, n := range list {
		if n == nil {
			continue
		}
		if err := f.sync.Resync(ctx, n.ID); err != nil {
			return err
		}
	}
	return nil
}

// Node resyncs one node when nodeID is set, otherwise all nodes.
func (f *FleetResync) Node(ctx context.Context, nodeID *uuid.UUID) error {
	if f == nil || f.sync == nil {
		return nil
	}
	if nodeID != nil {
		return f.sync.Resync(ctx, *nodeID)
	}
	return f.All(ctx)
}

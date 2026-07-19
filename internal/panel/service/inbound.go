package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/core/wireguard"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/pki"
)

// provisionSecurity fills in the material a secure inbound needs so the core
// never starts with an unusable (and crash-inducing) security block: REALITY
// inbounds get an auto-generated keypair + short id, and TLS inbounds get a
// self-signed certificate. Operators who supply a full streamSettings override,
// or who already provided the material, are left untouched.
func provisionSecurity(in *domain.Inbound) {
	if in.Raw == nil {
		in.Raw = map[string]any{}
	}
	if _, ok := in.Raw["streamSettings"]; ok {
		return // full manual override
	}
	// WireGuard server inbounds need a server keypair + interface defaults so the
	// sing-box endpoint config is valid. Mirrors the reality keypair provisioning.
	if in.Protocol == domain.ProtoWireGuard {
		if m, ok := in.Raw["wireguard"].(map[string]any); ok {
			if pk, _ := m["private_key"].(string); pk != "" {
				return
			}
		}
		priv, pub, err := wireguard.GenerateKeypair()
		if err != nil {
			return
		}
		port := in.Port
		in.Raw["wireguard"] = map[string]any{
			"private_key": priv,
			"public_key":  pub,
			"listen_port": port,
			"subnet":      "10.7.0.0/24",
			"mtu":         1420,
			"dns":         "1.1.1.1",
		}
		return
	}
	// Hysteria2 and TUIC mandate TLS; ensure they carry a certificate even if the
	// admin left security unset.
	if (in.Protocol == domain.ProtoHysteria2 || in.Protocol == domain.ProtoTUIC) && in.Security != domain.SecurityReality {
		in.Security = domain.SecurityTLS
	}
	switch in.Security {
	case domain.SecurityReality:
		if reality.ParseParams(in.Raw["reality"]).PrivateKey != "" {
			syncRealitySNI(in)
			return
		}
		kp, err := reality.GenerateKeypair()
		if err != nil {
			return
		}
		sid, _ := reality.ShortID(8)
		sni := "www.cloudflare.com"
		if len(in.SNI) > 0 && in.SNI[0] != "" {
			sni = in.SNI[0]
		}
		in.Raw["reality"] = map[string]any{
			"private_key":  kp.PrivateKey,
			"public_key":   kp.PublicKey,
			"short_ids":    []string{sid},
			"server_names": []string{sni},
			"dest":         sni + ":443",
		}
	case domain.SecurityTLS:
		if _, ok := in.Raw["tls"]; ok {
			return
		}
		host := ""
		if len(in.SNI) > 0 {
			host = in.SNI[0]
		}
		cert, key, err := pki.SelfSignedServer(host)
		if err != nil {
			return
		}
		in.Raw["tls"] = map[string]any{"certificate": cert, "key": key}
	}
}

// InboundService manages inbounds and reconciles the owning node's live config
// after every change via the SyncService.
type InboundService struct {
	repo  port.InboundRepository
	nodes port.NodeRepository
	sync  *SyncService
}

// NewInboundService wires the service.
func NewInboundService(repo port.InboundRepository, nodes port.NodeRepository, sync *SyncService) *InboundService {
	return &InboundService{repo: repo, nodes: nodes, sync: sync}
}

// coreSupports reports whether the given core can run this protocol+network+
// security combination, returning a clear error describing the incompatibility
// (or nil if supported). It is a thin wrapper over the single per-core
// capability matrix (core.Supports) so the guard never drifts from the matrix
// the API and UI consume.
func coreSupports(coreType domain.CoreType, proto domain.Protocol, network string, security domain.Security) error {
	return core.Supports(coreType, proto, network, security)
}

func validateInboundCore(node *domain.Node, override domain.CoreType, proto domain.Protocol, network string, security domain.Security) error {
	ct := node.ResolveInboundCore(override)
	if !node.CoreEnabled(ct) {
		return fmt.Errorf("core %q is not enabled on node %q (enabled: %v)", ct, node.Name, node.NormalizedEnabledCores())
	}
	return coreSupports(ct, proto, network, security)
}

func effectiveInboundCore(existing *domain.Inbound, override *domain.CoreType) domain.CoreType {
	if override != nil {
		return *override
	}
	return existing.Core
}

// CreateInboundInput describes a new inbound.
type CreateInboundInput struct {
	NodeID     uuid.UUID
	Tag        string
	Core       domain.CoreType
	Protocol   domain.Protocol
	Listen     string
	Port       int
	PortEnd    int
	Network    string
	Security   domain.Security
	SNI        []string
	Path       string
	Host       []string
	Flow       string
	Raw        map[string]any
	Enabled    bool
	SpeedLimit int64
	Notes      string
	GeoPolicy  *domain.GeoPolicy
}

// needsSingbox reports whether the protocol requires the sing-box engine.
func needsSingbox(proto domain.Protocol) bool {
	switch proto {
	case domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoHysteria,
		domain.ProtoWireGuard, domain.ProtoShadowTLS, domain.ProtoAnyTLS, domain.ProtoNaive:
		return true
	default:
		return false
	}
}

// Create persists an inbound then resyncs its node so the core starts listening.
// The inbound is returned even if the resync fails (it is durable; a later
// resync reconciles), with the sync error wrapped as a warning.
func (s *InboundService) Create(ctx context.Context, in CreateInboundInput) (*domain.Inbound, error) {
	if in.Tag == "" || in.Port == 0 {
		return nil, errors.New("tag and port are required")
	}
	node, err := s.nodes.GetByID(ctx, in.NodeID)
	if err != nil {
		return nil, errors.New("node not found")
	}

	// Auto-enable sing-box on the node when creating a sing-box-only protocol.
	if needsSingbox(in.Protocol) && !node.CoreEnabled(domain.CoreSingbox) {
		node.EnabledCores = append(node.NormalizedEnabledCores(), domain.CoreSingbox)
		if in.Core == "" {
			in.Core = domain.CoreSingbox
		}
		_ = s.nodes.Update(ctx, node) // best-effort; sync will pick it up
	}

	if err := validateInboundCore(node, in.Core, in.Protocol, networkDefault(in.Protocol, in.Network), orSec(in.Security, domain.SecurityNone)); err != nil {
		return nil, err
	}
	inbound := &domain.Inbound{
		ID:         uuid.New(),
		NodeID:     in.NodeID,
		Tag:        in.Tag,
		Core:       in.Core,
		Protocol:   in.Protocol,
		Listen:     in.Listen,
		Port:       in.Port,
		PortEnd:    in.PortEnd,
		Network:    networkDefault(in.Protocol, in.Network),
		Security:   orSec(in.Security, domain.SecurityNone),
		SNI:        in.SNI,
		Path:       in.Path,
		Host:       in.Host,
		Flow:       in.Flow,
		Raw:        in.Raw,
		Enabled:    in.Enabled,
		SpeedLimit: in.SpeedLimit,
		Notes:      in.Notes,
		GeoPolicy:  in.GeoPolicy,
	}
	provisionSecurity(inbound)
	if inbound.Enabled {
		tag, err := s.portConflict(ctx, inbound.NodeID, uuid.Nil, inbound.Port, inbound.PortEnd, inbound.Listen)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			return nil, fmt.Errorf("port %d is already used by inbound %q on this node", inbound.Port, tag)
		}
	}
	if err := s.repo.Create(ctx, inbound); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, in.NodeID); err != nil {
		return inbound, errors.Join(errors.New("inbound saved but node resync failed"), err)
	}
	return inbound, nil
}

// UpdateInboundInput is the mutable subset of an inbound. NodeID, ID and tag are
// not changed here (moving an inbound between nodes is delete+create).
// Raw is a pointer so callers can omit it to leave the stored raw block untouched.
type UpdateInboundInput struct {
	Listen     string
	Port       int
	PortEnd    int
	Network    string
	Security   domain.Security
	Core       *domain.CoreType
	SNI        []string
	Path       string
	Host       []string
	Flow       string
	Raw        *map[string]any
	Enabled    bool
	SpeedLimit int64
	Notes      string
	GeoPolicy  *domain.GeoPolicy
}

// Update applies changes to an inbound and resyncs its node so the live core
// reflects them. Returns the durable object plus a wrapped resync warning if the
// node was unreachable.
func (s *InboundService) Update(ctx context.Context, id uuid.UUID, in UpdateInboundInput) (*domain.Inbound, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	node, err := s.nodes.GetByID(ctx, existing.NodeID)
	if err != nil {
		return nil, errors.New("node not found")
	}
	if err := validateInboundCore(node, effectiveInboundCore(existing, in.Core), existing.Protocol, networkDefault(existing.Protocol, in.Network), orSec(in.Security, domain.SecurityNone)); err != nil {
		return nil, err
	}
	existing.Listen = in.Listen
	existing.Port = in.Port
	existing.PortEnd = in.PortEnd
	existing.Network = networkDefault(existing.Protocol, in.Network)
	existing.Security = orSec(in.Security, domain.SecurityNone)
	existing.SNI = in.SNI
	existing.Path = in.Path
	existing.Host = in.Host
	existing.Flow = in.Flow
	if in.Core != nil {
		existing.Core = *in.Core
	}
	if in.Raw != nil {
		existing.Raw = *in.Raw
	}
	existing.Enabled = in.Enabled
	existing.SpeedLimit = in.SpeedLimit
	existing.Notes = in.Notes
	existing.GeoPolicy = in.GeoPolicy
	provisionSecurity(existing)
	if existing.Enabled {
		tag, err := s.portConflict(ctx, existing.NodeID, existing.ID, existing.Port, existing.PortEnd, existing.Listen)
		if err != nil {
			return nil, err
		}
		if tag != "" {
			return nil, fmt.Errorf("port %d is already used by inbound %q on this node", existing.Port, tag)
		}
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, existing.NodeID); err != nil {
		return existing, errors.Join(errors.New("inbound saved but node resync failed"), err)
	}
	return existing, nil
}

// Delete removes an inbound and resyncs its node so the core stops listening.
func (s *InboundService) Delete(ctx context.Context, id uuid.UUID) error {
	in, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.sync.Resync(ctx, in.NodeID)
}

// ListByNode returns a node's inbounds.
func (s *InboundService) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error) {
	return s.repo.ListByNode(ctx, nodeID)
}

// ListFleet returns every inbound with its node name (single query).
func (s *InboundService) ListFleet(ctx context.Context) ([]domain.InboundListItem, error) {
	return s.repo.ListFleet(ctx)
}

// Clone duplicates an inbound with a new tag and port.
// If newPort is 0, a random port in the 10000-60000 range is assigned.
func (s *InboundService) Clone(ctx context.Context, id uuid.UUID, newPort int) (*domain.Inbound, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if newPort == 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(50000))
		if err != nil {
			return nil, fmt.Errorf("generate random port: %w", err)
		}
		newPort = 10000 + int(n.Int64())
	}
	input := CreateInboundInput{
		NodeID:     existing.NodeID,
		Tag:        existing.Tag + "-copy",
		Core:       existing.Core,
		Protocol:   existing.Protocol,
		Listen:     existing.Listen,
		Port:       newPort,
		PortEnd:    0, // clone is always single-port
		Network:    existing.Network,
		Security:   existing.Security,
		SNI:        existing.SNI,
		Path:       existing.Path,
		Host:       existing.Host,
		Flow:       existing.Flow,
		Raw:        existing.Raw,
		Enabled:    existing.Enabled,
		SpeedLimit: existing.SpeedLimit,
		Notes:      existing.Notes,
		GeoPolicy:  existing.GeoPolicy,
	}
	return s.Create(ctx, input)
}

// BulkAction applies enable/disable/delete to multiple inbounds.
// It returns the count of successfully processed inbounds.
func (s *InboundService) BulkAction(ctx context.Context, ids []uuid.UUID, action string) (int, error) {
	if action != "enable" && action != "disable" && action != "delete" {
		return 0, fmt.Errorf("unsupported bulk action: %q", action)
	}
	affected := 0
	nodeSet := make(map[uuid.UUID]struct{})
	for _, id := range ids {
		switch action {
		case "enable":
			existing, err := s.repo.GetByID(ctx, id)
			if err != nil {
				continue
			}
			existing.Enabled = true
			if err := s.repo.Update(ctx, existing); err != nil {
				continue
			}
			nodeSet[existing.NodeID] = struct{}{}
			affected++
		case "disable":
			existing, err := s.repo.GetByID(ctx, id)
			if err != nil {
				continue
			}
			existing.Enabled = false
			if err := s.repo.Update(ctx, existing); err != nil {
				continue
			}
			nodeSet[existing.NodeID] = struct{}{}
			affected++
		case "delete":
			existing, err := s.repo.GetByID(ctx, id)
			if err != nil {
				continue
			}
			if err := s.repo.Delete(ctx, id); err != nil {
				continue
			}
			nodeSet[existing.NodeID] = struct{}{}
			affected++
		}
	}
	// Resync each unique node once.
	for nodeID := range nodeSet {
		_ = s.sync.Resync(ctx, nodeID)
	}
	return affected, nil
}

// CheckPort reports whether a port is available on a node.
// Returns (available, conflictTag, error).
func (s *InboundService) CheckPort(ctx context.Context, nodeID uuid.UUID, port int) (bool, string, error) {
	tag, err := s.portConflict(ctx, nodeID, uuid.Nil, port, 0, "")
	if err != nil {
		return false, "", err
	}
	if tag != "" {
		return false, tag, nil
	}
	return true, "", nil
}

// syncRealitySNI keeps Raw["reality"].server_names and dest aligned with the
// inbound SNI list when operators change SNI (e.g. via the REALITY scanner).
// Without this, stale server_names in raw would override the updated SNI field.
func syncRealitySNI(in *domain.Inbound) {
	if len(in.SNI) == 0 {
		return
	}
	rm, ok := in.Raw["reality"].(map[string]any)
	if !ok {
		rm = map[string]any{}
	} else {
		cp := make(map[string]any, len(rm))
		for k, v := range rm {
			cp[k] = v
		}
		rm = cp
	}
	names := make([]any, len(in.SNI))
	for i, s := range in.SNI {
		names[i] = s
	}
	rm["server_names"] = names
	rm["dest"] = in.SNI[0] + ":443"
	in.Raw["reality"] = rm
}

// networkDefault returns the appropriate default network for a protocol.
// UDP-native protocols (hysteria2, tuic, wireguard, hysteria) default to empty
// (their transport is intrinsic), while TCP-based protocols default to "tcp".
func networkDefault(proto domain.Protocol, network string) string {
	if network != "" {
		return network
	}
	switch proto {
	case domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoWireGuard, domain.ProtoHysteria:
		return "" // UDP-native — no network layer
	default:
		return "tcp"
	}
}

func orStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func orSec(v, def domain.Security) domain.Security {
	if v == "" {
		return def
	}
	return v
}

// portsOverlap reports whether two port ranges [aStart, aEnd] and [bStart, bEnd] overlap.
// An end of 0 means single-port (end == start).
func portsOverlap(aStart, aEnd, bStart, bEnd int) bool {
	if aEnd == 0 {
		aEnd = aStart
	}
	if bEnd == 0 {
		bEnd = bStart
	}
	return aStart <= bEnd && bStart <= aEnd
}

// portConflict reports the tag of an existing ENABLED inbound on the same node
// that would collide with the given port range/listen, or "" if there is no conflict.
// Two inbounds clash when their port ranges overlap and their listen addresses overlap
// (an empty or 0.0.0.0 listen binds every interface). excludeID skips the
// inbound being updated.
func (s *InboundService) portConflict(ctx context.Context, nodeID, excludeID uuid.UUID, port, portEnd int, listen string) (string, error) {
	existing, err := s.repo.ListByNode(ctx, nodeID)
	if err != nil {
		return "", err
	}
	for _, e := range existing {
		if e.ID == excludeID || !e.Enabled {
			continue
		}
		if !portsOverlap(port, portEnd, e.Port, e.PortEnd) {
			continue
		}
		if listenOverlap(listen, e.Listen) {
			return e.Tag, nil
		}
	}
	return "", nil
}

// GetNodeIDForInbound returns the hosting node's ID for a given inbound.
func (s *InboundService) GetNodeIDForInbound(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	in, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return uuid.UUID{}, err
	}
	return in.NodeID, nil
}

// listenOverlap reports whether two listen addresses bind overlapping interfaces.
// An empty string or "0.0.0.0" means "all interfaces" and overlaps everything.
func listenOverlap(a, b string) bool {
	if a == "" || a == "0.0.0.0" || b == "" || b == "0.0.0.0" {
		return true
	}
	return a == b
}

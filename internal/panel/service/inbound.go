package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core/reality"
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
	// Hysteria2 and TUIC mandate TLS; ensure they carry a certificate even if the
	// admin left security unset.
	if (in.Protocol == domain.ProtoHysteria2 || in.Protocol == domain.ProtoTUIC) && in.Security != domain.SecurityReality {
		in.Security = domain.SecurityTLS
	}
	switch in.Security {
	case domain.SecurityReality:
		if reality.ParseParams(in.Raw["reality"]).PrivateKey != "" {
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
	repo port.InboundRepository
	sync *SyncService
}

// NewInboundService wires the service.
func NewInboundService(repo port.InboundRepository, sync *SyncService) *InboundService {
	return &InboundService{repo: repo, sync: sync}
}

// CreateInboundInput describes a new inbound.
type CreateInboundInput struct {
	NodeID   uuid.UUID
	Tag      string
	Protocol domain.Protocol
	Listen   string
	Port     int
	Network  string
	Security domain.Security
	SNI      []string
	Path     string
	Host     []string
	Flow     string
	Raw      map[string]any
	Enabled  bool
}

// Create persists an inbound then resyncs its node so the core starts listening.
// The inbound is returned even if the resync fails (it is durable; a later
// resync reconciles), with the sync error wrapped as a warning.
func (s *InboundService) Create(ctx context.Context, in CreateInboundInput) (*domain.Inbound, error) {
	if in.Tag == "" || in.Port == 0 {
		return nil, errors.New("tag and port are required")
	}
	inbound := &domain.Inbound{
		ID:       uuid.New(),
		NodeID:   in.NodeID,
		Tag:      in.Tag,
		Protocol: in.Protocol,
		Listen:   in.Listen,
		Port:     in.Port,
		Network:  orStr(in.Network, "tcp"),
		Security: orSec(in.Security, domain.SecurityNone),
		SNI:      in.SNI,
		Path:     in.Path,
		Host:     in.Host,
		Flow:     in.Flow,
		Raw:      in.Raw,
		Enabled:  in.Enabled,
	}
	provisionSecurity(inbound)
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
type UpdateInboundInput struct {
	Listen   string
	Port     int
	Network  string
	Security domain.Security
	SNI      []string
	Path     string
	Host     []string
	Flow     string
	Raw      map[string]any
	Enabled  bool
}

// Update applies changes to an inbound and resyncs its node so the live core
// reflects them. Returns the durable object plus a wrapped resync warning if the
// node was unreachable.
func (s *InboundService) Update(ctx context.Context, id uuid.UUID, in UpdateInboundInput) (*domain.Inbound, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing.Listen = in.Listen
	existing.Port = in.Port
	existing.Network = orStr(in.Network, "tcp")
	existing.Security = orSec(in.Security, domain.SecurityNone)
	existing.SNI = in.SNI
	existing.Path = in.Path
	existing.Host = in.Host
	existing.Flow = in.Flow
	existing.Raw = in.Raw
	existing.Enabled = in.Enabled
	provisionSecurity(existing)
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

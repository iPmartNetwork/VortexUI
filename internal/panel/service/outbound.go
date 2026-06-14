package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// OutboundService manages a node's egress handlers and reconciles the live core
// after every change via the SyncService.
type OutboundService struct {
	repo port.OutboundRepository
	sync *SyncService
}

// NewOutboundService wires the service.
func NewOutboundService(repo port.OutboundRepository, sync *SyncService) *OutboundService {
	return &OutboundService{repo: repo, sync: sync}
}

// CreateOutboundInput describes a new outbound.
type CreateOutboundInput struct {
	NodeID   uuid.UUID
	Tag      string
	Protocol domain.OutboundProtocol
	Address  string
	Port     int
	UUID     string
	Password string
	Username string
	Method   string
	Flow     string
	Network  string
	Security domain.Security
	SNI      string
	Path     string
	Host     string
	Raw      map[string]any
	Enabled  *bool // nil defaults to enabled
}

// Create validates and persists an outbound, then resyncs its node. Returns the
// durable object even if the resync fails (a later resync reconciles).
func (s *OutboundService) Create(ctx context.Context, in CreateOutboundInput) (*domain.Outbound, error) {
	o := &domain.Outbound{
		ID:       uuid.New(),
		NodeID:   in.NodeID,
		Tag:      in.Tag,
		Protocol: in.Protocol,
		Address:  in.Address,
		Port:     in.Port,
		UUID:     in.UUID,
		Password: in.Password,
		Username: in.Username,
		Method:   in.Method,
		Flow:     in.Flow,
		Network:  in.Network,
		Security: in.Security,
		SNI:      in.SNI,
		Path:     in.Path,
		Host:     in.Host,
		Raw:      in.Raw,
		Enabled:  boolOr(in.Enabled, true),
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, o.NodeID); err != nil {
		return o, errors.Join(errors.New("outbound saved but node resync failed"), err)
	}
	return o, nil
}

// UpdateOutboundInput is the mutable subset of an outbound. NodeID, ID and tag
// are immutable here (moving/renaming is delete+create).
type UpdateOutboundInput struct {
	Protocol domain.OutboundProtocol
	Address  string
	Port     int
	UUID     string
	Password string
	Username string
	Method   string
	Flow     string
	Network  string
	Security domain.Security
	SNI      string
	Path     string
	Host     string
	Raw      map[string]any
	Enabled  bool
}

// Update applies changes to an outbound and resyncs its node.
func (s *OutboundService) Update(ctx context.Context, id uuid.UUID, in UpdateOutboundInput) (*domain.Outbound, error) {
	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Protocol = in.Protocol
	o.Address = in.Address
	o.Port = in.Port
	o.UUID = in.UUID
	o.Password = in.Password
	o.Username = in.Username
	o.Method = in.Method
	o.Flow = in.Flow
	o.Network = in.Network
	o.Security = in.Security
	o.SNI = in.SNI
	o.Path = in.Path
	o.Host = in.Host
	o.Raw = in.Raw
	o.Enabled = in.Enabled
	if err := o.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, o); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, o.NodeID); err != nil {
		return o, errors.Join(errors.New("outbound saved but node resync failed"), err)
	}
	return o, nil
}

// Delete removes an outbound and resyncs its node.
func (s *OutboundService) Delete(ctx context.Context, id uuid.UUID) error {
	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.sync.Resync(ctx, o.NodeID)
}

// ListByNode returns a node's outbounds.
func (s *OutboundService) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Outbound, error) {
	return s.repo.ListByNode(ctx, nodeID)
}

// boolOr returns *p when non-nil, otherwise def.
func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

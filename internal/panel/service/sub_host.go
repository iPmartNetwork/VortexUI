package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SubHostService manages Marzban-style subscription host overrides bound to
// inbounds (CRUD + reorder). Validation is delegated to domain.SubHost.Validate.
type SubHostService struct {
	repo port.SubHostRepository
	now  func() time.Time
}

// NewSubHostService wires the subscription host service.
func NewSubHostService(repo port.SubHostRepository) *SubHostService {
	return &SubHostService{repo: repo, now: time.Now}
}

// SubHostInput carries the mutable fields of a subscription host. It is shared
// by create and update; ID/InboundID are supplied out-of-band.
type SubHostInput struct {
	InboundID     uuid.UUID
	Remark        string
	Address       string
	Port          *int
	SNI           string
	HostHeader    string
	Path          string
	ALPN          string
	Fingerprint   string
	Security      domain.HostSecurity
	AllowInsecure bool
	MuxEnable     bool
	Fragment      string
	Priority      int
	Enabled       bool
}

func (in SubHostInput) apply(h *domain.SubHost) {
	h.Remark = in.Remark
	h.Address = in.Address
	h.Port = in.Port
	h.SNI = in.SNI
	h.HostHeader = in.HostHeader
	h.Path = in.Path
	h.ALPN = in.ALPN
	h.Fingerprint = in.Fingerprint
	h.Security = in.Security
	h.AllowInsecure = in.AllowInsecure
	h.MuxEnable = in.MuxEnable
	h.Fragment = in.Fragment
	h.Priority = in.Priority
	h.Enabled = in.Enabled
}

// Create validates and persists a new host bound to an inbound.
func (s *SubHostService) Create(ctx context.Context, in SubHostInput) (*domain.SubHost, error) {
	if in.InboundID == uuid.Nil {
		return nil, errors.New("inbound_id is required")
	}
	if in.Security == "" {
		in.Security = domain.HostSecurityInboundDefault
	}
	h := &domain.SubHost{ID: uuid.New(), InboundID: in.InboundID, CreatedAt: s.now()}
	in.apply(h)
	if err := h.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

// Update validates and persists changes to an existing host.
func (s *SubHostService) Update(ctx context.Context, id uuid.UUID, in SubHostInput) (*domain.SubHost, error) {
	h, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, domain.ErrNotFound
	}
	if in.Security == "" {
		in.Security = domain.HostSecurityInboundDefault
	}
	// InboundID is immutable on update; ignore any value supplied in the input.
	in.InboundID = h.InboundID
	in.apply(h)
	if err := h.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

// Delete removes a host by ID.
func (s *SubHostService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// Get returns a single host by ID.
func (s *SubHostService) Get(ctx context.Context, id uuid.UUID) (*domain.SubHost, error) {
	h, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, domain.ErrNotFound
	}
	return h, nil
}

// ListByInbound returns an inbound's hosts in priority order.
func (s *SubHostService) ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SubHost, error) {
	return s.repo.ListByInbound(ctx, inboundID)
}

// Reorder assigns ascending priorities to the given hosts in the supplied order
// (first ID gets priority 0). IDs not present are left untouched.
func (s *SubHostService) Reorder(ctx context.Context, ordered []uuid.UUID) error {
	for i, id := range ordered {
		h, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if h == nil {
			return domain.ErrNotFound
		}
		h.Priority = i
		if err := s.repo.Update(ctx, h); err != nil {
			return err
		}
	}
	return nil
}

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SNIService manages multi-domain SNI routing and SSL certificates.
type SNIService struct {
	repo   port.SNIDomainRepository
	now    func() time.Time
	issuer CertIssuer
}

// CertIssuer obtains TLS certificates for domains (ACME or self-signed).
type CertIssuer interface {
	ObtainOrRenew(ctx context.Context, domain string) (certPEM, keyPEM string, err error)
}

// SetCertIssuer wires automatic certificate issuance.
func (s *SNIService) SetCertIssuer(i CertIssuer) { s.issuer = i }

// NewSNIService wires the service.
func NewSNIService(repo port.SNIDomainRepository) *SNIService {
	return &SNIService{repo: repo, now: time.Now}
}

// --- Domains ---

type AddDomainInput struct {
	InboundID uuid.UUID
	Domain    string
	AutoCert  bool
}

func (s *SNIService) AddDomain(ctx context.Context, in AddDomainInput) (*domain.SNIDomain, error) {
	if in.Domain == "" {
		return nil, errors.New("domain is required")
	}
	d := &domain.SNIDomain{
		ID:         uuid.New(),
		InboundID:  in.InboundID,
		Domain:     in.Domain,
		AutoCert:   in.AutoCert,
		CertStatus: domain.CertPending,
		CreatedAt:  s.now(),
	}
	if err := s.repo.CreateDomain(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *SNIService) ListDomains(ctx context.Context, inboundID *uuid.UUID) ([]*domain.SNIDomain, error) {
	if inboundID != nil {
		return s.repo.ListByInbound(ctx, *inboundID)
	}
	return s.repo.ListAll(ctx)
}

func (s *SNIService) DeleteDomain(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteDomain(ctx, id)
}

// --- Certificates ---

type IssueCertInput struct {
	Domain    string
	Wildcard  bool
	Issuer    string
	AutoRenew bool
}

func (s *SNIService) IssueCert(ctx context.Context, in IssueCertInput) (*domain.SSLCertificate, error) {
	if in.Domain == "" {
		return nil, errors.New("domain is required")
	}
	if in.Issuer == "" {
		in.Issuer = "letsencrypt"
	}
	c := &domain.SSLCertificate{
		ID:        uuid.New(),
		Domain:    in.Domain,
		Wildcard:  in.Wildcard,
		Issuer:    in.Issuer,
		Status:    domain.CertPending,
		AutoRenew: in.AutoRenew,
		CreatedAt: s.now(),
	}
	if err := s.repo.CreateCert(ctx, c); err != nil {
		return nil, err
	}
	if s.issuer != nil {
		_, _, err := s.issuer.ObtainOrRenew(ctx, in.Domain)
		if err != nil {
			c.Status = domain.CertFailed
		} else {
			c.Status = domain.CertActive
			exp := s.now().Add(90 * 24 * time.Hour)
			c.ExpiresAt = &exp
		}
		_ = s.repo.UpdateCert(ctx, c)
	}
	return c, nil
}

func (s *SNIService) ListCerts(ctx context.Context) ([]*domain.SSLCertificate, error) {
	return s.repo.ListCerts(ctx)
}

func (s *SNIService) DeleteCert(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCert(ctx, id)
}

func (s *SNIService) RenewCert(ctx context.Context, id uuid.UUID) (*domain.SSLCertificate, error) {
	c, err := s.repo.GetCert(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Status = domain.CertRenewing
	if err := s.repo.UpdateCert(ctx, c); err != nil {
		return nil, err
	}
	if s.issuer != nil {
		_, _, err := s.issuer.ObtainOrRenew(ctx, c.Domain)
		if err != nil {
			c.Status = domain.CertFailed
		} else {
			c.Status = domain.CertActive
			exp := s.now().Add(90 * 24 * time.Hour)
			c.ExpiresAt = &exp
		}
		_ = s.repo.UpdateCert(ctx, c)
	}
	return c, nil
}

// --- SNI Routes ---

type AddRouteInput struct {
	InboundID uuid.UUID
	SNI       string
	Action    string
	TargetTag string
	Priority  int
}

func (s *SNIService) AddRoute(ctx context.Context, in AddRouteInput) (*domain.SNIRoute, error) {
	if in.SNI == "" {
		return nil, errors.New("sni is required")
	}
	r := &domain.SNIRoute{
		ID:        uuid.New(),
		InboundID: in.InboundID,
		SNI:       in.SNI,
		Action:    in.Action,
		TargetTag: in.TargetTag,
		Priority:  in.Priority,
		Enabled:   true,
	}
	if err := s.repo.CreateRoute(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *SNIService) ListRoutes(ctx context.Context, inboundID uuid.UUID) ([]*domain.SNIRoute, error) {
	return s.repo.ListRoutesByInbound(ctx, inboundID)
}

func (s *SNIService) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRoute(ctx, id)
}

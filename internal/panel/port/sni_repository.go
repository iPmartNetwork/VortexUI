package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// SNIDomainRepository persists SNI domain routing and SSL certificates.
type SNIDomainRepository interface {
	// Domains
	CreateDomain(ctx context.Context, d *domain.SNIDomain) error
	GetDomain(ctx context.Context, id uuid.UUID) (*domain.SNIDomain, error)
	UpdateDomain(ctx context.Context, d *domain.SNIDomain) error
	DeleteDomain(ctx context.Context, id uuid.UUID) error
	ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SNIDomain, error)
	ListAll(ctx context.Context) ([]*domain.SNIDomain, error)

	// Certificates
	CreateCert(ctx context.Context, c *domain.SSLCertificate) error
	GetCert(ctx context.Context, id uuid.UUID) (*domain.SSLCertificate, error)
	GetCertByDomain(ctx context.Context, domain string) (*domain.SSLCertificate, error)
	UpdateCert(ctx context.Context, c *domain.SSLCertificate) error
	DeleteCert(ctx context.Context, id uuid.UUID) error
	ListCerts(ctx context.Context) ([]*domain.SSLCertificate, error)
	ListExpiringSoon(ctx context.Context, withinDays int) ([]*domain.SSLCertificate, error)

	// SNI Routes
	CreateRoute(ctx context.Context, r *domain.SNIRoute) error
	UpdateRoute(ctx context.Context, r *domain.SNIRoute) error
	DeleteRoute(ctx context.Context, id uuid.UUID) error
	ListRoutesByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SNIRoute, error)
}

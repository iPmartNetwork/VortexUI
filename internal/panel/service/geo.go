package service

import (
	"context"

	"github.com/google/uuid"
)

// GeoResolver resolves an IP to an ISO country code or ISP. *geoip.Resolver satisfies it.
type GeoResolver interface {
	Country(ip string) string
	ISP(ip string) string
}

// UserGeoRepository persists each user's last-known country (from sub fetches).
type UserGeoRepository interface {
	Upsert(ctx context.Context, userID uuid.UUID, country, ip string) error
}

// GeoService records where users fetch their subscription from, enabling the
// "Traffic by Country" analytics. Best-effort: failures never block a request.
type GeoService struct {
	resolver GeoResolver
	repo     UserGeoRepository
}

// NewGeoService wires the service.
func NewGeoService(resolver GeoResolver, repo UserGeoRepository) *GeoService {
	return &GeoService{resolver: resolver, repo: repo}
}

// RecordUserIP resolves ip to a country and upserts it for the user. No-op when
// the resolver is disabled or the IP is not resolvable.
func (s *GeoService) RecordUserIP(ctx context.Context, userID uuid.UUID, ip string) {
	if s == nil || s.resolver == nil || s.repo == nil {
		return
	}
	country := s.resolver.Country(ip)
	if country == "" {
		return
	}
	_ = s.repo.Upsert(ctx, userID, country, ip)
}

// DetectISP resolves an IP to a short ISP identifier (e.g. "mci", "irancell").
// Returns "" when the resolver is disabled or the ISP is unknown.
func (s *GeoService) DetectISP(ip string) string {
	if s == nil || s.resolver == nil {
		return ""
	}
	return s.resolver.ISP(ip)
}

package service

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// IPLimitService manages the IP-limit enforcement policy and exposes its event
// audit trail. The actual enforcement runs in ShareGuard, which reads the same
// policy from the repository each detection pass.
type IPLimitService struct {
	repo port.IPLimitRepository
}

// NewIPLimitService wires the service over the persistence port.
func NewIPLimitService(repo port.IPLimitRepository) *IPLimitService {
	return &IPLimitService{repo: repo}
}

// GetPolicy returns the persisted policy, falling back to the default when the
// store is unavailable so the admin UI always renders something sensible.
func (s *IPLimitService) GetPolicy(ctx context.Context) (*domain.IPLimitPolicy, error) {
	p, err := s.repo.GetPolicy(ctx)
	if err != nil {
		def := domain.DefaultIPLimitPolicy()
		return &def, nil
	}
	return p, nil
}

// UpdatePolicy persists the enforcement policy.
func (s *IPLimitService) UpdatePolicy(ctx context.Context, p *domain.IPLimitPolicy) error {
	return s.repo.UpdatePolicy(ctx, p)
}

// ListEvents returns the most recent enforcement events, newest first.
func (s *IPLimitService) ListEvents(ctx context.Context, limit int) ([]*domain.IPLimitEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, limit)
}

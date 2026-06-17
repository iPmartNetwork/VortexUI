package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// QuotaService manages fair-use quota policies.
type QuotaService struct {
	repo port.QuotaPolicyRepository
	now  func() time.Time
}

// NewQuotaService wires the quota policy service.
func NewQuotaService(repo port.QuotaPolicyRepository) *QuotaService {
	return &QuotaService{repo: repo, now: time.Now}
}

// CreatePolicyInput describes a new quota policy.
type CreatePolicyInput struct {
	Name  string
	Tiers []domain.QuotaTier
}

// CreatePolicy persists a new fair-use policy.
func (s *QuotaService) CreatePolicy(ctx context.Context, in CreatePolicyInput) (*domain.QuotaPolicy, error) {
	if in.Name == "" {
		return nil, errors.New("policy name is required")
	}
	if len(in.Tiers) == 0 {
		return nil, errors.New("at least one tier is required")
	}
	p := &domain.QuotaPolicy{
		ID:        uuid.New(),
		Name:      in.Name,
		Tiers:     in.Tiers,
		Enabled:   true,
		CreatedAt: s.now(),
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPolicies returns all quota policies.
func (s *QuotaService) ListPolicies(ctx context.Context) ([]*domain.QuotaPolicy, error) {
	return s.repo.List(ctx)
}

// UpdatePolicy updates an existing policy.
func (s *QuotaService) UpdatePolicy(ctx context.Context, id uuid.UUID, name string, tiers []domain.QuotaTier, enabled bool) (*domain.QuotaPolicy, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if tiers != nil {
		p.Tiers = tiers
	}
	p.Enabled = enabled
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeletePolicy removes a quota policy.
func (s *QuotaService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// EvaluateUser determines the current speed tier for a user given the active
// policy. Returns nil tier if no policy applies or user is within normal limits.
func (s *QuotaService) EvaluateUser(ctx context.Context, user *domain.User) (*domain.QuotaTier, error) {
	if user.DataLimit <= 0 {
		return nil, nil // unlimited user
	}
	policy, err := s.repo.GetDefault(ctx)
	if err != nil || policy == nil || !policy.Enabled {
		return nil, nil
	}
	ratio := float64(user.UsedTraffic) / float64(user.DataLimit)
	return policy.ActiveTier(ratio), nil
}

package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// FingerprintService manages TLS client fingerprint validation.
type FingerprintService struct {
	repo port.FingerprintRepository
	now  func() time.Time
}

func NewFingerprintService(repo port.FingerprintRepository) *FingerprintService {
	return &FingerprintService{repo: repo, now: time.Now}
}

func (s *FingerprintService) GetPolicy(ctx context.Context) (*domain.FingerprintPolicy, error) {
	p, err := s.repo.GetPolicy(ctx)
	if err != nil {
		def := domain.DefaultFingerprintPolicy()
		return &def, nil
	}
	return p, nil
}

func (s *FingerprintService) UpdatePolicy(ctx context.Context, p *domain.FingerprintPolicy) error {
	return s.repo.SavePolicy(ctx, p)
}

func (s *FingerprintService) CreateRule(ctx context.Context, name, fingerprint, ja3, action string, priority int) (*domain.FingerprintRule, error) {
	r := &domain.FingerprintRule{
		ID:          uuid.New(),
		Name:        name,
		Fingerprint: fingerprint,
		JA3Hash:     ja3,
		Action:      domain.FingerprintAction(action),
		Priority:    priority,
		Enabled:     true,
		CreatedAt:   s.now(),
	}
	if err := s.repo.CreateRule(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *FingerprintService) ListRules(ctx context.Context) ([]*domain.FingerprintRule, error) {
	return s.repo.ListRules(ctx)
}

func (s *FingerprintService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRule(ctx, id)
}

func (s *FingerprintService) ListEvents(ctx context.Context, limit int) ([]*domain.FingerprintEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, limit)
}

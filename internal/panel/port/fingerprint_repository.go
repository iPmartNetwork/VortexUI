package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// FingerprintRepository persists fingerprint rules and events.
type FingerprintRepository interface {
	GetPolicy(ctx context.Context) (*domain.FingerprintPolicy, error)
	SavePolicy(ctx context.Context, p *domain.FingerprintPolicy) error

	CreateRule(ctx context.Context, r *domain.FingerprintRule) error
	UpdateRule(ctx context.Context, r *domain.FingerprintRule) error
	DeleteRule(ctx context.Context, id uuid.UUID) error
	ListRules(ctx context.Context) ([]*domain.FingerprintRule, error)

	SaveEvent(ctx context.Context, e *domain.FingerprintEvent) error
	ListEvents(ctx context.Context, limit int) ([]*domain.FingerprintEvent, error)
}

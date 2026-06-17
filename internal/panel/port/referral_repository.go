package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// ReferralRepository persists referral codes, events, and config.
type ReferralRepository interface {
	GetConfig(ctx context.Context) (*domain.ReferralConfig, error)
	SaveConfig(ctx context.Context, c *domain.ReferralConfig) error

	CreateCode(ctx context.Context, rc *domain.ReferralCode) error
	GetCodeByCode(ctx context.Context, code string) (*domain.ReferralCode, error)
	GetCodeByUser(ctx context.Context, userID uuid.UUID) (*domain.ReferralCode, error)
	IncrementUses(ctx context.Context, id uuid.UUID) error
	ListCodes(ctx context.Context, limit, offset int) ([]*domain.ReferralCode, int, error)

	SaveEvent(ctx context.Context, e *domain.ReferralEvent) error
	ListEvents(ctx context.Context, userID *uuid.UUID, limit int) ([]*domain.ReferralEvent, error)
}

package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// DeviceRepository persists user device (HWID) registrations.
type DeviceRepository interface {
	Upsert(ctx context.Context, d *domain.Device) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, userID uuid.UUID, hwid string) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteAllForUsers(ctx context.Context, userIDs []uuid.UUID) error
	Exists(ctx context.Context, userID uuid.UUID, hwid string) (bool, error)
}

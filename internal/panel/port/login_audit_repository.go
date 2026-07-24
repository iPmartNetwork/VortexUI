package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// LoginAuditRepository persists login audit trail entries.
type LoginAuditRepository interface {
	Create(ctx context.Context, entry *domain.LoginAuditEntry) error
	ListByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.LoginAuditEntry, error)
	ListByIP(ctx context.Context, ip string, limit int) ([]*domain.LoginAuditEntry, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.LoginAuditEntry, error)
	CountFailedSince(ctx context.Context, adminID uuid.UUID, since time.Time) (int, error)
}

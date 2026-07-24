package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// SecurityAuditRepository persists security audit log entries.
type SecurityAuditRepository interface {
	Create(ctx context.Context, entry *domain.SecurityAuditEntry) error
	ListByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.SecurityAuditEntry, error)
	ListByOperation(ctx context.Context, operation string, limit int) ([]*domain.SecurityAuditEntry, error)
	ListRecent(ctx context.Context, limit int) ([]*domain.SecurityAuditEntry, error)
}

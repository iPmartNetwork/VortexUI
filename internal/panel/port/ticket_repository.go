package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// TicketRepository persists support tickets and messages.
type TicketRepository interface {
	Create(ctx context.Context, t *domain.Ticket) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error)
	Update(ctx context.Context, t *domain.Ticket) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Ticket, error)
	ListAll(ctx context.Context, status string, limit, offset int) ([]*domain.Ticket, int, error)

	AddMessage(ctx context.Context, m *domain.TicketMessage) error
	Messages(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error)
}

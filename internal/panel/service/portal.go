package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// PortalService handles end-user self-service operations: login via sub_token,
// view usage, purchase plans, and submit support tickets.
type PortalService struct {
	users   port.UserRepository
	tickets port.TicketRepository
	plans   PlanRepository
	now     func() time.Time
}

// NewPortalService wires dependencies.
func NewPortalService(users port.UserRepository, tickets port.TicketRepository, plans PlanRepository) *PortalService {
	return &PortalService{users: users, tickets: tickets, plans: plans, now: time.Now}
}

// LoginByToken authenticates an end-user by their subscription token and returns
// the user record (caller issues a portal JWT).
func (s *PortalService) LoginByToken(ctx context.Context, subToken string) (*domain.User, error) {
	u, err := s.users.GetBySubToken(ctx, subToken)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	if u.Status == domain.UserStatusDisabled {
		return nil, errors.New("account disabled")
	}
	return u, nil
}

// GetUsage returns the user's current usage stats.
func (s *PortalService) GetUsage(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, userID)
}

// ListPlans returns enabled plans for the self-service store.
func (s *PortalService) ListPlans(ctx context.Context) ([]*domain.Plan, error) {
	all, err := s.plans.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	var enabled []*domain.Plan
	for _, p := range all {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled, nil
}

// CreateTicketInput is the data needed to open a new ticket.
type CreateTicketInput struct {
	UserID   uuid.UUID
	Subject  string
	Body     string
	Priority domain.TicketPriority
}

// CreateTicket opens a new support ticket with an initial message.
func (s *PortalService) CreateTicket(ctx context.Context, in CreateTicketInput) (*domain.Ticket, error) {
	if in.Subject == "" || in.Body == "" {
		return nil, errors.New("subject and body are required")
	}
	if in.Priority == "" {
		in.Priority = domain.TicketPriorityMedium
	}
	now := s.now()
	t := &domain.Ticket{
		ID:        uuid.New(),
		UserID:    in.UserID,
		Subject:   in.Subject,
		Status:    domain.TicketOpen,
		Priority:  in.Priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.tickets.Create(ctx, t); err != nil {
		return nil, err
	}
	msg := &domain.TicketMessage{
		ID:        uuid.New(),
		TicketID:  t.ID,
		Sender:    "user",
		SenderID:  in.UserID,
		Body:      in.Body,
		CreatedAt: now,
	}
	if err := s.tickets.AddMessage(ctx, msg); err != nil {
		return nil, err
	}
	t.Messages = []domain.TicketMessage{*msg}
	return t, nil
}

// ReplyTicket adds a user reply to an existing ticket.
func (s *PortalService) ReplyTicket(ctx context.Context, ticketID, userID uuid.UUID, body string) error {
	t, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if t.UserID != userID {
		return errors.New("forbidden")
	}
	if t.Status == domain.TicketClosed {
		return errors.New("ticket is closed")
	}
	now := s.now()
	msg := &domain.TicketMessage{
		ID:        uuid.New(),
		TicketID:  ticketID,
		Sender:    "user",
		SenderID:  userID,
		Body:      body,
		CreatedAt: now,
	}
	if err := s.tickets.AddMessage(ctx, msg); err != nil {
		return err
	}
	t.Status = domain.TicketOpen
	t.UpdatedAt = now
	return s.tickets.Update(ctx, t)
}

// ListUserTickets returns tickets for a given user.
func (s *PortalService) ListUserTickets(ctx context.Context, userID uuid.UUID) ([]*domain.Ticket, error) {
	return s.tickets.ListByUser(ctx, userID)
}

// GetTicket returns a single ticket with messages (owned by user).
func (s *PortalService) GetTicket(ctx context.Context, ticketID, userID uuid.UUID) (*domain.Ticket, error) {
	t, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, errors.New("forbidden")
	}
	msgs, err := s.tickets.Messages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	t.Messages = msgs
	return t, nil
}

// --- Admin-side ticket management ---

// ListAllTickets returns all tickets for admin view.
func (s *PortalService) ListAllTickets(ctx context.Context, status string, limit, offset int) ([]*domain.Ticket, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.tickets.ListAll(ctx, status, limit, offset)
}

// AdminGetTicket returns a single ticket with its messages (no ownership check).
func (s *PortalService) AdminGetTicket(ctx context.Context, ticketID uuid.UUID) (*domain.Ticket, error) {
	t, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	msgs, err := s.tickets.Messages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	t.Messages = msgs
	return t, nil
}

// AdminReply adds an admin response to a ticket.
func (s *PortalService) AdminReply(ctx context.Context, ticketID, adminID uuid.UUID, body string) error {
	t, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}
	now := s.now()
	msg := &domain.TicketMessage{
		ID:        uuid.New(),
		TicketID:  ticketID,
		Sender:    "admin",
		SenderID:  adminID,
		Body:      body,
		CreatedAt: now,
	}
	if err := s.tickets.AddMessage(ctx, msg); err != nil {
		return err
	}
	t.Status = domain.TicketAnswered
	t.UpdatedAt = now
	return s.tickets.Update(ctx, t)
}

// CloseTicket marks a ticket as closed.
func (s *PortalService) CloseTicket(ctx context.Context, ticketID uuid.UUID) error {
	t, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}
	t.Status = domain.TicketClosed
	t.UpdatedAt = s.now()
	return s.tickets.Update(ctx, t)
}

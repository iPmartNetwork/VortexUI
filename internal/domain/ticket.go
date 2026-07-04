package domain

import (
	"time"

	"github.com/google/uuid"
)

// TicketStatus is the lifecycle state of a support ticket.
type TicketStatus string

const (
	TicketOpen     TicketStatus = "open"
	TicketAnswered TicketStatus = "answered"
	TicketClosed   TicketStatus = "closed"
)

// TicketPriority controls urgency.
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityMedium TicketPriority = "medium"
	TicketPriorityHigh   TicketPriority = "high"
)

// Ticket is a support request submitted by an end-user through the self-service
// portal. Admins can view and reply from the admin panel.
type Ticket struct {
	ID       uuid.UUID      `json:"id"`
	UserID   uuid.UUID      `json:"user_id"`
	Username string         `json:"username,omitempty"`
	Subject  string         `json:"subject"`
	Status   TicketStatus   `json:"status"`
	Priority TicketPriority `json:"priority"`

	Messages []TicketMessage `json:"messages,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TicketMessage is a single message in a ticket thread.
type TicketMessage struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	Sender    string    `json:"sender"`    // "user" | "admin"
	SenderID  uuid.UUID `json:"sender_id"` // user or admin ID
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

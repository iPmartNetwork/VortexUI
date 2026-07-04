package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// TicketRepo implements port.TicketRepository on PostgreSQL.
type TicketRepo struct {
	pool *pgxpool.Pool
}

var _ port.TicketRepository = (*TicketRepo)(nil)

func (r *TicketRepo) Create(ctx context.Context, t *domain.Ticket) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tickets (id, user_id, subject, status, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.ID, t.UserID, t.Subject, t.Status, t.Priority, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TicketRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT t.id, t.user_id, COALESCE(u.username, ''), t.subject, t.status, t.priority, t.created_at, t.updated_at
		FROM tickets t
		LEFT JOIN users u ON u.id = t.user_id
		WHERE t.id = $1`, id)
	var t domain.Ticket
	err := row.Scan(&t.ID, &t.UserID, &t.Username, &t.Subject, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) Update(ctx context.Context, t *domain.Ticket) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tickets SET status = $2, priority = $3, updated_at = $4 WHERE id = $1`,
		t.ID, t.Status, t.Priority, t.UpdatedAt)
	return err
}

func (r *TicketRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Ticket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, subject, status, priority, created_at, updated_at
		FROM tickets WHERE user_id = $1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (r *TicketRepo) ListAll(ctx context.Context, status string, limit, offset int) ([]*domain.Ticket, int, error) {
	var args []interface{}
	q := `SELECT t.id, t.user_id, COALESCE(u.username, ''), t.subject, t.status, t.priority, t.created_at, t.updated_at
		FROM tickets t LEFT JOIN users u ON u.id = t.user_id`
	cq := `SELECT COUNT(*) FROM tickets`
	if status != "" {
		q += ` WHERE t.status = $1`
		cq += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY t.updated_at DESC LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.Username, &t.Subject, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	// Count
	var countArgs []interface{}
	if status != "" {
		countArgs = append(countArgs, status)
	}
	var total int
	_ = r.pool.QueryRow(ctx, cq, countArgs...).Scan(&total)
	return out, total, nil
}

func (r *TicketRepo) AddMessage(ctx context.Context, m *domain.TicketMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ticket_messages (id, ticket_id, sender, sender_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.TicketID, m.Sender, m.SenderID, m.Body, m.CreatedAt)
	return err
}

func (r *TicketRepo) Messages(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, ticket_id, sender, sender_id, body, created_at
		FROM ticket_messages WHERE ticket_id = $1 ORDER BY created_at ASC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TicketMessage
	for rows.Next() {
		var m domain.TicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.Sender, &m.SenderID, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// itoa converts int to string for query building.
func itoa(n int) string {
	_ = time.Now // keep time import used
	return fmt.Sprintf("%d", n)
}

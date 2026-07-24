package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NotificationChannelRepo implements port.NotificationChannelRepository on PostgreSQL.
type NotificationChannelRepo struct {
	pool *pgxpool.Pool
}

var _ port.NotificationChannelRepository = (*NotificationChannelRepo)(nil)

func (r *NotificationChannelRepo) Create(ctx context.Context, ch *domain.NotificationChannel) error {
	configJSON, err := json.Marshal(ch.Config)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO notification_channels (id, name, type, config, scope_type, scope_id, events, template, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		ch.ID, ch.Name, ch.Type, configJSON, ch.ScopeType, nilIfEmpty(ch.ScopeID),
		ch.Events, nilIfEmpty(ch.Template), ch.Enabled, ch.CreatedAt, ch.UpdatedAt)
	return err
}

func (r *NotificationChannelRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.NotificationChannel, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, type, config, scope_type, scope_id, events, template, enabled, created_at, updated_at
		FROM notification_channels
		WHERE id = $1`, id)

	ch, err := scanNotificationChannel(row)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (r *NotificationChannelRepo) Update(ctx context.Context, ch *domain.NotificationChannel) error {
	configJSON, err := json.Marshal(ch.Config)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE notification_channels
		SET name = $2, type = $3, config = $4, scope_type = $5, scope_id = $6,
		    events = $7, template = $8, enabled = $9, updated_at = $10
		WHERE id = $1`,
		ch.ID, ch.Name, ch.Type, configJSON, ch.ScopeType, nilIfEmpty(ch.ScopeID),
		ch.Events, nilIfEmpty(ch.Template), ch.Enabled, ch.UpdatedAt)
	return err
}

func (r *NotificationChannelRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notification_channels WHERE id = $1`, id)
	return err
}

func (r *NotificationChannelRepo) List(ctx context.Context) ([]*domain.NotificationChannel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, type, config, scope_type, scope_id, events, template, enabled, created_at, updated_at
		FROM notification_channels
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectNotificationChannels(rows)
}

func (r *NotificationChannelRepo) ListByScope(ctx context.Context, scopeType, scopeID string) ([]*domain.NotificationChannel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, type, config, scope_type, scope_id, events, template, enabled, created_at, updated_at
		FROM notification_channels
		WHERE enabled = true AND (scope_type = 'global' OR (scope_type = $1 AND scope_id = $2))
		ORDER BY created_at DESC`, scopeType, scopeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectNotificationChannels(rows)
}

// WebhookDeliveryRepo implements port.WebhookDeliveryRepository on PostgreSQL.
type WebhookDeliveryRepo struct {
	pool *pgxpool.Pool
}

var _ port.WebhookDeliveryRepository = (*WebhookDeliveryRepo)(nil)

func (r *WebhookDeliveryRepo) Create(ctx context.Context, d *domain.WebhookDelivery) error {
	payloadJSON, err := json.Marshal(d.Payload)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO webhook_delivery_log (id, channel_id, event_type, payload, status_code, attempts, next_retry, delivered, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.ID, d.ChannelID, d.EventType, payloadJSON, d.StatusCode,
		d.Attempts, d.NextRetry, d.Delivered, d.CreatedAt)
	return err
}

func (r *WebhookDeliveryRepo) ListPending(ctx context.Context) ([]*domain.WebhookDelivery, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, channel_id, event_type, payload, status_code, attempts, next_retry, delivered, created_at
		FROM webhook_delivery_log
		WHERE NOT delivered AND (next_retry IS NULL OR next_retry <= now())
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*domain.WebhookDelivery
	for rows.Next() {
		d, err := scanWebhookDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (r *WebhookDeliveryRepo) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE webhook_delivery_log SET delivered = true WHERE id = $1`, id)
	return err
}

func (r *WebhookDeliveryRepo) IncrementAttempt(ctx context.Context, id uuid.UUID, nextRetry *time.Time, statusCode int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE webhook_delivery_log
		SET attempts = attempts + 1, next_retry = $2, status_code = $3
		WHERE id = $1`, id, nextRetry, statusCode)
	return err
}

// --- helpers ---

func scanNotificationChannel(row pgx.Row) (*domain.NotificationChannel, error) {
	var ch domain.NotificationChannel
	var configJSON []byte
	var scopeID, template *string

	if err := row.Scan(&ch.ID, &ch.Name, &ch.Type, &configJSON, &ch.ScopeType,
		&scopeID, &ch.Events, &template, &ch.Enabled, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
		return nil, err
	}

	if scopeID != nil {
		ch.ScopeID = *scopeID
	}
	if template != nil {
		ch.Template = *template
	}
	if err := json.Unmarshal(configJSON, &ch.Config); err != nil {
		return nil, err
	}
	return &ch, nil
}

func collectNotificationChannels(rows pgx.Rows) ([]*domain.NotificationChannel, error) {
	var channels []*domain.NotificationChannel
	for rows.Next() {
		var ch domain.NotificationChannel
		var configJSON []byte
		var scopeID, template *string

		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Type, &configJSON, &ch.ScopeType,
			&scopeID, &ch.Events, &template, &ch.Enabled, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, err
		}

		if scopeID != nil {
			ch.ScopeID = *scopeID
		}
		if template != nil {
			ch.Template = *template
		}
		if err := json.Unmarshal(configJSON, &ch.Config); err != nil {
			return nil, err
		}
		channels = append(channels, &ch)
	}
	return channels, rows.Err()
}

func scanWebhookDelivery(rows pgx.Rows) (*domain.WebhookDelivery, error) {
	var d domain.WebhookDelivery
	var payloadJSON []byte

	if err := rows.Scan(&d.ID, &d.ChannelID, &d.EventType, &payloadJSON,
		&d.StatusCode, &d.Attempts, &d.NextRetry, &d.Delivered, &d.CreatedAt); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadJSON, &d.Payload); err != nil {
		return nil, err
	}
	return &d, nil
}

// nilIfEmpty returns nil for empty strings, allowing NULL DB storage.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

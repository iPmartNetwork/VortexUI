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

// ClientTemplateRepo implements port.ClientTemplateRepository on PostgreSQL.
type ClientTemplateRepo struct {
	pool *pgxpool.Pool
}

var _ port.ClientTemplateRepository = (*ClientTemplateRepo)(nil)

func (r *ClientTemplateRepo) Create(ctx context.Context, t *domain.ClientTemplate) error {
	rulesJSON, err := json.Marshal(t.RoutingRules)
	if err != nil {
		return err
	}
	dnsJSON, err := json.Marshal(t.DNSSettings)
	if err != nil {
		return err
	}
	outboundsJSON, err := json.Marshal(t.CustomOutbounds)
	if err != nil {
		return err
	}

	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	_, err = r.pool.Exec(ctx, `
		INSERT INTO client_templates (id, name, client_pattern, routing_rules, dns_settings, custom_outbounds, priority, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		t.ID, t.Name, t.ClientPattern, rulesJSON, dnsJSON, outboundsJSON,
		t.Priority, t.Enabled, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *ClientTemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientTemplate, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, client_pattern, routing_rules, dns_settings, custom_outbounds, priority, enabled, created_at, updated_at
		FROM client_templates
		WHERE id = $1`, id)
	return scanClientTemplate(row)
}

func (r *ClientTemplateRepo) Update(ctx context.Context, t *domain.ClientTemplate) error {
	rulesJSON, err := json.Marshal(t.RoutingRules)
	if err != nil {
		return err
	}
	dnsJSON, err := json.Marshal(t.DNSSettings)
	if err != nil {
		return err
	}
	outboundsJSON, err := json.Marshal(t.CustomOutbounds)
	if err != nil {
		return err
	}

	t.UpdatedAt = time.Now()

	_, err = r.pool.Exec(ctx, `
		UPDATE client_templates
		SET name = $2, client_pattern = $3, routing_rules = $4, dns_settings = $5,
		    custom_outbounds = $6, priority = $7, enabled = $8, updated_at = $9
		WHERE id = $1`,
		t.ID, t.Name, t.ClientPattern, rulesJSON, dnsJSON, outboundsJSON,
		t.Priority, t.Enabled, t.UpdatedAt)
	return err
}

func (r *ClientTemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM client_templates WHERE id = $1`, id)
	return err
}

func (r *ClientTemplateRepo) List(ctx context.Context) ([]*domain.ClientTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, client_pattern, routing_rules, dns_settings, custom_outbounds, priority, enabled, created_at, updated_at
		FROM client_templates
		ORDER BY priority DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectClientTemplates(rows)
}

func (r *ClientTemplateRepo) ListEnabled(ctx context.Context) ([]*domain.ClientTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, client_pattern, routing_rules, dns_settings, custom_outbounds, priority, enabled, created_at, updated_at
		FROM client_templates
		WHERE enabled = true
		ORDER BY priority DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectClientTemplates(rows)
}

// --- helpers ---

func scanClientTemplate(row pgx.Row) (*domain.ClientTemplate, error) {
	var t domain.ClientTemplate
	var rulesJSON, dnsJSON, outboundsJSON []byte

	if err := row.Scan(&t.ID, &t.Name, &t.ClientPattern, &rulesJSON, &dnsJSON,
		&outboundsJSON, &t.Priority, &t.Enabled, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rulesJSON, &t.RoutingRules); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dnsJSON, &t.DNSSettings); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(outboundsJSON, &t.CustomOutbounds); err != nil {
		return nil, err
	}
	return &t, nil
}

func collectClientTemplates(rows pgx.Rows) ([]*domain.ClientTemplate, error) {
	var templates []*domain.ClientTemplate
	for rows.Next() {
		var t domain.ClientTemplate
		var rulesJSON, dnsJSON, outboundsJSON []byte

		if err := rows.Scan(&t.ID, &t.Name, &t.ClientPattern, &rulesJSON, &dnsJSON,
			&outboundsJSON, &t.Priority, &t.Enabled, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rulesJSON, &t.RoutingRules); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(dnsJSON, &t.DNSSettings); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(outboundsJSON, &t.CustomOutbounds); err != nil {
			return nil, err
		}
		templates = append(templates, &t)
	}
	return templates, rows.Err()
}

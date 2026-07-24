package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// UserTemplateRepo implements port.UserTemplateRepository on PostgreSQL.
type UserTemplateRepo struct {
	pool *pgxpool.Pool
}

var _ port.UserTemplateRepository = (*UserTemplateRepo)(nil)

func (r *UserTemplateRepo) Create(ctx context.Context, t *domain.UserTemplate) error {
	settingsJSON, err := json.Marshal(t.ProtocolSettings)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO user_templates (id, name, data_limit, expire_duration, device_limit, reset_strategy, note, protocol_settings, groups, allowed_admins, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		t.ID, t.Name, t.DataLimit, t.ExpireDuration, t.DeviceLimit, string(t.ResetStrategy),
		t.Note, settingsJSON, t.Groups, uuidSliceToStrings(t.AllowedAdmins),
		t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *UserTemplateRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserTemplate, error) {
	var t domain.UserTemplate
	var settingsJSON []byte
	var allowedAdmins []string

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, data_limit, expire_duration, device_limit, reset_strategy, note, protocol_settings, groups, allowed_admins, created_at, updated_at
		FROM user_templates WHERE id = $1`, id).
		Scan(&t.ID, &t.Name, &t.DataLimit, &t.ExpireDuration, &t.DeviceLimit, &t.ResetStrategy,
			&t.Note, &settingsJSON, &t.Groups, &allowedAdmins, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(settingsJSON, &t.ProtocolSettings); err != nil {
		return nil, err
	}
	t.AllowedAdmins = stringsToUUIDSlice(allowedAdmins)
	return &t, nil
}

func (r *UserTemplateRepo) GetByName(ctx context.Context, name string) (*domain.UserTemplate, error) {
	var t domain.UserTemplate
	var settingsJSON []byte
	var allowedAdmins []string

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, data_limit, expire_duration, device_limit, reset_strategy, note, protocol_settings, groups, allowed_admins, created_at, updated_at
		FROM user_templates WHERE name = $1`, name).
		Scan(&t.ID, &t.Name, &t.DataLimit, &t.ExpireDuration, &t.DeviceLimit, &t.ResetStrategy,
			&t.Note, &settingsJSON, &t.Groups, &allowedAdmins, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(settingsJSON, &t.ProtocolSettings); err != nil {
		return nil, err
	}
	t.AllowedAdmins = stringsToUUIDSlice(allowedAdmins)
	return &t, nil
}

func (r *UserTemplateRepo) Update(ctx context.Context, t *domain.UserTemplate) error {
	settingsJSON, err := json.Marshal(t.ProtocolSettings)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE user_templates
		SET name = $2, data_limit = $3, expire_duration = $4, device_limit = $5, reset_strategy = $6,
		    note = $7, protocol_settings = $8, groups = $9, allowed_admins = $10, updated_at = $11
		WHERE id = $1`,
		t.ID, t.Name, t.DataLimit, t.ExpireDuration, t.DeviceLimit, string(t.ResetStrategy),
		t.Note, settingsJSON, t.Groups, uuidSliceToStrings(t.AllowedAdmins), t.UpdatedAt)
	return err
}

func (r *UserTemplateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_templates WHERE id = $1`, id)
	return err
}

func (r *UserTemplateRepo) List(ctx context.Context) ([]*domain.UserTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, data_limit, expire_duration, device_limit, reset_strategy, note, protocol_settings, groups, allowed_admins, created_at, updated_at
		FROM user_templates ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.UserTemplate
	for rows.Next() {
		t, err := scanUserTemplate(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

func (r *UserTemplateRepo) ListForAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.UserTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, data_limit, expire_duration, device_limit, reset_strategy, note, protocol_settings, groups, allowed_admins, created_at, updated_at
		FROM user_templates
		WHERE allowed_admins IS NULL OR $1 = ANY(allowed_admins)
		ORDER BY created_at`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.UserTemplate
	for rows.Next() {
		t, err := scanUserTemplate(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// scanUserTemplate scans a single row into a UserTemplate.
func scanUserTemplate(rows pgx.Rows) (*domain.UserTemplate, error) {
	var t domain.UserTemplate
	var settingsJSON []byte
	var allowedAdmins []string

	if err := rows.Scan(&t.ID, &t.Name, &t.DataLimit, &t.ExpireDuration, &t.DeviceLimit, &t.ResetStrategy,
		&t.Note, &settingsJSON, &t.Groups, &allowedAdmins, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(settingsJSON, &t.ProtocolSettings); err != nil {
		return nil, err
	}
	t.AllowedAdmins = stringsToUUIDSlice(allowedAdmins)
	return &t, nil
}

// uuidSliceToStrings converts []uuid.UUID to []string for UUID[] column storage.
// Returns nil when the input is nil (preserving the NULL semantics for allowed_admins).
func uuidSliceToStrings(ids []uuid.UUID) []string {
	if ids == nil {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

// stringsToUUIDSlice converts []string back to []uuid.UUID.
// Returns nil when the input is nil (preserving the NULL semantics).
func stringsToUUIDSlice(strs []string) []uuid.UUID {
	if strs == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		out = append(out, id)
	}
	return out
}

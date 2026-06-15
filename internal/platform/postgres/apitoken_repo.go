package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// APITokenRepo persists API automation tokens.
type APITokenRepo struct{ q *db.Queries }

// ResolvedToken is the auth context a token maps to (its owning admin).
type ResolvedToken struct {
	ID      uuid.UUID
	AdminID uuid.UUID
	Sudo    bool
	RoleID  *uuid.UUID
}

func (r *APITokenRepo) Insert(ctx context.Context, id uuid.UUID, name, hash string, adminID uuid.UUID) error {
	return r.q.InsertAPIToken(ctx, db.InsertAPITokenParams{ID: id, Name: name, TokenHash: hash, AdminID: adminID})
}

func (r *APITokenRepo) Resolve(ctx context.Context, hash string) (ResolvedToken, error) {
	row, err := r.q.ResolveAPIToken(ctx, hash)
	if err != nil {
		return ResolvedToken{}, err
	}
	return ResolvedToken{ID: row.ID, AdminID: row.AdminID, Sudo: row.Sudo, RoleID: uuidToPtr(row.RoleID)}, nil
}

func (r *APITokenRepo) Touch(ctx context.Context, id uuid.UUID) error {
	return r.q.TouchAPIToken(ctx, id)
}

func (r *APITokenRepo) List(ctx context.Context) ([]domain.APIToken, error) {
	rows, err := r.q.ListAPITokens(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.APIToken, len(rows))
	for i, row := range rows {
		t := domain.APIToken{ID: row.ID, Name: row.Name, AdminID: row.AdminID, CreatedAt: row.CreatedAt.Time}
		if row.LastUsedAt.Valid {
			lu := row.LastUsedAt.Time
			t.LastUsedAt = &lu
		}
		out[i] = t
	}
	return out, nil
}

func (r *APITokenRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteAPIToken(ctx, id)
}

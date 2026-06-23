package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// APITokenRepo persists API automation tokens.
type APITokenRepo struct{ q *db.Queries }

// Resolve implements auth.TokenResolver for vtx_ API tokens.
func (r *APITokenRepo) Resolve(ctx context.Context, hash string) (uuid.UUID, bool, *uuid.UUID, uuid.UUID, error) {
	row, err := r.q.ResolveAPIToken(ctx, hash)
	if err != nil {
		return uuid.Nil, false, nil, uuid.Nil, err
	}
	return row.AdminID, row.Sudo, uuidToPtr(row.RoleID), row.ID, nil
}

func (r *APITokenRepo) Insert(ctx context.Context, id uuid.UUID, name, hash string, adminID uuid.UUID) error {
	return r.q.InsertAPIToken(ctx, db.InsertAPITokenParams{ID: id, Name: name, TokenHash: hash, AdminID: adminID})
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

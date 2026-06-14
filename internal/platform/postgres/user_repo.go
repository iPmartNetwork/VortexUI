package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// UserRepo implements port.UserRepository. It needs the pool directly (not just
// Queries) so SetInbounds can run its delete+insert batch in one transaction.
type UserRepo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

var _ port.UserRepository = (*UserRepo)(nil)

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	return r.q.CreateUser(ctx, db.CreateUserParams{
		ID:            u.ID,
		Username:      u.Username,
		Status:        string(u.Status),
		Note:          u.Note,
		DataLimit:     u.DataLimit,
		UsedTraffic:   u.UsedTraffic,
		ExpireAt:      ptrToTS(u.ExpireAt),
		OnHoldExpire:  ptrToInt8(u.OnHoldExpire),
		ResetStrategy: string(u.ResetStrategy),
		LastReset:     ptrToTS(u.LastReset),
		DeviceLimit:   int32(u.DeviceLimit),
		AllowedHwids:  jsonbStrings(u.AllowedHWIDs),
		VmessUuid:     u.Proxies.VMessUUID,
		VlessUuid:     u.Proxies.VLESSUUID,
		TrojanPass:    u.Proxies.TrojanPass,
		SsPassword:    u.Proxies.ShadowsocksP,
		SsMethod:      u.Proxies.SSMethod,
		SubToken:      u.SubToken,
		CreatedAt:     timeToTS(u.CreatedAt),
		UpdatedAt:     timeToTS(u.UpdatedAt),
	})
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return userToDomain(row), nil
}

func (r *UserRepo) GetBySubToken(ctx context.Context, token string) (*domain.User, error) {
	row, err := r.q.GetUserBySubToken(ctx, token)
	if err != nil {
		return nil, mapErr(err)
	}
	return userToDomain(row), nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	return r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:            u.ID,
		Username:      u.Username,
		Status:        string(u.Status),
		Note:          u.Note,
		DataLimit:     u.DataLimit,
		UsedTraffic:   u.UsedTraffic,
		ExpireAt:      ptrToTS(u.ExpireAt),
		OnHoldExpire:  ptrToInt8(u.OnHoldExpire),
		ResetStrategy: string(u.ResetStrategy),
		LastReset:     ptrToTS(u.LastReset),
		DeviceLimit:   int32(u.DeviceLimit),
		AllowedHwids:  jsonbStrings(u.AllowedHWIDs),
		TrojanPass:    u.Proxies.TrojanPass,
		SsPassword:    u.Proxies.ShadowsocksP,
		SsMethod:      u.Proxies.SSMethod,
	})
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *UserRepo) List(ctx context.Context, f port.UserFilter) ([]*domain.User, int, error) {
	limit := int32(f.Limit)
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.ListUsers(ctx, db.ListUsersParams{
		Search: f.Search,
		Status: string(f.Status),
		Off:    int32(f.Offset),
		Lim:    limit,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.q.CountUsers(ctx, db.CountUsersParams{Search: f.Search, Status: string(f.Status)})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = userToDomain(rows[i])
	}
	return out, int(total), nil
}

func (r *UserRepo) AddUsedTraffic(ctx context.Context, id uuid.UUID, delta int64) error {
	return r.q.AddUsedTraffic(ctx, db.AddUsedTrafficParams{ID: id, UsedTraffic: delta})
}

func (r *UserRepo) AddUsedTrafficBatch(ctx context.Context, deltas map[uuid.UUID]int64) error {
	if len(deltas) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(deltas))
	vals := make([]int64, 0, len(deltas))
	for id, d := range deltas {
		ids = append(ids, id)
		vals = append(vals, d)
	}
	return r.q.AddUsedTrafficBatch(ctx, db.AddUsedTrafficBatchParams{Ids: ids, Deltas: vals})
}

// SetInbounds replaces a user's inbound bindings atomically: a stale partial
// state can never be observed because the clear and inserts share one tx.
func (r *UserRepo) SetInbounds(ctx context.Context, userID uuid.UUID, inboundIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful Commit

	qtx := db.New(tx)
	if err := qtx.ClearUserInbounds(ctx, userID); err != nil {
		return err
	}
	for _, inID := range inboundIDs {
		if err := qtx.AddUserInbound(ctx, db.AddUserInboundParams{UserID: userID, InboundID: inID}); err != nil {
			return fmt.Errorf("bind inbound %s: %w", inID, err)
		}
	}
	return tx.Commit(ctx)
}

// BindInbound adds a single user→inbound binding (idempotent), used by failover
// migration to attach a user to a healthy node's matching inbound without
// disturbing its existing bindings.
func (r *UserRepo) BindInbound(ctx context.Context, userID, inboundID uuid.UUID) error {
	return r.q.AddUserInbound(ctx, db.AddUserInboundParams{UserID: userID, InboundID: inboundID})
}

// UnbindInbound removes a single user→inbound binding, used by migrate-back to
// undo a temporary failover binding once the home node recovers.
func (r *UserRepo) UnbindInbound(ctx context.Context, userID, inboundID uuid.UUID) error {
	return r.q.RemoveUserInbound(ctx, db.RemoveUserInboundParams{UserID: userID, InboundID: inboundID})
}

func (r *UserRepo) InboundsFor(ctx context.Context, userID uuid.UUID) ([]domain.Inbound, error) {
	rows, err := r.q.InboundsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Inbound, len(rows))
	for i := range rows {
		out[i] = inboundToDomain(rows[i])
	}
	return out, nil
}

// UsersToLimit returns active users who have exhausted their quota or expired,
// for the enforcement loop. Concrete-only (not in the port) since only the
// enforcer needs it.
func (r *UserRepo) UsersToLimit(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.q.UsersToLimit(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = userToDomain(rows[i])
	}
	return out, nil
}

// UsersToReset returns users whose periodic traffic reset is due. Concrete-only.
func (r *UserRepo) UsersToReset(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.q.UsersToReset(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.User, len(rows))
	for i := range rows {
		out[i] = userToDomain(rows[i])
	}
	return out, nil
}

// UsersByNode assembles the per-inbound user map a node needs for a full config
// sync. It is an extra read on the concrete repo (not part of the port) because
// only the sync path needs it.
func (r *UserRepo) UsersByNode(ctx context.Context, nodeID uuid.UUID) (map[string][]*domain.User, error) {
	rows, err := r.q.UsersByNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]*domain.User)
	for _, row := range rows {
		out[row.Tag] = append(out[row.Tag], userToDomain(row.User))
	}
	return out, nil
}

func userToDomain(u db.User) *domain.User {
	return &domain.User{
		ID:            u.ID,
		Username:      u.Username,
		Status:        domain.UserStatus(u.Status),
		Note:          u.Note,
		DataLimit:     u.DataLimit,
		UsedTraffic:   u.UsedTraffic,
		ExpireAt:      tsToPtr(u.ExpireAt),
		OnHoldExpire:  int8ToPtr(u.OnHoldExpire),
		ResetStrategy: domain.ResetStrategy(u.ResetStrategy),
		LastReset:     tsToPtr(u.LastReset),
		DeviceLimit:   int(u.DeviceLimit),
		AllowedHWIDs:  stringsFromJSONB(u.AllowedHwids),
		Proxies: domain.UserCredentials{
			VMessUUID:    u.VmessUuid,
			VLESSUUID:    u.VlessUuid,
			TrojanPass:   u.TrojanPass,
			ShadowsocksP: u.SsPassword,
			SSMethod:     u.SsMethod,
		},
		SubToken:  u.SubToken,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

// mapErr normalizes pgx's no-rows sentinel to the domain-level ErrNotFound so
// callers don't depend on the driver.
func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

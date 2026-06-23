package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// AdminRepo implements port.AdminRepository.
type AdminRepo struct{ q *db.Queries }

var _ port.AdminRepository = (*AdminRepo)(nil)

func (r *AdminRepo) Create(ctx context.Context, a *domain.Admin) error {
	return r.q.CreateAdmin(ctx, db.CreateAdminParams{
		ID:           a.ID,
		Username:     a.Username,
		PasswordHash: a.PasswordHash,
		Sudo:         a.Sudo,
		RoleID:       ptrToUUID(a.RoleID),
		TotpSecret:   a.TOTPSecret,
		TotpEnabled:  a.TOTPEnabled,
		UserQuota:    int32(a.UserQuota),
		TrafficQuota: a.TrafficQuota,
		CreatedAt:    timeToTS(a.CreatedAt),
	})
}

func (r *AdminRepo) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	row, err := r.q.GetAdminByUsername(ctx, username)
	if err != nil {
		return nil, mapErr(err)
	}
	return adminToDomain(row), nil
}

func (r *AdminRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	row, err := r.q.GetAdminByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return adminToDomain(row), nil
}

func (r *AdminRepo) Update(ctx context.Context, a *domain.Admin) error {
	return r.q.UpdateAdmin(ctx, db.UpdateAdminParams{
		ID:           a.ID,
		PasswordHash: a.PasswordHash,
		Sudo:         a.Sudo,
		RoleID:       ptrToUUID(a.RoleID),
		TotpSecret:   a.TOTPSecret,
		TotpEnabled:  a.TOTPEnabled,
		UserQuota:    int32(a.UserQuota),
		TrafficQuota: a.TrafficQuota,
		LastLogin:    ptrToTS(a.LastLogin),
	})
}

func (r *AdminRepo) List(ctx context.Context) ([]*domain.Admin, error) {
	rows, err := r.q.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Admin, len(rows))
	for i := range rows {
		out[i] = adminToDomain(rows[i])
	}
	return out, nil
}

func (r *AdminRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteAdmin(ctx, id)
}

func (r *AdminRepo) CountSudo(ctx context.Context) (int, error) {
	n, err := r.q.CountSudoAdmins(ctx)
	return int(n), err
}

func (r *AdminRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	perms, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}
	return r.q.CreateRole(ctx, db.CreateRoleParams{ID: role.ID, Name: role.Name, Permissions: perms})
}

func (r *AdminRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	rows, err := r.q.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Role, len(rows))
	for i := range rows {
		var perms []domain.Permission
		_ = json.Unmarshal(rows[i].Permissions, &perms)
		out[i] = &domain.Role{ID: rows[i].ID, Name: rows[i].Name, Permissions: perms}
	}
	return out, nil
}

func (r *AdminRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	row, err := r.q.GetRole(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	var perms []domain.Permission
	_ = json.Unmarshal(row.Permissions, &perms)
	return &domain.Role{ID: row.ID, Name: row.Name, Permissions: perms}, nil
}

func (r *AdminRepo) UpdateRole(ctx context.Context, role *domain.Role) error {
	perms, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}
	return r.q.UpdateRole(ctx, db.UpdateRoleParams{ID: role.ID, Name: role.Name, Permissions: perms})
}

func (r *AdminRepo) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteRole(ctx, id)
}

func adminToDomain(a db.Admin) *domain.Admin {
	return &domain.Admin{
		ID:           a.ID,
		Username:     a.Username,
		PasswordHash: a.PasswordHash,
		Sudo:         a.Sudo,
		RoleID:       uuidToPtr(a.RoleID),
		TOTPSecret:   a.TotpSecret,
		TOTPEnabled:  a.TotpEnabled,
		UserQuota:    int(a.UserQuota),
		TrafficQuota: a.TrafficQuota,
		LastLogin:    tsToPtr(a.LastLogin),
		CreatedAt:    a.CreatedAt.Time,
	}
}

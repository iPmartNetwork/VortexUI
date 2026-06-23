package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// AdminRepo implements port.AdminRepository.
type AdminRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

var _ port.AdminRepository = (*AdminRepo)(nil)

func (r *AdminRepo) Create(ctx context.Context, a *domain.Admin) error {
	return r.q.CreateAdmin(ctx, db.CreateAdminParams{
		ID:                 a.ID,
		Username:           a.Username,
		PasswordHash:       a.PasswordHash,
		Sudo:               a.Sudo,
		RoleID:             ptrToUUID(a.RoleID),
		TotpSecret:         a.TOTPSecret,
		TotpEnabled:        a.TOTPEnabled,
		UserQuota:          int32(a.UserQuota),
		TrafficQuota:       a.TrafficQuota,
		TrafficQuotaMode:   string(a.TrafficQuotaMode),
		ParentAdminID:      ptrToUUID(a.ParentAdminID),
		WalletTrafficBytes: a.WalletTrafficBytes,
		WalletUserCredits:  int32(a.WalletUserCredits),
		WebhookUrl:         a.WebhookURL,
		WebhookSecret:      a.WebhookSecret,
		WebhookEnabled:     a.WebhookEnabled,
		CreatedAt:          timeToTS(a.CreatedAt),
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
		ID:                 a.ID,
		PasswordHash:       a.PasswordHash,
		Sudo:               a.Sudo,
		RoleID:             ptrToUUID(a.RoleID),
		TotpSecret:         a.TOTPSecret,
		TotpEnabled:        a.TOTPEnabled,
		UserQuota:          int32(a.UserQuota),
		TrafficQuota:       a.TrafficQuota,
		TrafficQuotaMode:   string(a.TrafficQuotaMode),
		ParentAdminID:      ptrToUUID(a.ParentAdminID),
		WalletTrafficBytes: a.WalletTrafficBytes,
		WalletUserCredits:  int32(a.WalletUserCredits),
		WebhookUrl:         a.WebhookURL,
		WebhookSecret:      a.WebhookSecret,
		WebhookEnabled:     a.WebhookEnabled,
		LastLogin:          ptrToTS(a.LastLogin),
		PolicyMaxDataLimit:          a.PolicyMaxDataLimit,
		PolicyMaxExpireDays:         int32(a.PolicyMaxExpireDays),
		PolicyAllowBulkDelete:       a.PolicyAllowBulkDelete,
		PolicyAllowBulkCreate:       a.PolicyAllowBulkCreate,
		AutoSuspendEnabled:          a.AutoSuspendEnabled,
		IpViolationSuspendThreshold: int32(a.IPViolationSuspendThreshold),
		SuspendGraceMinutes:         int32(a.SuspendGraceMinutes),
		AllowSubResellers:           a.AllowSubResellers,
		AllowUserBackup:             a.AllowUserBackup,
		ResellerSettings:            resellerSettingsToJSON(a.ResellerSettings),
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

// SetInbounds replaces the inbound allowlist for a reseller admin.
func (r *AdminRepo) SetInbounds(ctx context.Context, adminID uuid.UUID, inboundIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	qtx := db.New(tx)
	if err := qtx.ClearAdminInbounds(ctx, adminID); err != nil {
		return err
	}
	for _, inID := range inboundIDs {
		if err := qtx.AddAdminInbound(ctx, db.AddAdminInboundParams{AdminID: adminID, InboundID: inID}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *AdminRepo) ListInboundIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	return r.q.ListInboundIDsForAdmin(ctx, adminID)
}

func (r *AdminRepo) CountInboundAccess(ctx context.Context, adminID uuid.UUID, inboundIDs []uuid.UUID) (int64, error) {
	if len(inboundIDs) == 0 {
		return 0, nil
	}
	return r.q.CountAdminInboundAccess(ctx, db.CountAdminInboundAccessParams{
		AdminID:    adminID,
		InboundIds: inboundIDs,
	})
}

func (r *AdminRepo) SetPlans(ctx context.Context, adminID uuid.UUID, planIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := db.New(tx)
	if err := qtx.ClearAdminPlans(ctx, adminID); err != nil {
		return err
	}
	for _, pid := range planIDs {
		if err := qtx.AddAdminPlan(ctx, db.AddAdminPlanParams{AdminID: adminID, PlanID: pid}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *AdminRepo) ListPlanIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	return r.q.ListPlanIDsForAdmin(ctx, adminID)
}

func (r *AdminRepo) CountPlanAccess(ctx context.Context, adminID uuid.UUID, planIDs []uuid.UUID) (int64, error) {
	if len(planIDs) == 0 {
		return 0, nil
	}
	return r.q.CountAdminPlanAccess(ctx, db.CountAdminPlanAccessParams{
		AdminID: adminID,
		PlanIds: planIDs,
	})
}

func (r *AdminRepo) SetNodes(ctx context.Context, adminID uuid.UUID, nodeIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := db.New(tx)
	if err := qtx.ClearAdminNodes(ctx, adminID); err != nil {
		return err
	}
	for _, nid := range nodeIDs {
		if err := qtx.AddAdminNode(ctx, db.AddAdminNodeParams{AdminID: adminID, NodeID: nid}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *AdminRepo) ListNodeIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	return r.q.ListNodeIDsForAdmin(ctx, adminID)
}

func (r *AdminRepo) CountNodeAccess(ctx context.Context, adminID uuid.UUID, nodeIDs []uuid.UUID) (int64, error) {
	if len(nodeIDs) == 0 {
		return 0, nil
	}
	return r.q.CountAdminNodeAccess(ctx, db.CountAdminNodeAccessParams{
		AdminID: adminID,
		NodeIds: nodeIDs,
	})
}

func (r *AdminRepo) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Admin, error) {
	rows, err := r.q.ListChildAdmins(ctx, ptrToUUID(&parentID))
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Admin, len(rows))
	for i := range rows {
		out[i] = adminToDomain(rows[i])
	}
	return out, nil
}

func (r *AdminRepo) ApplyWalletDelta(ctx context.Context, adminID uuid.UUID, actorID *uuid.UUID, deltaTraffic int64, deltaUsers int, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := db.New(tx)
	if err := qtx.AdjustAdminWallet(ctx, db.AdjustAdminWalletParams{
		ID: adminID, WalletTrafficBytes: deltaTraffic, WalletUserCredits: int32(deltaUsers),
	}); err != nil {
		return err
	}
	now := timeToTS(time.Now())
	if err := qtx.InsertWalletLedger(ctx, db.InsertWalletLedgerParams{
		ID: uuid.New(), AdminID: adminID, DeltaTraffic: deltaTraffic, DeltaUsers: int32(deltaUsers),
		Reason: reason, ActorAdminID: ptrToUUID(actorID), CreatedAt: now,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *AdminRepo) ListWalletLedger(ctx context.Context, adminID uuid.UUID, limit int) ([]domain.WalletLedgerEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.q.ListWalletLedger(ctx, db.ListWalletLedgerParams{AdminID: adminID, Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.WalletLedgerEntry, len(rows))
	for i, row := range rows {
		out[i] = domain.WalletLedgerEntry{
			ID: row.ID, AdminID: row.AdminID, DeltaTraffic: row.DeltaTraffic,
			DeltaUsers: int(row.DeltaUsers), Reason: row.Reason,
			ActorAdminID: uuidToPtr(row.ActorAdminID), CreatedAt: row.CreatedAt.Time,
		}
	}
	return out, nil
}

func (r *AdminRepo) UpdateWebhook(ctx context.Context, adminID uuid.UUID, url, secret string, enabled bool) error {
	return r.q.UpdateAdminWebhook(ctx, db.UpdateAdminWebhookParams{
		ID: adminID, WebhookUrl: url, WebhookSecret: secret, WebhookEnabled: enabled,
	})
}

func (r *AdminRepo) ListWebhookAdmins(ctx context.Context) ([]domain.Admin, error) {
	rows, err := r.q.ListAdminsWithWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Admin, len(rows))
	for i, row := range rows {
		out[i] = domain.Admin{
			ID: row.ID, Username: row.Username, WebhookURL: row.WebhookUrl,
			WebhookSecret: row.WebhookSecret, WebhookEnabled: row.WebhookEnabled,
		}
	}
	return out, nil
}

func (r *AdminRepo) GetPortalBranding(ctx context.Context, adminID uuid.UUID) (*domain.PortalBranding, error) {
	row, err := r.q.GetPortalBranding(ctx, adminID)
	if err != nil {
		return nil, mapErr(err)
	}
	return portalBrandingToDomain(row), nil
}

func (r *AdminRepo) GetPortalBrandingBySlug(ctx context.Context, slug string) (*domain.PortalBranding, error) {
	row, err := r.q.GetPortalBrandingBySlug(ctx, pgText(slug))
	if err != nil {
		return nil, mapErr(err)
	}
	return portalBrandingToDomain(row), nil
}

func (r *AdminRepo) UpsertPortalBranding(ctx context.Context, b *domain.PortalBranding) error {
	return r.q.UpsertPortalBranding(ctx, db.UpsertPortalBrandingParams{
		AdminID: b.AdminID, PanelTitle: b.PanelTitle, LogoUrl: b.LogoURL,
		AccentColor: b.AccentColor, FooterText: b.FooterText,
		PortalSlug: pgText(b.PortalSlug), CustomDomain: b.CustomDomain,
	})
}

func portalBrandingToDomain(row db.PortalBranding) *domain.PortalBranding {
	slug := ""
	if row.PortalSlug.Valid {
		slug = row.PortalSlug.String
	}
	return &domain.PortalBranding{
		AdminID: row.AdminID, PanelTitle: row.PanelTitle, LogoURL: row.LogoUrl,
		AccentColor: row.AccentColor, FooterText: row.FooterText,
		PortalSlug: slug, CustomDomain: row.CustomDomain,
	}
}

func adminToDomain(a db.Admin) *domain.Admin {
	mode := domain.TrafficQuotaMode(a.TrafficQuotaMode)
	if mode == "" {
		mode = domain.TrafficQuotaAllocated
	}
	return &domain.Admin{
		ID:                 a.ID,
		Username:           a.Username,
		PasswordHash:       a.PasswordHash,
		Sudo:               a.Sudo,
		RoleID:             uuidToPtr(a.RoleID),
		TOTPSecret:         a.TotpSecret,
		TOTPEnabled:        a.TotpEnabled,
		UserQuota:          int(a.UserQuota),
		TrafficQuota:       a.TrafficQuota,
		TrafficQuotaMode:   mode,
		ParentAdminID:      uuidToPtr(a.ParentAdminID),
		WalletTrafficBytes: a.WalletTrafficBytes,
		WalletUserCredits:  int(a.WalletUserCredits),
		WebhookURL:         a.WebhookUrl,
		WebhookSecret:      a.WebhookSecret,
		WebhookEnabled:     a.WebhookEnabled,
		PolicyMaxDataLimit:          a.PolicyMaxDataLimit,
		PolicyMaxExpireDays:         int(a.PolicyMaxExpireDays),
		PolicyAllowBulkDelete:       a.PolicyAllowBulkDelete,
		PolicyAllowBulkCreate:       a.PolicyAllowBulkCreate,
		Suspended:                   a.Suspended,
		SuspendedAt:                 tsToPtr(a.SuspendedAt),
		SuspendReason:               a.SuspendReason,
		AutoSuspendEnabled:          a.AutoSuspendEnabled,
		IPViolationSuspendThreshold: int(a.IpViolationSuspendThreshold),
		SuspendGraceMinutes:         int(a.SuspendGraceMinutes),
		QuotaBreachedAt:             tsToPtr(a.QuotaBreachedAt),
		AllowSubResellers:           a.AllowSubResellers,
		AllowUserBackup:             a.AllowUserBackup,
		ResellerSettings:            jsonToResellerSettings(a.ResellerSettings),
		LastLogin:          tsToPtr(a.LastLogin),
		CreatedAt:          a.CreatedAt.Time,
	}
}

func (r *AdminRepo) Suspend(ctx context.Context, id uuid.UUID, at time.Time, reason string) error {
	return r.q.SuspendAdmin(ctx, db.SuspendAdminParams{ID: id, SuspendedAt: timeToTS(at), SuspendReason: reason})
}

func (r *AdminRepo) Unsuspend(ctx context.Context, id uuid.UUID) error {
	return r.q.UnsuspendAdmin(ctx, id)
}

func (r *AdminRepo) SetQuotaBreachedAt(ctx context.Context, id uuid.UUID, at *time.Time) error {
	return r.q.SetAdminQuotaBreachedAt(ctx, db.SetAdminQuotaBreachedAtParams{ID: id, QuotaBreachedAt: ptrToTS(at)})
}

func (r *AdminRepo) ListResellerCandidates(ctx context.Context) ([]*domain.Admin, error) {
	rows, err := r.q.ListResellerCandidates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Admin, len(rows))
	for i := range rows {
		out[i] = adminToDomain(rows[i])
	}
	return out, nil
}

func resellerSettingsToJSON(m map[string]bool) []byte {
	if len(m) == 0 {
		return []byte("{}")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func jsonToResellerSettings(b []byte) map[string]bool {
	if len(b) == 0 {
		return nil
	}
	var m map[string]bool
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

package api

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
)

type createAdminRequest struct {
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Sudo         bool       `json:"sudo"`
	EnableTOTP   bool       `json:"enable_totp"`
	RoleID       *uuid.UUID `json:"role_id"`
	UserQuota          int        `json:"user_quota"`
	TrafficQuota       int64      `json:"traffic_quota"`
	TrafficQuotaMode   string     `json:"traffic_quota_mode"`
	InboundIDs         []string   `json:"inbound_ids"`
	NodeIDs            []string   `json:"node_ids"`
	PlanIDs            []string   `json:"plan_ids"`
}

// CreateAdmin provisions a new operator. The otpauth URL (if 2FA enabled) is
// returned once.
func (h *Handlers) CreateAdmin(c echo.Context) error {
	var req createAdminRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	mode := domain.TrafficQuotaMode(req.TrafficQuotaMode)
	if mode == "" {
		mode = domain.TrafficQuotaAllocated
	}
	admin, totpURL, err := h.Admins.Create(c.Request().Context(), service.CreateAdminInput{
		Username: req.Username, Password: req.Password, Sudo: req.Sudo, EnableTOTP: req.EnableTOTP,
		RoleID: req.RoleID, UserQuota: req.UserQuota, TrafficQuota: req.TrafficQuota,
		TrafficQuotaMode: mode,
		InboundIDs: mustParseInboundIDs(req.InboundIDs),
		NodeIDs:    mustParseInboundIDs(req.NodeIDs),
		PlanIDs:    mustParseInboundIDs(req.PlanIDs),
	})
	if errors.Is(err, service.ErrAdminExists) {
		return echo.NewHTTPError(http.StatusConflict, "admin already exists")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	resp := echo.Map{"admin": admin}
	if totpURL != "" {
		resp["totp_url"] = totpURL
	}
	return c.JSON(http.StatusCreated, resp)
}

// ListAdmins returns all operators.
func (h *Handlers) ListAdmins(c echo.Context) error {
	admins, err := h.Admins.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if admins == nil {
		admins = []*domain.Admin{}
	}
	return c.JSON(http.StatusOK, echo.Map{"admins": admins})
}

type updateAdminRequest struct {
	Password     string     `json:"password"`
	Sudo         bool       `json:"sudo"`
	UserQuota          int        `json:"user_quota"`
	TrafficQuota       int64      `json:"traffic_quota"`
	TrafficQuotaMode   *string    `json:"traffic_quota_mode"`
	RoleID             *uuid.UUID `json:"role_id"`
	DisableTOTP        bool       `json:"disable_totp"`
	InboundIDs         *[]string  `json:"inbound_ids"`
	NodeIDs            *[]string  `json:"node_ids"`
	PlanIDs            *[]string  `json:"plan_ids"`
	PolicyMaxDataLimit          *int64 `json:"policy_max_data_limit"`
	PolicyMaxExpireDays         *int   `json:"policy_max_expire_days"`
	PolicyAllowBulkDelete       *bool  `json:"policy_allow_bulk_delete"`
	PolicyAllowBulkCreate       *bool  `json:"policy_allow_bulk_create"`
	AutoSuspendEnabled          *bool  `json:"auto_suspend_enabled"`
	IPViolationSuspendThreshold *int   `json:"ip_violation_suspend_threshold"`
	SuspendGraceMinutes         *int   `json:"suspend_grace_minutes"`
	AllowSubResellers           *bool  `json:"allow_sub_resellers"`
	AllowUserBackup             *bool  `json:"allow_user_backup"`
	ResellerSettings            *map[string]bool `json:"reseller_settings"`
}

// UpdateAdmin edits an operator. Demoting the last sudo admin is refused.
func (h *Handlers) UpdateAdmin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateAdminRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	var inboundIDs, nodeIDs, planIDs *[]uuid.UUID
	if req.InboundIDs != nil {
		parsed, perr := parseUUIDs(*req.InboundIDs)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound id")
		}
		inboundIDs = &parsed
	}
	if req.NodeIDs != nil {
		parsed, perr := parseUUIDs(*req.NodeIDs)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid node id")
		}
		nodeIDs = &parsed
	}
	if req.PlanIDs != nil {
		parsed, perr := parseUUIDs(*req.PlanIDs)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid plan id")
		}
		planIDs = &parsed
	}
	var quotaMode *domain.TrafficQuotaMode
	if req.TrafficQuotaMode != nil {
		m := domain.TrafficQuotaMode(*req.TrafficQuotaMode)
		quotaMode = &m
	}
	admin, err := h.Admins.Update(c.Request().Context(), id, service.UpdateAdminInput{
		Password: req.Password, Sudo: req.Sudo, UserQuota: req.UserQuota,
		TrafficQuota: req.TrafficQuota, TrafficQuotaMode: quotaMode, RoleID: req.RoleID, DisableTOTP: req.DisableTOTP,
		InboundIDs: inboundIDs, NodeIDs: nodeIDs, PlanIDs: planIDs,
		PolicyMaxDataLimit: req.PolicyMaxDataLimit, PolicyMaxExpireDays: req.PolicyMaxExpireDays,
		PolicyAllowBulkDelete: req.PolicyAllowBulkDelete, PolicyAllowBulkCreate: req.PolicyAllowBulkCreate,
		AutoSuspendEnabled: req.AutoSuspendEnabled, IPViolationSuspendThreshold: req.IPViolationSuspendThreshold,
		SuspendGraceMinutes: req.SuspendGraceMinutes,
		AllowSubResellers: req.AllowSubResellers, AllowUserBackup: req.AllowUserBackup,
		ResellerSettings: req.ResellerSettings,
	})
	if errors.Is(err, service.ErrLastSudo) {
		return echo.NewHTTPError(http.StatusConflict, "cannot demote the last sudo admin")
	}
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "update failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"admin": admin})
}

type adjustQuotaRequest struct {
	UserQuotaDelta    int   `json:"user_quota_delta"`
	TrafficQuotaDelta int64 `json:"traffic_quota_delta"`
}

// AdjustAdminQuota adds to a reseller's pool limits without opening the edit modal.
func (h *Handlers) AdjustAdminQuota(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req adjustQuotaRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.UserQuotaDelta == 0 && req.TrafficQuotaDelta == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no adjustment specified")
	}
	admin, err := h.Admins.AdjustQuota(c.Request().Context(), id, req.UserQuotaDelta, req.TrafficQuotaDelta)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusOK, echo.Map{"admin": admin})
}

// UnsuspendAdmin re-enables a suspended reseller (sudo only).
func (h *Handlers) UnsuspendAdmin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	admin, err := h.Admins.UnsuspendAdmin(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusOK, echo.Map{"admin": admin})
}

// GetAdminInbounds returns the inbound allowlist for a reseller admin.
func (h *Handlers) GetAdminInbounds(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	admin, err := h.Admins.Get(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	if admin.Sudo {
		return c.JSON(http.StatusOK, echo.Map{"inbound_ids": []string{}})
	}
	ids, err := h.Admins.InboundIDsForAdmin(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	out := make([]string, len(ids))
	for i, inID := range ids {
		out[i] = inID.String()
	}
	return c.JSON(http.StatusOK, echo.Map{"inbound_ids": out})
}

// GetAdminNodes returns the node allowlist for a reseller admin.
func (h *Handlers) GetAdminNodes(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	admin, err := h.Admins.Get(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	if admin.Sudo {
		return c.JSON(http.StatusOK, echo.Map{"node_ids": []string{}})
	}
	ids, err := h.Admins.NodeIDsForAdmin(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	out := uuidSliceToStrings(ids)
	return c.JSON(http.StatusOK, echo.Map{"node_ids": out})
}

// GetAdminPlans returns the plan allowlist for a reseller admin.
func (h *Handlers) GetAdminPlans(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	admin, err := h.Admins.Get(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	if admin.Sudo {
		return c.JSON(http.StatusOK, echo.Map{"plan_ids": []string{}})
	}
	ids, err := h.Admins.PlanIDsForAdmin(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	out := uuidSliceToStrings(ids)
	return c.JSON(http.StatusOK, echo.Map{"plan_ids": out})
}

func uuidSliceToStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func mustParseInboundIDs(ss []string) []uuid.UUID {
	if len(ss) == 0 {
		return nil
	}
	ids, err := parseUUIDs(ss)
	if err != nil {
		return nil
	}
	return ids
}

// DeleteAdmin removes an operator. Self-deletion and removing the last sudo admin
// are refused.
func (h *Handlers) DeleteAdmin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if claims := claimsFrom(c); claims != nil && claims.AdminID == id {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot delete yourself")
	}
	err = h.Admins.Delete(c.Request().Context(), id)
	if errors.Is(err, service.ErrLastSudo) {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete the last sudo admin")
	}
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "admin not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// --- roles ---

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// CreateRole defines a permission bundle.
func (h *Handlers) CreateRole(c echo.Context) error {
	var req createRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	perms := make([]domain.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		perms[i] = domain.Permission(p)
	}
	role, err := h.Admins.CreateRole(c.Request().Context(), req.Name, perms)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusCreated, echo.Map{"role": role})
}

// UpdateRole edits a permission bundle.
func (h *Handlers) UpdateRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req createRoleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	perms := make([]domain.Permission, len(req.Permissions))
	for i, p := range req.Permissions {
		perms[i] = domain.Permission(p)
	}
	role, err := h.Admins.UpdateRole(c.Request().Context(), id, req.Name, perms)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusOK, echo.Map{"role": role})
}

// DeleteRole removes a role.
func (h *Handlers) DeleteRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Admins.DeleteRole(c.Request().Context(), id); errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// ListAudit returns the recent admin action log. Sudo sees all entries; resellers see only their own.
func (h *Handlers) ListAudit(c echo.Context) error {
	if h.Audit == nil {
		return c.JSON(http.StatusOK, echo.Map{"entries": []any{}})
	}
	limit := atoiDefault(c.QueryParam("limit"), 100)
	offset := atoiDefault(c.QueryParam("offset"), 0)
	claims := claimsFrom(c)
	var (
		entries []domain.AuditEntry
		err     error
	)
	if claims != nil && !claims.Sudo {
		entries, err = h.Audit.ListForAdmin(c.Request().Context(), claims.AdminID, limit, offset)
	} else {
		entries, err = h.Audit.List(c.Request().Context(), limit, offset)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "audit list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"entries": entries})
}

// ListRoles returns all roles.
func (h *Handlers) ListRoles(c echo.Context) error {
	roles, err := h.Admins.ListRoles(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if roles == nil {
		roles = []*domain.Role{}
	}
	for _, r := range roles {
		if r != nil && r.Permissions == nil {
			r.Permissions = []domain.Permission{}
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"roles": roles})
}

// GetAccount returns the calling admin and their effective RBAC permissions.
func (h *Handlers) GetAccount(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	admin, err := h.Admins.Get(c.Request().Context(), claims.AdminID)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusUnauthorized, "admin not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	perms, err := h.Admins.Permissions(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "permissions failed")
	}
	out := make([]string, len(perms))
	for i, p := range perms {
		out[i] = string(p)
	}
	if !admin.Sudo {
		admin.ResellerSettings = domain.MergeResellerSettings(admin.ResellerSettings)
	}
	return c.JSON(http.StatusOK, echo.Map{"admin": admin, "permissions": out, "impersonator_id": claims.ImpersonatorID})
}

// GetAccountQuota returns the calling admin's reseller quota usage.
func (h *Handlers) GetAccountQuota(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	usage, err := h.Admins.QuotaUsage(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "quota failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"usage": usage})
}

// GetResellerDashboard returns the reseller home summary for the calling admin.
func (h *Handlers) GetResellerDashboard(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	dash, err := h.Admins.ResellerDashboard(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "dashboard failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"dashboard": dash})
}

// ExportAccountUsers returns a CSV of users for accounting (scoped to the reseller).
func (h *Handlers) ExportAccountUsers(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	if h.Repo == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "export unavailable")
	}
	f := port.UserFilter{Limit: 100000}
	if !claims.Sudo {
		f.AdminID = &claims.AdminID
	}
	users, _, err := h.Repo.List(c.Request().Context(), f)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "export failed")
	}

	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="reseller-users.csv"`)
	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{"username", "status", "data_limit_bytes", "used_traffic_bytes", "remaining_bytes", "expire_at", "created_at", "note"})
	for _, u := range users {
		remaining := ""
		if u.DataLimit > 0 {
			remaining = strconv.FormatInt(u.DataLimit-u.UsedTraffic, 10)
		}
		expire := ""
		if u.ExpireAt != nil {
			expire = u.ExpireAt.UTC().Format(time.RFC3339)
		}
		_ = w.Write([]string{
			u.Username,
			string(u.Status),
			strconv.FormatInt(u.DataLimit, 10),
			strconv.FormatInt(u.UsedTraffic, 10),
			remaining,
			expire,
			u.CreatedAt.UTC().Format(time.RFC3339),
			u.Note,
		})
	}
	w.Flush()
	return nil
}

// ExportAccountUsersBackup returns a JSON snapshot of users owned by the caller (reseller backup).
func (h *Handlers) ExportAccountUsersBackup(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	admin, err := h.Admins.Get(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	if !claims.Sudo && !admin.AllowUserBackup {
		return echo.NewHTTPError(http.StatusForbidden, "user backup is disabled for this account")
	}
	if h.Repo == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "export unavailable")
	}
	f := port.UserFilter{Limit: 100_000}
	if !claims.Sudo {
		f.AdminID = &claims.AdminID
	}
	users, _, err := h.Repo.List(c.Request().Context(), f)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "export failed")
	}
	c.Response().Header().Set("Content-Disposition", `attachment; filename="my-users-backup.json"`)
	return c.JSON(http.StatusOK, echo.Map{
		"version":     1,
		"exported_at": time.Now().UTC(),
		"admin_id":    claims.AdminID,
		"users":       users,
	})
}

// ListResellerQuotaUsage returns quota usage for all resellers (sudo).
func (h *Handlers) ListResellerQuotaUsage(c echo.Context) error {
	usage, err := h.Admins.ListResellerQuotaUsage(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if usage == nil {
		usage = []*domain.AdminQuotaUsage{}
	}
	return c.JSON(http.StatusOK, echo.Map{"usage": usage})
}

// GetAdminQuotaUsage returns quota usage for one admin (sudo or self).
func (h *Handlers) GetAdminQuotaUsage(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if !claims.Sudo && claims.AdminID != id {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	usage, err := h.Admins.QuotaUsage(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "quota failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"usage": usage})
}

// GetAdminWallet returns prepaid wallet balance and ledger for one admin (sudo or self).
func (h *Handlers) GetAdminWallet(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if !claims.Sudo && claims.AdminID != id {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	wallet, ledger, err := h.Admins.WalletView(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "wallet failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"wallet": wallet, "ledger": ledger})
}

// ChangePassword updates the calling admin's own password.
func (h *Handlers) ChangePassword(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var req struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	err := h.Admins.ChangePassword(c.Request().Context(), claims.AdminID, req.Current, req.New)
	if errors.Is(err, service.ErrWrongPassword) {
		return echo.NewHTTPError(http.StatusBadRequest, "current password is incorrect")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

// --- 2FA self-enrollment (acts on the caller's own account, from the token) ---

type totpCodeRequest struct {
	Code string `json:"code"`
}

// SetupTOTP begins enrollment for the calling admin and returns the otpauth URL.
func (h *Handlers) SetupTOTP(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	secret, url, err := h.Admins.BeginTOTP(c.Request().Context(), claims.AdminID)
	if errors.Is(err, service.ErrTOTPAlreadyEnabled) {
		return echo.NewHTTPError(http.StatusConflict, "2fa already enabled")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "setup failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"secret": secret, "url": url})
}

// ConfirmTOTP activates 2FA after verifying a code from the caller's app.
func (h *Handlers) ConfirmTOTP(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var req totpCodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	err := h.Admins.ConfirmTOTP(c.Request().Context(), claims.AdminID, req.Code)
	if errors.Is(err, service.ErrTOTPInvalidCode) || errors.Is(err, service.ErrTOTPNotEnrolled) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid or no pending code")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "confirm failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"enabled": true})
}

// DisableTOTP turns 2FA off for the caller after verifying a current code.
func (h *Handlers) DisableTOTP(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var req totpCodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	err := h.Admins.DisableTOTP(c.Request().Context(), claims.AdminID, req.Code)
	if errors.Is(err, service.ErrTOTPInvalidCode) || errors.Is(err, service.ErrTOTPNotEnabled) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid code or 2fa not enabled")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "disable failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"enabled": false})
}

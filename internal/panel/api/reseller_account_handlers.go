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
	"github.com/vortexui/vortexui/internal/panel/service"
)

// GetAccountWallet returns prepaid wallet balance and ledger.
func (h *Handlers) GetAccountWallet(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	wallet, ledger, err := h.Admins.WalletView(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "wallet failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"wallet": wallet, "ledger": ledger})
}

// ExportAccountWallet returns a CSV statement of wallet ledger entries (invoice-style).
func (h *Handlers) ExportAccountWallet(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	wallet, ledger, err := h.Admins.WalletView(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "wallet failed")
	}
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="wallet-statement.csv"`)
	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{"exported_at", time.Now().UTC().Format(time.RFC3339)})
	_ = w.Write([]string{"traffic_balance_bytes", strconv.FormatInt(wallet.TrafficBytes, 10)})
	_ = w.Write([]string{"user_credits_balance", strconv.Itoa(wallet.UserCredits)})
	_ = w.Write(nil)
	_ = w.Write([]string{"created_at", "delta_traffic_bytes", "delta_users", "reason"})
	for _, e := range ledger {
		_ = w.Write([]string{
			e.CreatedAt.UTC().Format(time.RFC3339),
			strconv.FormatInt(e.DeltaTraffic, 10),
			strconv.Itoa(e.DeltaUsers),
			e.Reason,
		})
	}
	w.Flush()
	return nil
}

type topUpWalletRequest struct {
	TrafficBytes int64  `json:"traffic_bytes"`
	UserCredits  int    `json:"user_credits"`
	Reason       string `json:"reason"`
}

// TopUpAdminWallet credits a reseller wallet (sudo or parent).
func (h *Handlers) TopUpAdminWallet(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req topUpWalletRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	reason := req.Reason
	if reason == "" {
		reason = "top-up"
	}
	if err := h.Admins.TopUpWallet(c.Request().Context(), claims.AdminID, id, req.TrafficBytes, req.UserCredits, reason); err != nil {
		if errors.Is(err, service.ErrNotParentAdmin) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ListSubAdmins returns child resellers for the calling admin.
func (h *Handlers) ListSubAdmins(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	admins, err := h.Admins.ListSubAdmins(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"admins": admins})
}

type createSubAdminRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	RoleID           string `json:"role_id"`
	UserQuota        int    `json:"user_quota"`
	TrafficQuota     int64  `json:"traffic_quota"`
	TrafficQuotaMode string `json:"traffic_quota_mode"`
}

// CreateSubAdmin provisions a child reseller under the caller.
func (h *Handlers) CreateSubAdmin(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var req createSubAdminRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role_id")
	}
	mode := domain.TrafficQuotaMode(req.TrafficQuotaMode)
	if mode == "" {
		mode = domain.TrafficQuotaAllocated
	}
	admin, err := h.Admins.CreateSubAdmin(c.Request().Context(), service.CreateSubAdminInput{
		ParentID: claims.AdminID, Username: req.Username, Password: req.Password,
		RoleID: roleID, UserQuota: req.UserQuota, TrafficQuota: req.TrafficQuota, TrafficQuotaMode: mode,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"admin": admin})
}

// GetAccountBranding returns portal branding for the reseller.
func (h *Handlers) GetAccountBranding(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	b, err := h.Admins.GetPortalBranding(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "branding failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"branding": b})
}

// UpdateAccountBranding saves portal branding.
func (h *Handlers) UpdateAccountBranding(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var b domain.PortalBranding
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	b.AdminID = claims.AdminID
	if err := h.Admins.SavePortalBranding(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"branding": b})
}

// PublicPortalBranding resolves branding by slug (public).
func (h *Handlers) PublicPortalBranding(c echo.Context) error {
	slug := c.QueryParam("slug")
	if slug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "slug required")
	}
	b, err := h.Admins.GetPortalBrandingBySlug(c.Request().Context(), slug)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "branding failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"branding": b})
}

type webhookConfigRequest struct {
	URL     string `json:"url"`
	Secret  string `json:"secret"`
	Enabled bool   `json:"enabled"`
}

// GetAccountWebhook returns webhook settings.
func (h *Handlers) GetAccountWebhook(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	admin, err := h.Admins.GetWebhookConfig(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"url": admin.WebhookURL, "enabled": admin.WebhookEnabled, "has_secret": admin.WebhookSecret != "",
	})
}

// UpdateAccountWebhook saves webhook settings.
func (h *Handlers) UpdateAccountWebhook(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var req webhookConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Admins.UpdateWebhookConfig(c.Request().Context(), claims.AdminID, req.URL, req.Secret, req.Enabled); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ImpersonateAdmin issues a short-lived token as another admin (sudo).
func (h *Handlers) ImpersonateAdmin(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil || !claims.Sudo {
		return echo.NewHTTPError(http.StatusForbidden, "sudo required")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	token, err := h.Admins.Impersonate(c.Request().Context(), claims.AdminID, id, h.Issuer)
	if errors.Is(err, service.ErrCannotImpersonateSudo) {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"token": token, "impersonating": id})
}

// StopImpersonation restores the impersonator's own session.
func (h *Handlers) StopImpersonation(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil || claims.ImpersonatorID == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "not impersonating")
	}
	token, err := h.Admins.StopImpersonation(c.Request().Context(), *claims.ImpersonatorID, h.Issuer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

package api

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// WalletBillingHandlers serves wallet package and deposit APIs.
type WalletBillingHandlers struct {
	Svc *service.WalletBillingService
}

func (h *WalletBillingHandlers) ListWalletPackages(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	pkgs, err := h.Svc.ListPackages(c.Request().Context(), false)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"packages": pkgs})
}

func (h *WalletBillingHandlers) CreateWalletPackage(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	var req struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		TrafficBytes int64    `json:"traffic_bytes"`
		UserCredits  int      `json:"user_credits"`
		PriceAmount  int64    `json:"price_amount"`
		Currency     string   `json:"currency"`
		Methods      []string `json:"methods"`
		Enabled      bool     `json:"enabled"`
		SortOrder    int      `json:"sort_order"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pkg, err := h.Svc.CreatePackage(c.Request().Context(), service.CreateWalletPackageInput{
		Name: req.Name, Description: req.Description, TrafficBytes: req.TrafficBytes,
		UserCredits: req.UserCredits, PriceAmount: req.PriceAmount, Currency: req.Currency,
		Methods: req.Methods, Enabled: req.Enabled, SortOrder: req.SortOrder,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"package": pkg})
}

func (h *WalletBillingHandlers) UpdateWalletPackage(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pkg, err := h.Svc.GetPackage(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if err := c.Bind(pkg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pkg.ID = id
	if err := h.Svc.UpdatePackage(c.Request().Context(), pkg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"package": pkg})
}

func (h *WalletBillingHandlers) DeleteWalletPackage(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Svc.DeletePackage(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *WalletBillingHandlers) GetBillingSettings(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	s, err := h.Svc.GetBillingSettings(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "settings failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"settings": s})
}

func (h *WalletBillingHandlers) UpdateBillingSettings(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	var s domain.BillingSettings
	if err := c.Bind(&s); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Svc.UpdateBillingSettings(c.Request().Context(), &s); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "update failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"settings": s})
}

func (h *WalletBillingHandlers) ListWalletDeposits(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	var status *domain.WalletDepositStatus
	if s := c.QueryParam("status"); s != "" {
		st := domain.WalletDepositStatus(s)
		status = &st
	}
	deposits, err := h.Svc.ListDeposits(c.Request().Context(), nil, status, 200)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"deposits": deposits})
}

func (h *WalletBillingHandlers) ReviewWalletDeposit(c echo.Context) error {
	if h.Svc == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	approve := strings.ToLower(req.Action) == "approve"
	deposit, err := h.Svc.ReviewDeposit(c.Request().Context(), claims.AdminID, id, approve, req.Note)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"deposit": deposit})
}

func (h *Handlers) ListAccountWalletPackages(c echo.Context) error {
	if h.WalletBilling == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	pkgs, err := h.WalletBilling.ListPackages(c.Request().Context(), true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"packages": pkgs})
}

func (h *Handlers) GetAccountPaymentInfo(c echo.Context) error {
	if h.WalletBilling == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	s, err := h.WalletBilling.GetBillingSettings(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "settings failed")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"settings": s,
		"zarinpal_enabled": h.ZarinPal != nil,
		"crypto_enabled":   h.NowPayments != nil,
	})
}

func (h *Handlers) ListAccountWalletDeposits(c echo.Context) error {
	if h.WalletBilling == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	adminID := claims.AdminID
	deposits, err := h.WalletBilling.ListDeposits(c.Request().Context(), &adminID, nil, 50)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"deposits": deposits})
}

func (h *Handlers) InitAccountWalletDeposit(c echo.Context) error {
	if h.WalletBilling == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	var req struct {
		PackageID    string `json:"package_id"`
		Method       string `json:"method"`
		TxID         string `json:"tx_id"`
		ProofImage   string `json:"proof_image"`
		ResellerNote string `json:"reseller_note"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pkgID, err := uuid.Parse(req.PackageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid package_id")
	}
	callbackBase := c.Scheme() + "://" + c.Request().Host
	res, err := h.WalletBilling.InitDeposit(c.Request().Context(), service.InitWalletDepositInput{
		AdminID:      claims.AdminID,
		PackageID:    pkgID,
		Method:       domain.WalletDepositMethod(req.Method),
		TxID:         req.TxID,
		ProofImage:   req.ProofImage,
		ResellerNote: req.ResellerNote,
		CallbackBase: callbackBase,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	out := echo.Map{"deposit": res.Deposit}
	if res.RedirectURL != "" {
		out["redirect_url"] = res.RedirectURL
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handlers) WalletDepositCallback(c echo.Context) error {
	if h.WalletBilling == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "billing unavailable")
	}
	depositID, err := uuid.Parse(c.QueryParam("deposit_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid deposit_id")
	}
	deposit, err := h.WalletBilling.GetDeposit(c.Request().Context(), depositID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "deposit not found")
	}
	if deposit.Status != domain.WalletDepositPending {
		return c.Redirect(http.StatusFound, "/payment/result?status="+string(deposit.Status))
	}
	if err := h.WalletBilling.CompleteOnlineDeposit(c.Request().Context(), depositID, deposit.GatewayID, deposit.Amount); err != nil {
		return c.Redirect(http.StatusFound, "/payment/result?status=failed")
	}
	return c.Redirect(http.StatusFound, "/payment/result?status=success&type=wallet")
}

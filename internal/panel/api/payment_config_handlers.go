package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// PaymentConfigHandlers exposes per-reseller payment config endpoints.
type PaymentConfigHandlers struct {
	Svc    *service.ResellerPaymentService
	Admins *service.AdminService
}

// GetPaymentConfig returns the calling admin's per-reseller payment configuration.
func (h *PaymentConfigHandlers) GetPaymentConfig(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	// Gate: billing feature must be enabled (sudo always allowed).
	if !claims.Sudo {
		if err := h.requireBilling(c); err != nil {
			return err
		}
	}

	cfg, err := h.Svc.GetPaymentConfig(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load payment config")
	}
	return c.JSON(http.StatusOK, echo.Map{"payment_config": cfg})
}

// SavePaymentConfig persists the calling admin's per-reseller payment configuration.
func (h *PaymentConfigHandlers) SavePaymentConfig(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	// Gate: billing feature must be enabled (sudo always allowed).
	if !claims.Sudo {
		if err := h.requireBilling(c); err != nil {
			return err
		}
	}

	var cfg domain.ResellerPaymentConfig
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Svc.SavePaymentConfig(c.Request().Context(), claims.AdminID, &cfg); err != nil {
		if errors.Is(err, service.ErrInvalidPaymentMethod) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, service.ErrInvalidCardNumber) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save payment config")
	}

	// Return the saved config for confirmation.
	saved, _ := h.Svc.GetPaymentConfig(c.Request().Context(), claims.AdminID)
	return c.JSON(http.StatusOK, echo.Map{"payment_config": saved})
}

// requireBilling checks that the calling admin has the "billing" reseller setting enabled.
func (h *PaymentConfigHandlers) requireBilling(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	admin, err := h.Admins.Get(c.Request().Context(), claims.AdminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "admin lookup failed")
	}
	if !domain.ResellerSettingEnabled(admin.ResellerSettings, domain.ResellerSettingBilling) {
		return echo.NewHTTPError(http.StatusForbidden, "billing feature not enabled for this account")
	}
	return nil
}

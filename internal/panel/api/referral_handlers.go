package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// ReferralHandlers serves invite/referral system endpoints.
type ReferralHandlers struct {
	Referral *service.ReferralService
}

// GetReferralConfig returns the referral program settings.
func (h *ReferralHandlers) GetReferralConfig(c echo.Context) error {
	cfg, err := h.Referral.GetConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

type updateReferralConfigRequest struct {
	Enabled      bool   `json:"enabled"`
	RewardType   string `json:"reward_type"`
	RewardAmount int64  `json:"reward_amount"`
	MaxReferrals int    `json:"max_referrals"`
	RequirePaid  bool   `json:"require_paid"`
}

// UpdateReferralConfig saves the referral config.
func (h *ReferralHandlers) UpdateReferralConfig(c echo.Context) error {
	var req updateReferralConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	cfg := &domain.ReferralConfig{
		Enabled:      req.Enabled,
		RewardType:   domain.RewardType(req.RewardType),
		RewardAmount: req.RewardAmount,
		MaxReferrals: req.MaxReferrals,
		RequirePaid:  req.RequirePaid,
	}
	if err := h.Referral.UpdateConfig(c.Request().Context(), cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

// GetMyCode returns the calling user's referral code (portal endpoint).
func (h *ReferralHandlers) GetMyCode(c echo.Context) error {
	userID := portalUserID(c)
	code, err := h.Referral.GetOrCreateCode(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"code": code})
}

type applyReferralRequest struct {
	Code string `json:"code"`
}

// ApplyReferral processes a referral code for the portal user.
func (h *ReferralHandlers) ApplyReferral(c echo.Context) error {
	var req applyReferralRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID := portalUserID(c)
	if err := h.Referral.ApplyReferral(c.Request().Context(), req.Code, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "reward applied"})
}

// ListReferralCodes returns all codes (admin).
func (h *ReferralHandlers) ListReferralCodes(c echo.Context) error {
	codes, total, err := h.Referral.ListCodes(c.Request().Context(), 50, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"codes": codes, "total": total})
}

// ListReferralEvents returns referral events.
func (h *ReferralHandlers) ListReferralEvents(c echo.Context) error {
	var userID *uuid.UUID
	if uid := c.QueryParam("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err == nil {
			userID = &id
		}
	}
	events, err := h.Referral.ListEvents(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"events": events})
}

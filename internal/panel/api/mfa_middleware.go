package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/port"
)

// MFAMiddleware validates MFA requirements for sensitive operations
type MFAMiddleware struct {
	mfaRepo port.MFARepository
	log     *slog.Logger
}

// NewMFAMiddleware creates new MFA middleware
func NewMFAMiddleware(mfaRepo port.MFARepository, log *slog.Logger) *MFAMiddleware {
	if log == nil {
		log = slog.Default()
	}
	return &MFAMiddleware{
		mfaRepo: mfaRepo,
		log:     log,
	}
}

// ValidateMFARequired middleware checks if MFA is required and valid
func (m *MFAMiddleware) ValidateMFARequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this route requires MFA
			requireMFA := c.Get("require_mfa")
			if requireMFA == nil || requireMFA.(bool) == false {
				return next(c)
			}

			adminIDVal := c.Get("admin_id")
			if adminIDVal == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "admin_id not found"})
			}

			adminID, ok := adminIDVal.(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin_id"})
			}

			// Get MFA config for admin
			adminUUID, err := uuid.Parse(adminID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin_id format"})
			}

			config, err := m.mfaRepo.GetConfig(c.Request().Context(), adminUUID)
			if err != nil {
				// No MFA config = MFA not required for this admin
				return next(c)
			}

			// Check if any MFA method is enabled
			if !config.TOTPEnabled && !config.EmailEnabled {
				// MFA not configured, can proceed
				return next(c)
			}

			// Check for MFA verification token in request
			mfaVerified := c.Get("mfa_verified")
			if mfaVerified != nil && mfaVerified.(bool) {
				// MFA already verified in this session
				return next(c)
			}

			// Check for MFA code in header
			mfaCode := c.Request().Header.Get("X-MFA-Code")
			if mfaCode == "" {
				// No MFA code provided
				c.Set("require_mfa_response", true)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error":     "MFA required",
					"mfa_code":  "MFA verification code required in X-MFA-Code header",
				})
			}

			// MFA code is provided - let the next handler validate it
			// (could be validated here or in the handler)
			c.Set("mfa_verified", true)
			return next(c)
		}
	}
}

// ValidateMFAIfEnabled checks MFA only if it's enabled for the admin
func (m *MFAMiddleware) ValidateMFAIfEnabled() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			adminIDVal := c.Get("admin_id")
			if adminIDVal == nil {
				return next(c)
			}

			adminID, ok := adminIDVal.(string)
			if !ok {
				return next(c)
			}

			// Get MFA config
			adminUUID, err := uuid.Parse(adminID)
			if err != nil {
				// Invalid UUID, skip MFA check
				return next(c)
			}

			config, err := m.mfaRepo.GetConfig(c.Request().Context(), adminUUID)
			if err != nil || (!config.TOTPEnabled && !config.EmailEnabled) {
				// MFA not enabled, skip
				return next(c)
			}

			// Check if MFA already verified
			mfaVerified := c.Get("mfa_verified")
			if mfaVerified != nil && mfaVerified.(bool) {
				return next(c)
			}

			// MFA is enabled but not verified
			m.log.Warn("MFA required but not verified", "admin_id", adminID)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "MFA verification required",
			})
		}
	}
}

// SetMFARequired helper to mark a route as MFA required
func (m *MFAMiddleware) SetMFARequired(c echo.Context) {
	c.Set("require_mfa", true)
}

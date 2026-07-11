package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
)

// MFAHandlers handles MFA-related requests
type MFAHandlers struct {
	totpService *service.TOTPService
	mfaRepo     *postgres.MFARepository
	log         *slog.Logger
}

// NewMFAHandlers creates new MFA handlers
func NewMFAHandlers(totpService *service.TOTPService, mfaRepo *postgres.MFARepository, log *slog.Logger) *MFAHandlers {
	if log == nil {
		log = slog.Default()
	}
	return &MFAHandlers{
		totpService: totpService,
		mfaRepo:     mfaRepo,
		log:         log,
	}
}

// SetupTOTPRequest initiates TOTP setup
type SetupTOTPRequest struct {
	Issuer      string `json:"issuer"`
	AccountName string `json:"account_name"`
}

// SetupTOTPResponse returns TOTP secret and QR code
type SetupTOTPResponse struct {
	Secret    string `json:"secret"`
	QRCode    string `json:"qr_code"`
	ExpiresIn int    `json:"expires_in"` // Seconds until setup expires
}

// SetupTOTP initiates TOTP setup for an admin
func (h *MFAHandlers) SetupTOTP(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	var req SetupTOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Issuer == "" {
		req.Issuer = "VortexUI"
	}
	if req.AccountName == "" {
		req.AccountName = adminID.String()
	}

	secret, qrCode, err := h.totpService.GenerateSecret(req.Issuer, req.AccountName)
	if err != nil {
		h.log.Error("failed to generate TOTP secret", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate secret"})
	}

	// Return setup response (secret not yet verified)
	return c.JSON(http.StatusOK, SetupTOTPResponse{
		Secret:    secret,
		QRCode:    qrCode,
		ExpiresIn: 600, // 10 minutes
	})
}

// VerifyTOTPRequest verifies a TOTP setup
type VerifyTOTPRequest struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

// VerifyTOTP verifies and confirms TOTP setup
func (h *MFAHandlers) VerifyTOTP(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	var req VerifyTOTPRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Verify the TOTP code
	valid, err := h.totpService.ValidateCode(req.Secret, req.Code)
	if err != nil || !valid {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid TOTP code"})
	}

	// Save MFA config
	mfaConfig := &domain.MFAConfig{
		ID:           uuid.New(),
		AdminID:      adminID,
		TOTPEnabled:  true,
		EmailEnabled: false,
		BackupCodes:  []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.mfaRepo.SaveConfig(c.Request().Context(), mfaConfig); err != nil {
		h.log.Error("failed to save MFA config", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save config"})
	}

	h.log.Info("TOTP enabled", "admin_id", adminID)
	return c.JSON(http.StatusOK, map[string]string{"message": "TOTP enabled successfully"})
}

// GenerateBackupCodesResponse returns backup codes
type GenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
	Message     string   `json:"message"`
}

// GenerateBackupCodes generates new backup codes
func (h *MFAHandlers) GenerateBackupCodes(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	// Generate backup codes
	codes, err := h.totpService.GenerateBackupCodes(10)
	if err != nil {
		h.log.Error("failed to generate backup codes", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate codes"})
	}

	// Update MFA config
	config, err := h.mfaRepo.GetConfig(c.Request().Context(), adminID)
	if err != nil {
		h.log.Error("failed to get MFA config", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get config"})
	}

	config.BackupCodes = codes
	config.UpdatedAt = time.Now()

	if err := h.mfaRepo.UpdateConfig(c.Request().Context(), config); err != nil {
		h.log.Error("failed to update MFA config", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update config"})
	}

	return c.JSON(http.StatusOK, GenerateBackupCodesResponse{
		BackupCodes: codes,
		Message:     "Backup codes generated successfully. Store them in a safe place.",
	})
}

// DisableMFARequest disables MFA for an admin
type DisableMFARequest struct {
	Password string `json:"password"` // Require password for confirmation
}

// DisableMFA disables MFA
func (h *MFAHandlers) DisableMFA(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	var req DisableMFARequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Get current MFA config
	config, err := h.mfaRepo.GetConfig(c.Request().Context(), adminID)
	if err != nil {
		h.log.Error("failed to get MFA config", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get config"})
	}

	// Disable MFA
	config.TOTPEnabled = false
	config.EmailEnabled = false
	config.BackupCodes = []string{}
	config.UpdatedAt = time.Now()

	if err := h.mfaRepo.UpdateConfig(c.Request().Context(), config); err != nil {
		h.log.Error("failed to disable MFA", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to disable MFA"})
	}

	h.log.Info("MFA disabled", "admin_id", adminID)
	return c.JSON(http.StatusOK, map[string]string{"message": "MFA disabled successfully"})
}

// GetMFAStatusResponse returns MFA status
type GetMFAStatusResponse struct {
	TOTPEnabled  bool   `json:"totp_enabled"`
	EmailEnabled bool   `json:"email_enabled"`
	HasBackupCodes bool  `json:"has_backup_codes"`
}

// GetMFAStatus retrieves current MFA status
func (h *MFAHandlers) GetMFAStatus(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	config, err := h.mfaRepo.GetConfig(c.Request().Context(), adminID)
	if err != nil {
		// If not found, return default (no MFA)
		return c.JSON(http.StatusOK, GetMFAStatusResponse{
			TOTPEnabled:    false,
			EmailEnabled:   false,
			HasBackupCodes: false,
		})
	}

	return c.JSON(http.StatusOK, GetMFAStatusResponse{
		TOTPEnabled:    config.TOTPEnabled,
		EmailEnabled:   config.EmailEnabled,
		HasBackupCodes: len(config.BackupCodes) > 0,
	})
}

// Helper function to extract admin ID from context
func extractAdminID(c echo.Context) (uuid.UUID, error) {
	adminIDStr := c.Get("admin_id")
	if adminIDStr == nil {
		return uuid.Nil, domain.ErrUnauthorized
	}

	adminID, err := uuid.Parse(adminIDStr.(string))
	if err != nil {
		return uuid.Nil, err
	}

	return adminID, nil
}

package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
)

// PasswordHandlers handles password-related requests
type PasswordHandlers struct {
	passwordService    *service.PasswordPolicyService
	passwordRepo       *postgres.PasswordPolicyRepository
	log                *slog.Logger
}

// NewPasswordHandlers creates new password handlers
func NewPasswordHandlers(passwordService *service.PasswordPolicyService, passwordRepo *postgres.PasswordPolicyRepository, log *slog.Logger) *PasswordHandlers {
	if log == nil {
		log = slog.Default()
	}
	return &PasswordHandlers{
		passwordService: passwordService,
		passwordRepo:    passwordRepo,
		log:             log,
	}
}

// ChangePasswordRequest changes admin password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePasswordResponse confirms password change
type ChangePasswordResponse struct {
	Message   string `json:"message"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// ChangePassword allows admin to change their password
func (h *PasswordHandlers) ChangePassword(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "new password cannot be empty"})
	}

	// Get password policy
	policy, err := h.passwordRepo.GetPolicy(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get password policy", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get policy"})
	}

	// Validate new password
	valid, errors, err := h.passwordService.ValidatePassword(req.NewPassword, policy)
	if !valid {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":  "password does not meet policy requirements",
			"errors": errors,
		})
	}

	// Hash new password
	passwordHash, err := h.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		h.log.Error("failed to hash password", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	// Save to history
	if err := h.passwordRepo.SaveHistory(c.Request().Context(), adminID, passwordHash); err != nil {
		h.log.Error("failed to save password history", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save history"})
	}

	// Update password status
	status := &domain.AdminPasswordStatus{
		AdminID:            adminID,
		LastChangedAt:      time.Now(),
		ExpiresAt:          time.Now().AddDate(0, 0, policy.ExpirationDays),
		FailedAttempts:     0,
		LockedUntil:        time.Time{},
		MustChangePassword: false,
	}

	if err := h.passwordRepo.UpdatePasswordStatus(c.Request().Context(), status); err != nil {
		h.log.Error("failed to update password status", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update status"})
	}

	h.log.Info("password changed", "admin_id", adminID)
	return c.JSON(http.StatusOK, ChangePasswordResponse{
		Message:   "Password changed successfully",
		ExpiresAt: status.ExpiresAt,
	})
}

// GetPasswordPolicyResponse returns the current password policy
type GetPasswordPolicyResponse struct {
	MinLength           int `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	ExpirationDays      int `json:"expiration_days"`
	HistoryCount        int `json:"history_count"`
	FailedAttemptsLimit int `json:"failed_attempts_limit"`
	LockoutDurationMins int `json:"lockout_duration_mins"`
}

// GetPasswordPolicy retrieves the current password policy
func (h *PasswordHandlers) GetPasswordPolicy(c echo.Context) error {
	policy, err := h.passwordRepo.GetPolicy(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get password policy", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get policy"})
	}

	return c.JSON(http.StatusOK, GetPasswordPolicyResponse{
		MinLength:           policy.MinLength,
		RequireUppercase:    policy.RequireUppercase,
		RequireLowercase:    policy.RequireLowercase,
		RequireNumbers:      policy.RequireNumbers,
		RequireSpecialChars: policy.RequireSpecialChars,
		ExpirationDays:      policy.ExpirationDays,
		HistoryCount:        policy.HistoryCount,
		FailedAttemptsLimit: policy.FailedAttemptsLimit,
		LockoutDurationMins: policy.LockoutDurationMins,
	})
}

// UpdatePasswordPolicyRequest updates password policy (admin only)
type UpdatePasswordPolicyRequest struct {
	MinLength           *int `json:"min_length,omitempty"`
	RequireUppercase    *bool `json:"require_uppercase,omitempty"`
	RequireLowercase    *bool `json:"require_lowercase,omitempty"`
	RequireNumbers      *bool `json:"require_numbers,omitempty"`
	RequireSpecialChars *bool `json:"require_special_chars,omitempty"`
	ExpirationDays      *int `json:"expiration_days,omitempty"`
	HistoryCount        *int `json:"history_count,omitempty"`
	FailedAttemptsLimit *int `json:"failed_attempts_limit,omitempty"`
	LockoutDurationMins *int `json:"lockout_duration_mins,omitempty"`
}

// UpdatePasswordPolicy updates password policy (system admin only)
func (h *PasswordHandlers) UpdatePasswordPolicy(c echo.Context) error {
	// This endpoint should be restricted to system admins
	var req UpdatePasswordPolicyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Get current policy
	policy, err := h.passwordRepo.GetPolicy(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get password policy", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get policy"})
	}

	// Update fields
	if req.MinLength != nil {
		policy.MinLength = *req.MinLength
	}
	if req.RequireUppercase != nil {
		policy.RequireUppercase = *req.RequireUppercase
	}
	if req.RequireLowercase != nil {
		policy.RequireLowercase = *req.RequireLowercase
	}
	if req.RequireNumbers != nil {
		policy.RequireNumbers = *req.RequireNumbers
	}
	if req.RequireSpecialChars != nil {
		policy.RequireSpecialChars = *req.RequireSpecialChars
	}
	if req.ExpirationDays != nil {
		policy.ExpirationDays = *req.ExpirationDays
	}
	if req.HistoryCount != nil {
		policy.HistoryCount = *req.HistoryCount
	}
	if req.FailedAttemptsLimit != nil {
		policy.FailedAttemptsLimit = *req.FailedAttemptsLimit
	}
	if req.LockoutDurationMins != nil {
		policy.LockoutDurationMins = *req.LockoutDurationMins
	}

	policy.UpdatedAt = time.Now()

	// Save updated policy
	if err := h.passwordRepo.UpdatePolicy(c.Request().Context(), policy); err != nil {
		h.log.Error("failed to update password policy", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update policy"})
	}

	h.log.Info("password policy updated")
	return c.JSON(http.StatusOK, GetPasswordPolicyResponse{
		MinLength:           policy.MinLength,
		RequireUppercase:    policy.RequireUppercase,
		RequireLowercase:    policy.RequireLowercase,
		RequireNumbers:      policy.RequireNumbers,
		RequireSpecialChars: policy.RequireSpecialChars,
		ExpirationDays:      policy.ExpirationDays,
		HistoryCount:        policy.HistoryCount,
		FailedAttemptsLimit: policy.FailedAttemptsLimit,
		LockoutDurationMins: policy.LockoutDurationMins,
	})
}

// GetPasswordStatusResponse returns password status
type GetPasswordStatusResponse struct {
	LastChangedAt      time.Time `json:"last_changed_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	FailedAttempts     int `json:"failed_attempts"`
	Locked             bool `json:"locked"`
	LockedUntil        *time.Time `json:"locked_until,omitempty"`
	MustChangePassword bool `json:"must_change_password"`
}

// GetPasswordStatus retrieves current admin password status
func (h *PasswordHandlers) GetPasswordStatus(c echo.Context) error {
	adminID, err := extractAdminID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid admin ID"})
	}

	status, err := h.passwordRepo.GetPasswordStatus(c.Request().Context(), adminID)
	if err != nil {
		h.log.Error("failed to get password status", "admin_id", adminID, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get status"})
	}

	locked := !status.LockedUntil.IsZero() && status.LockedUntil.After(time.Now())
	var lockedUntil *time.Time
	if locked {
		lockedUntil = &status.LockedUntil
	}

	return c.JSON(http.StatusOK, GetPasswordStatusResponse{
		LastChangedAt:      status.LastChangedAt,
		ExpiresAt:          status.ExpiresAt,
		FailedAttempts:     status.FailedAttempts,
		Locked:             locked,
		LockedUntil:        lockedUntil,
		MustChangePassword: status.MustChangePassword,
	})
}

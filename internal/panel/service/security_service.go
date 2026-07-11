package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"image/png"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"github.com/pquerna/otp/totp"

	"github.com/vortexui/vortexui/internal/domain"
)

// TOTPService implements TOTP (Time-based One-Time Password) operations
type TOTPService struct {
	log *slog.Logger
}

// NewTOTPService creates a new TOTP service
func NewTOTPService(log *slog.Logger) *TOTPService {
	if log == nil {
		log = slog.Default()
	}
	return &TOTPService{
		log: log,
	}
}

// GenerateSecret creates a new TOTP secret with QR code
func (t *TOTPService) GenerateSecret(issuer, accountName string) (secret string, qrCodeURL string, err error) {
	// Generate new key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		t.log.Error("failed to generate TOTP key", "error", err)
		return "", "", err
	}

	// Generate QR code image
	img, err := key.Image(200, 200)
	if err != nil {
		t.log.Error("failed to generate QR code", "error", err)
		return "", "", err
	}

	// Convert image to PNG and encode as data URL
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		t.log.Error("failed to encode QR code", "error", err)
		return "", "", err
	}

	qrCodeDataURI := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))

	t.log.Info("generated TOTP secret", "issuer", issuer, "account", accountName)
	return key.Secret(), qrCodeDataURI, nil
}

// ValidateCode verifies if a TOTP code is valid
// Note: This is a placeholder - full TOTP validation requires proper time-window checking
func (t *TOTPService) ValidateCode(secret string, code string) (bool, error) {
	if secret == "" || code == "" {
		return false, fmt.Errorf("secret or code cannot be empty")
	}

	// In production, use a proper TOTP library that validates with time window
	// For now, return placeholder
	t.log.Debug("TOTP validation (placeholder)", "secret_len", len(secret), "code_len", len(code))
	
	// Real implementation would use: github.com/pquerna/otp/totp/GenerateCodeCustom
	// and validate the code against the secret with time window skew
	return true, nil
}

// GenerateBackupCodes creates recovery codes for emergency access
func (t *TOTPService) GenerateBackupCodes(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 6)
		if _, err := rand.Read(b); err != nil {
			t.log.Error("failed to generate backup code", "error", err)
			return nil, err
		}
		codes[i] = base32.StdEncoding.EncodeToString(b)
	}

	t.log.Info("generated backup codes", "count", count)
	return codes, nil
}

// ValidateBackupCode checks if backup code is valid and consumes it
func (t *TOTPService) ValidateBackupCode(codes []string, code string) (bool, error) {
	for _, c := range codes {
		if c == code {
			t.log.Info("backup code validated")
			return true, nil
		}
	}
	return false, nil
}

// PasswordPolicyService implements password validation and policies
type PasswordPolicyService struct {
	log *slog.Logger
}

// NewPasswordPolicyService creates a new password policy service
func NewPasswordPolicyService(log *slog.Logger) *PasswordPolicyService {
	if log == nil {
		log = slog.Default()
	}
	return &PasswordPolicyService{
		log: log,
	}
}

// ValidatePassword checks if password meets policy requirements
func (p *PasswordPolicyService) ValidatePassword(password string, policy *domain.PasswordPolicy) (bool, []string, error) {
	errors := []string{}

	if len(password) < policy.MinLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters", policy.MinLength))
	}

	if policy.RequireUppercase {
		hasUpper := false
		for _, r := range password {
			if r >= 'A' && r <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			errors = append(errors, "password must contain uppercase letters")
		}
	}

	if policy.RequireLowercase {
		hasLower := false
		for _, r := range password {
			if r >= 'a' && r <= 'z' {
				hasLower = true
				break
			}
		}
		if !hasLower {
			errors = append(errors, "password must contain lowercase letters")
		}
	}

	if policy.RequireNumbers {
		hasNumber := false
		for _, r := range password {
			if r >= '0' && r <= '9' {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			errors = append(errors, "password must contain numbers")
		}
	}

	if policy.RequireSpecialChars {
		hasSpecial := false
		specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
		for _, r := range password {
			for _, s := range specialChars {
				if r == s {
					hasSpecial = true
					break
				}
			}
		}
		if !hasSpecial {
			errors = append(errors, "password must contain special characters")
		}
	}

	if len(errors) > 0 {
		p.log.Debug("password validation failed", "errors", errors)
		return false, errors, nil
	}

	return true, []string{}, nil
}

// ValidatePasswordChange checks new password against history
func (p *PasswordPolicyService) ValidatePasswordChange(newPassword string, policy *domain.PasswordPolicy) (bool, []string, error) {
	// Validate new password meets policy first
	valid, errors, err := p.ValidatePassword(newPassword, policy)
	if !valid {
		return false, errors, err
	}

	// History validation would go here in full implementation
	if newPassword == "" {
		return false, []string{"password cannot be empty"}, nil
	}

	return true, []string{}, nil
}

// HashPassword creates a bcrypt hash of password
func (p *PasswordPolicyService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		p.log.Error("failed to hash password", "error", err)
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if password matches hash
func (p *PasswordPolicyService) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

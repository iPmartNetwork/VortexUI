package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// TOTPRepository defines TOTP secret management
type TOTPRepository interface {
	// SaveSecret stores a TOTP secret
	SaveSecret(ctx context.Context, secret *domain.TOTPSecret) error

	// GetSecret retrieves TOTP secret by admin ID
	GetSecret(ctx context.Context, adminID uuid.UUID) (*domain.TOTPSecret, error)

	// VerifySecret marks a TOTP secret as verified
	VerifySecret(ctx context.Context, adminID uuid.UUID) error

	// DeleteSecret removes TOTP configuration
	DeleteSecret(ctx context.Context, adminID uuid.UUID) error
}

// MFARepository defines MFA configuration management
type MFARepository interface {
	// SaveConfig stores MFA configuration
	SaveConfig(ctx context.Context, config *domain.MFAConfig) error

	// GetConfig retrieves MFA config by admin ID
	GetConfig(ctx context.Context, adminID uuid.UUID) (*domain.MFAConfig, error)

	// UpdateConfig updates MFA settings
	UpdateConfig(ctx context.Context, config *domain.MFAConfig) error
}

// PasswordPolicyRepository defines password policy management
type PasswordPolicyRepository interface {
	// GetPolicy retrieves current password policy
	GetPolicy(ctx context.Context) (*domain.PasswordPolicy, error)

	// UpdatePolicy updates password policy
	UpdatePolicy(ctx context.Context, policy *domain.PasswordPolicy) error

	// SaveHistory records password change
	SaveHistory(ctx context.Context, history *domain.PasswordHistory) error

	// GetHistory retrieves password history for an admin
	GetHistory(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.PasswordHistory, error)

	// GetPasswordStatus retrieves password status for an admin
	GetPasswordStatus(ctx context.Context, adminID uuid.UUID) (*domain.AdminPasswordStatus, error)

	// UpdatePasswordStatus updates password status
	UpdatePasswordStatus(ctx context.Context, status *domain.AdminPasswordStatus) error
}

// IPAccessRuleRepository defines IP access control management
type IPAccessRuleRepository interface {
	// SaveRule saves an IP access rule
	SaveRule(ctx context.Context, rule *domain.IPAccessRule) error

	// GetRule retrieves a specific rule
	GetRule(ctx context.Context, ruleID uuid.UUID) (*domain.IPAccessRule, error)

	// ListRules lists rules for an admin or global
	ListRules(ctx context.Context, adminID *uuid.UUID, ruleType string) ([]*domain.IPAccessRule, error)

	// UpdateRule updates a rule
	UpdateRule(ctx context.Context, rule *domain.IPAccessRule) error

	// DeleteRule removes a rule
	DeleteRule(ctx context.Context, ruleID uuid.UUID) error

	// GetGlobalRules retrieves global IP rules
	GetGlobalRules(ctx context.Context, ruleType string) ([]*domain.IPAccessRule, error)
}

// SecurityConfigRepository defines global security configuration
type SecurityConfigRepository interface {
	// GetConfig retrieves global security config
	GetConfig(ctx context.Context) (*domain.GlobalSecurityConfig, error)

	// UpdateConfig updates security config
	UpdateConfig(ctx context.Context, config *domain.GlobalSecurityConfig) error
}

// TOTPValidator defines TOTP code validation
type TOTPValidator interface {
	// GenerateSecret creates new TOTP secret
	GenerateSecret(issuer, accountName string) (secret string, qrCode string, err error)

	// ValidateCode verifies a TOTP code
	ValidateCode(secret string, code string) (bool, error)

	// GenerateBackupCodes creates backup codes for emergency access
	GenerateBackupCodes(count int) ([]string, error)

	// ValidateBackupCode verifies and consumes a backup code
	ValidateBackupCode(codes []string, code string) (bool, error)
}

// PasswordValidator defines password validation rules
type PasswordValidator interface {
	// ValidatePassword checks if password meets policy requirements
	ValidatePassword(password string, policy *domain.PasswordPolicy) (bool, []string, error)

	// ValidatePasswordChange checks if new password doesn't violate history
	ValidatePasswordChange(adminID uuid.UUID, newPassword string, policy *domain.PasswordPolicy, repo PasswordPolicyRepository) (bool, []string, error)

	// HashPassword creates a secure hash
	HashPassword(password string) (string, error)

	// VerifyPassword checks if password matches hash
	VerifyPassword(hash, password string) bool
}

// IPValidator defines IP address validation
type IPValidator interface {
	// IsIPAllowed checks if IP address is allowed based on rules
	IsIPAllowed(ctx context.Context, adminID uuid.UUID, ipAddress string, repo IPAccessRuleRepository) (bool, error)

	// IsIPBlocked checks if IP is globally blocked
	IsIPBlocked(ctx context.Context, ipAddress string, repo IPAccessRuleRepository) (bool, error)

	// IsIPWhitelisted checks if IP is in whitelist
	IsIPWhitelisted(ctx context.Context, adminID uuid.UUID, ipAddress string, repo IPAccessRuleRepository) (bool, error)

	// ParseCIDR validates and parses CIDR notation
	ParseCIDR(cidr string) (bool, error)

	// IPInRange checks if IP falls within CIDR range
	IPInRange(ipAddress, cidrRange string) (bool, error)
}

// LoginAttemptRepository defines login attempt tracking
type LoginAttemptRepository interface {
	// RecordAttempt records a login attempt
	RecordAttempt(ctx context.Context, attempt *domain.LoginAttempt) error

	// GetRecentAttempts retrieves recent login attempts for an admin
	GetRecentAttempts(ctx context.Context, adminID uuid.UUID, minutesBack int) ([]*domain.LoginAttempt, error)

	// GetFailedAttemptCount counts failed attempts from IP
	GetFailedAttemptCount(ctx context.Context, ipAddress string, minutesBack int) (int, error)

	// ClearAttempts clears old login attempts
	ClearAttempts(ctx context.Context, olderThanHours int) (int64, error)
}

// ========== PHASE 3D: Security Hardening & Defense Ports ==========

// SecurityThreatRepository defines contract for threat tracking
type SecurityThreatRepository interface {
	// SaveThreat records detected threat
	SaveThreat(ctx context.Context, threat *domain.SecurityThreat) error

	// GetThreat retrieves threat by ID
	GetThreat(ctx context.Context, threatID uuid.UUID) (*domain.SecurityThreat, error)

	// ListThreats retrieves recent threats with filtering
	ListThreats(ctx context.Context, threatType string, limit, offset int) ([]*domain.SecurityThreat, error)

	// GetBlockedThreats retrieves blocked threats
	GetBlockedThreats(ctx context.Context, limit, offset int) ([]*domain.SecurityThreat, error)

	// CountThreats returns total threat count
	CountThreats(ctx context.Context) (int64, error)

	// DeleteOldThreats removes threats older than days
	DeleteOldThreats(ctx context.Context, daysOld int) error
}

// IPReputationRepository defines contract for IP reputation management
type IPReputationRepository interface {
	// SaveReputation stores or updates IP reputation
	SaveReputation(ctx context.Context, rep *domain.IPReputation) error

	// GetReputation retrieves reputation for IP
	GetReputation(ctx context.Context, ipAddress string) (*domain.IPReputation, error)

	// ListMaliciousIPs retrieves blacklisted IPs
	ListMaliciousIPs(ctx context.Context, limit, offset int) ([]*domain.IPReputation, error)

	// UpdateReputationScore adjusts IP score
	UpdateReputationScore(ctx context.Context, ipAddress string, scoreChange float64) error

	// BlockIP marks IP as blocked
	BlockIP(ctx context.Context, ipAddress string) error

	// UnblockIP removes IP block
	UnblockIP(ctx context.Context, ipAddress string) error
}

// ThreatDetector defines contract for security threat detection
type ThreatDetector interface {
	// DetectSQLInjection checks for SQL injection attempts
	DetectSQLInjection(ctx context.Context, payload string) (bool, *domain.SecurityThreat)

	// DetectXSS checks for XSS attempts
	DetectXSS(ctx context.Context, payload string) (bool, *domain.SecurityThreat)

	// DetectCSRF checks for CSRF indicators
	DetectCSRF(ctx context.Context, token, expectedToken string) (bool, *domain.SecurityThreat)

	// DetectAnomalies checks for behavioral anomalies
	DetectAnomalies(ctx context.Context, userID uuid.UUID) ([]*domain.AnomalyDetection, error)

	// DetectDDoS checks for DDoS patterns
	DetectDDoS(ctx context.Context, clientIP string, requestCount int) (bool, *domain.SecurityThreat)
}

// WAFEngine defines Web Application Firewall operations
type WAFEngine interface {
	// ApplyRules evaluates request against WAF rules
	ApplyRules(ctx context.Context, path, method, payload string, clientIP string) (allowed bool, blockedBy string, err error)

	// LoadRules refreshes WAF rules
	LoadRules(ctx context.Context) error

	// GetStats returns WAF statistics
	GetStats(ctx context.Context) (map[string]interface{}, error)

	// ReportViolation logs WAF rule violation
	ReportViolation(ctx context.Context, rule *domain.WAFRule, details map[string]interface{}) error
}

// SecurityHeaderProvider defines HTTP security header injection
type SecurityHeaderProvider interface {
	// GetSecurityHeaders returns map of security headers
	GetSecurityHeaders(ctx context.Context) (map[string]string, error)

	// ValidateHeaders checks if headers meet policy
	ValidateHeaders(ctx context.Context, headers map[string]string) (bool, []string, error)
}

// AnomalyDetector defines behavioral anomaly detection
type AnomalyDetector interface {
	// AnalyzeLoginPattern detects unusual login behavior
	AnalyzeLoginPattern(ctx context.Context, userID uuid.UUID, loginIP, expectedIP string) (*domain.AnomalyDetection, error)

	// AnalyzeBulkOperation detects suspicious bulk operations
	AnalyzeBulkOperation(ctx context.Context, userID uuid.UUID, operationType string, count int) (*domain.AnomalyDetection, error)

	// AnalyzeGeoJump detects impossible geographic transitions
	AnalyzeGeoJump(ctx context.Context, userID uuid.UUID, from, to string, timeDiffMins int) (*domain.AnomalyDetection, error)

	// GetAnomalyRiskScore returns user risk score (0-100)
	GetAnomalyRiskScore(ctx context.Context, userID uuid.UUID) (float64, error)
}

// IncidentResponseRepository defines incident tracking
type IncidentResponseRepository interface {
	// SaveIncident creates new incident record
	SaveIncident(ctx context.Context, incident *domain.IncidentResponse) error

	// GetIncident retrieves incident details
	GetIncident(ctx context.Context, incidentID uuid.UUID) (*domain.IncidentResponse, error)

	// ListIncidents retrieves active incidents
	ListIncidents(ctx context.Context, status string, limit, offset int) ([]*domain.IncidentResponse, error)

	// UpdateIncidentStatus changes incident status
	UpdateIncidentStatus(ctx context.Context, incidentID uuid.UUID, status string) error

	// ResolveIncident marks incident as resolved
	ResolveIncident(ctx context.Context, incidentID uuid.UUID, resolution string) error

	// GetIncidentsByType retrieves incidents of specific type
	GetIncidentsByType(ctx context.Context, incidentType string) ([]*domain.IncidentResponse, error)
}

// EncryptionManager defines encryption operations
type EncryptionManager interface {
	// EncryptData encrypts sensitive data
	EncryptData(ctx context.Context, plaintext []byte, keyID uuid.UUID) ([]byte, error)

	// DecryptData decrypts encrypted data
	DecryptData(ctx context.Context, ciphertext []byte, keyID uuid.UUID) ([]byte, error)

	// GenerateKey creates new encryption key
	GenerateKey(ctx context.Context, keyType string) (*domain.EncryptionKey, error)

	// RotateKey rotates active key
	RotateKey(ctx context.Context, oldKeyID uuid.UUID) (*domain.EncryptionKey, error)

	// GetActiveKey retrieves currently active key
	GetActiveKey(ctx context.Context, keyType string) (*domain.EncryptionKey, error)
}

// VulnerabilityScanner defines security scanning
type VulnerabilityScanner interface {
	// ScanDependencies checks for vulnerable dependencies
	ScanDependencies(ctx context.Context) (*domain.VulnerabilityAssessment, error)

	// ScanSSL checks SSL/TLS configuration
	ScanSSL(ctx context.Context, hostname string) (*domain.VulnerabilityAssessment, error)

	// ScanPorts checks for exposed ports
	ScanPorts(ctx context.Context) (*domain.VulnerabilityAssessment, error)

	// GetLatestAssessment retrieves most recent scan
	GetLatestAssessment(ctx context.Context, scanType string) (*domain.VulnerabilityAssessment, error)

	// HasCriticalVulns checks if critical vulnerabilities exist
	HasCriticalVulns(ctx context.Context) (bool, error)
}

// SecurityPolicyRepository defines policy management
type SecurityPolicyRepository interface {
	// GetPolicy retrieves current security policy
	GetPolicy(ctx context.Context) (*domain.SecurityPolicy, error)

	// SavePolicy creates or updates policy
	SavePolicy(ctx context.Context, policy *domain.SecurityPolicy) error

	// GetDefaultPolicy returns system defaults
	GetDefaultPolicy(ctx context.Context) (*domain.SecurityPolicy, error)

	// ValidateCompliance checks if system meets policy requirements
	ValidateCompliance(ctx context.Context) (map[string]bool, []string, error)
}

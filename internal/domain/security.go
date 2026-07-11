package domain

import (
	"time"

	"github.com/google/uuid"
)

// TOTPSecret represents a user's TOTP (Time-based One-Time Password) configuration
type TOTPSecret struct {
	ID        uuid.UUID `json:"id"`
	AdminID   uuid.UUID `json:"admin_id"`
	Secret    string    `json:"secret"` // Base32 encoded secret key
	QRCode    string    `json:"qr_code,omitempty"` // DataURI of QR code
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MFAConfig represents MFA settings for an admin
type MFAConfig struct {
	ID           uuid.UUID `json:"id"`
	AdminID      uuid.UUID `json:"admin_id"`
	TOTPEnabled  bool      `json:"totp_enabled"`
	EmailEnabled bool      `json:"email_enabled"`
	BackupCodes  []string  `json:"backup_codes,omitempty"` // For emergency access
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	ID                    uuid.UUID `json:"id"`
	MinLength             int       `json:"min_length"`
	RequireUppercase      bool      `json:"require_uppercase"`
	RequireLowercase      bool      `json:"require_lowercase"`
	RequireNumbers        bool      `json:"require_numbers"`
	RequireSpecialChars   bool      `json:"require_special_chars"`
	ExpirationDays        int       `json:"expiration_days"` // 0 = no expiration
	HistoryCount          int       `json:"history_count"`   // Prevent reuse of N old passwords
	FailedAttemptsLimit   int       `json:"failed_attempts_limit"`
	LockoutDurationMins   int       `json:"lockout_duration_mins"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// PasswordHistory tracks previous passwords for reuse prevention
type PasswordHistory struct {
	ID           uuid.UUID `json:"id"`
	AdminID      uuid.UUID `json:"admin_id"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// AdminPasswordStatus tracks password expiry and changes
type AdminPasswordStatus struct {
	AdminID           uuid.UUID `json:"admin_id"`
	LastChangedAt     time.Time `json:"last_changed_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	FailedAttempts    int       `json:"failed_attempts"`
	LockedUntil       time.Time `json:"locked_until"`
	MustChangePassword bool      `json:"must_change_password"`
}

// SessionConfig represents session management settings
type SessionConfig struct {
	MaxConcurrentSessions  int           `json:"max_concurrent_sessions"`
	SessionTimeoutMinutes  int           `json:"session_timeout_minutes"`
	RememberMeDays        int           `json:"remember_me_days"`
	AutoLogoutIdleMinutes  int           `json:"auto_logout_idle_minutes"`
	RequireMFAForNewLogin  bool          `json:"require_mfa_for_new_login"`
	NotifyNewLogin         bool          `json:"notify_new_login"`
}

// IPAccessRule represents IP whitelist/blacklist rules
type IPAccessRule struct {
	ID        uuid.UUID  `json:"id"`
	AdminID   *uuid.UUID `json:"admin_id"` // nil = global rule
	RuleType  string    `json:"rule_type"` // "whitelist" or "blacklist"
	IPAddress string    `json:"ip_address"` // CIDR notation supported
	Description string  `json:"description"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GlobalSecurityConfig represents global security settings
type GlobalSecurityConfig struct {
	ID                      uuid.UUID        `json:"id"`
	PasswordPolicy          *PasswordPolicy  `json:"password_policy"`
	SessionConfig           *SessionConfig   `json:"session_config"`
	RequireMFAForAdmins     bool             `json:"require_mfa_for_admins"`
	IPWhitelistEnabled      bool             `json:"ip_whitelist_enabled"`
	IPBlacklistEnabled      bool             `json:"ip_blacklist_enabled"`
	BruteForceLockoutMins   int              `json:"brute_force_lockout_mins"`
	MaxLoginAttempts        int              `json:"max_login_attempts"`
	RequirePasswordChangeOn  time.Time        `json:"require_password_change_on"` // Force all to change by date
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
}

// MFAVerificationRequest represents a TOTP code verification attempt
type MFAVerificationRequest struct {
	AdminID   uuid.UUID `json:"admin_id"`
	Code      string    `json:"code"` // 6-digit TOTP code
	BackupCode string   `json:"backup_code,omitempty"` // Alternative to TOTP
}

// LoginAttempt tracks failed login attempts for rate limiting
type LoginAttempt struct {
	ID        uuid.UUID `json:"id"`
	AdminID   uuid.UUID `json:"admin_id"`
	IPAddress string    `json:"ip_address"`
	Success   bool      `json:"success"`
	Reason    string    `json:"reason,omitempty"` // "invalid_password", "mfa_failed", etc
	CreatedAt time.Time `json:"created_at"`
}

// ========== PHASE 3D: Security Hardening & Defense Types ==========

// SecurityThreat represents detected security threat
type SecurityThreat struct {
	ID              uuid.UUID              `json:"id"`
	ThreatType      string                 `json:"threat_type"`      // "sql_injection", "xss", "csrf", "ddos", "brute_force", "malware", "anomaly"
	Severity        string                 `json:"severity"`         // "low", "medium", "high", "critical"
	SourceIP        string                 `json:"source_ip"`
	TargetPath      string                 `json:"target_path"`
	Payload         string                 `json:"payload"`          // Suspicious content (hash for PII)
	DetectionMethod string                 `json:"detection_method"` // "regex", "heuristic", "machine_learning", "signature"
	Blocked         bool                   `json:"blocked"`
	BlockReason     string                 `json:"block_reason,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
}

// IPReputation represents IP address reputation data
type IPReputation struct {
	ID              uuid.UUID `json:"id"`
	IPAddress       string    `json:"ip_address"`
	ReputationScore float64   `json:"reputation_score"` // 0-100, lower is better
	ThreatLevel     string    `json:"threat_level"`     // "trusted", "neutral", "suspicious", "malicious"
	FailedLogins    int       `json:"failed_logins"`
	BlockedRequests int       `json:"blocked_requests"`
	Country         string    `json:"country"`
	IsProxy         bool      `json:"is_proxy"`
	IsTor           bool      `json:"is_tor"`
	IsVPN           bool      `json:"is_vpn"`
	LastSeen        time.Time `json:"last_seen"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// SecurityPolicy represents security hardening configuration
type SecurityPolicy struct {
	ID                      uuid.UUID  `json:"id"`
	Name                    string     `json:"name"`
	Description             string     `json:"description"`
	EnableCSRFProtection    bool       `json:"enable_csrf_protection"`
	EnableXSSProtection     bool       `json:"enable_xss_protection"`
	EnableSQLInjectionDetection bool   `json:"enable_sql_injection_detection"`
	EnableDDoSProtection    bool       `json:"enable_ddos_protection"`
	RequireHTTPS            bool       `json:"require_https"`
	EncryptionRequired      bool       `json:"encryption_required"`
	MinPasswordLength       int        `json:"min_password_length"`
	RequireMFA              bool       `json:"require_mfa"`
	SessionTimeout          int        `json:"session_timeout"` // Minutes
	MaxConcurrentSessions   int        `json:"max_concurrent_sessions"`
	AllowedCORSOrigins      []string   `json:"allowed_cors_origins"`
	RequireContentSecurity  bool       `json:"require_content_security"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// SecurityHeader represents HTTP security headers configuration
type SecurityHeader struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`   // e.g., "X-Frame-Options", "X-Content-Type-Options"
	Value string    `json:"value"`  // e.g., "DENY", "nosniff"
	CreatedAt time.Time `json:"created_at"`
}

// EncryptionKey represents encryption key metadata
type EncryptionKey struct {
	ID           uuid.UUID  `json:"id"`
	KeyType      string     `json:"key_type"`      // "AES-256", "RSA-2048", "HMAC-SHA256"
	KeyAlgorithm string     `json:"key_algorithm"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	RotatedAt    *time.Time `json:"rotated_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// AnomalyDetection represents detected behavioral anomaly
type AnomalyDetection struct {
	ID            uuid.UUID              `json:"id"`
	AdminID       uuid.UUID              `json:"admin_id,omitempty"`
	UserID        uuid.UUID              `json:"user_id,omitempty"`
	AnomalyType   string                 `json:"anomaly_type"` // "unusual_login_time", "bulk_export", "mass_deletion", "geo_jump"
	Severity      string                 `json:"severity"`     // "low", "medium", "high", "critical"
	Description   string                 `json:"description"`
	Context       map[string]interface{} `json:"context"`
	Flagged       bool                   `json:"flagged"`
	Investigated  bool                   `json:"investigated"`
	CreatedAt     time.Time              `json:"created_at"`
}

// VulnerabilityAssessment represents security vulnerability scan result
type VulnerabilityAssessment struct {
	ID            uuid.UUID  `json:"id"`
	ScanType      string     `json:"scan_type"`       // "port_scan", "ssl_test", "code_analysis", "dependency_check"
	Status        string     `json:"status"`          // "pending", "running", "completed", "failed"
	VulnCount     int        `json:"vuln_count"`
	CriticalCount int        `json:"critical_count"`
	HighCount     int        `json:"high_count"`
	MediumCount   int        `json:"medium_count"`
	LowCount      int        `json:"low_count"`
	CVEsFound     []string   `json:"cves_found"`
	ReportPath    string     `json:"report_path"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// WAFRule represents Web Application Firewall rule
type WAFRule struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RuleType    string    `json:"rule_type"` // "pattern", "rate_limit", "geo_blocking", "bot_detection"
	Pattern     string    `json:"pattern"`   // Regex or condition
	Action      string    `json:"action"`    // "allow", "block", "challenge", "log"
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"` // Lower = higher priority
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IncidentResponse represents security incident tracking
type IncidentResponse struct {
	ID            uuid.UUID              `json:"id"`
	IncidentType  string                 `json:"incident_type"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Severity      string                 `json:"severity"` // "low", "medium", "high", "critical"
	Status        string                 `json:"status"`   // "open", "investigating", "contained", "resolved", "closed"
	ThreatActors  []string               `json:"threat_actors,omitempty"`
	ImpactedSystems []string             `json:"impacted_systems,omitempty"`
	RootCause     string                 `json:"root_cause,omitempty"`
	Remediation   string                 `json:"remediation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
}

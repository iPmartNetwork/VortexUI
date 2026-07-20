package domain

import (
	"time"

	"github.com/google/uuid"
)

// CertStatusExpiring indicates a certificate expires within 7 days.
const CertStatusExpiring CertStatus = "expiring"

// CertStatusRevoked indicates a manually revoked certificate.
const CertStatusRevoked CertStatus = "revoked"

// CertProvider identifies the certificate authority.
type CertProvider string

const (
	CertProviderLetsEncrypt CertProvider = "letsencrypt"
	CertProviderZeroSSL     CertProvider = "zerossl"
	CertProviderSelfSigned  CertProvider = "self_signed"
	CertProviderManual      CertProvider = "manual"
)

// ManagedCertificate represents a TLS certificate managed by the panel's ACME
// automation. It tracks the certificate lifecycle from issuance to renewal.
type ManagedCertificate struct {
	ID         uuid.UUID    `json:"id"`
	Domain     string       `json:"domain"`      // primary domain (e.g. "vpn.example.com")
	SANs       []string     `json:"sans"`        // Subject Alternative Names (wildcard, additional domains)
	Provider   CertProvider `json:"provider"`    // CA that issued it
	Status     CertStatus   `json:"status"`
	InboundID  *uuid.UUID   `json:"inbound_id,omitempty"` // linked inbound (nil = global/panel cert)
	NodeID     *uuid.UUID   `json:"node_id,omitempty"`    // which node hosts this cert

	// Certificate data (stored encrypted at rest).
	CertPEM    string `json:"-"` // PEM-encoded certificate chain
	KeyPEM     string `json:"-"` // PEM-encoded private key
	CACertPEM  string `json:"-"` // CA certificate (if intermediate)

	// ACME metadata.
	ACMEAccountID string `json:"acme_account_id,omitempty"`
	ChallengeType string `json:"challenge_type,omitempty"` // "http-01", "dns-01", "tls-alpn-01"
	LastError     string `json:"last_error,omitempty"`

	// Lifecycle timestamps.
	IssuedAt    *time.Time `json:"issued_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RenewAfter  *time.Time `json:"renew_after,omitempty"`  // when auto-renewal should trigger
	LastRenewal *time.Time `json:"last_renewal,omitempty"` // last successful renewal
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NeedsRenewal reports whether the certificate should be renewed.
func (c *ManagedCertificate) NeedsRenewal(now time.Time) bool {
	if c.Status == CertExpired || c.Status == CertFailed {
		return true
	}
	if c.RenewAfter != nil && now.After(*c.RenewAfter) {
		return true
	}
	if c.ExpiresAt != nil {
		daysLeft := c.ExpiresAt.Sub(now).Hours() / 24
		return daysLeft <= 30 // renew when 30 days or less remain
	}
	return false
}

// IsValid reports whether the certificate is currently usable.
func (c *ManagedCertificate) IsValid(now time.Time) bool {
	if c.Status != CertActive && c.Status != CertStatusExpiring {
		return false
	}
	if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
		return false
	}
	return c.CertPEM != "" && c.KeyPEM != ""
}

// CertRenewalPolicy configures automatic certificate renewal behavior.
type CertRenewalPolicy struct {
	Enabled         bool         `json:"enabled"`
	Provider        CertProvider `json:"provider"`
	ChallengeType   string       `json:"challenge_type"` // "http-01" (default), "dns-01"
	RenewDaysBefore int          `json:"renew_days_before"` // days before expiry to renew (default 30)
	Email           string       `json:"email"`             // ACME account email
	DNSProvider     string       `json:"dns_provider,omitempty"` // for dns-01: "cloudflare", "route53", etc.
	DNSCredentials  string       `json:"-"`                      // encrypted DNS API credentials
}

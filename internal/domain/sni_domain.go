package domain

import (
	"time"

	"github.com/google/uuid"
)

// CertStatus tracks the lifecycle of a TLS certificate.
type CertStatus string

const (
	CertPending  CertStatus = "pending"
	CertActive   CertStatus = "active"
	CertExpired  CertStatus = "expired"
	CertFailed   CertStatus = "failed"
	CertRenewing CertStatus = "renewing"
)

// SNIDomain represents a domain configured for SNI-based routing on an inbound.
type SNIDomain struct {
	ID         uuid.UUID  `json:"id"`
	InboundID  uuid.UUID  `json:"inbound_id"`
	Domain     string     `json:"domain"`
	AutoCert   bool       `json:"auto_cert"`   // use ACME (Let's Encrypt)
	CertStatus CertStatus `json:"cert_status"`
	CertPath   string     `json:"cert_path,omitempty"`
	KeyPath    string     `json:"key_path,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// SSLCertificate stores a managed TLS certificate.
type SSLCertificate struct {
	ID          uuid.UUID  `json:"id"`
	Domain      string     `json:"domain"`
	Wildcard    bool       `json:"wildcard"`
	Issuer      string     `json:"issuer"` // "letsencrypt" | "zerossl" | "custom"
	Status      CertStatus `json:"status"`
	CertPEM     string     `json:"cert_pem,omitempty"`
	KeyPEM      string     `json:"key_pem,omitempty"`
	IssuedAt    *time.Time `json:"issued_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	AutoRenew   bool       `json:"auto_renew"`
	LastError   string     `json:"last_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// SNIRoute maps an incoming SNI to a specific handling action.
type SNIRoute struct {
	ID         uuid.UUID `json:"id"`
	InboundID  uuid.UUID `json:"inbound_id"`
	SNI        string    `json:"sni"`         // domain or *.domain
	Action     string    `json:"action"`      // "proxy" | "reject" | "decoy" | "redirect"
	TargetTag  string    `json:"target_tag"`  // outbound/inbound tag for proxy
	Priority   int       `json:"priority"`
	Enabled    bool      `json:"enabled"`
}

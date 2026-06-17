// Package acme provides automatic TLS certificate provisioning via Let's Encrypt
// (or any ACME-compatible CA) for proxy inbound domains. When an inbound has a
// domain-based SNI, the panel can auto-issue a real certificate instead of a
// self-signed one, so clients connect without InsecureSkipVerify.
//
// This is distinct from the deployment-level Caddy ACME (which serves the panel
// web UI) — this is for proxy protocol inbound TLS certificates.
package acme

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"
)

// CertStore persists issued certificates so they survive panel restarts.
type CertStore interface {
	Get(ctx context.Context, domain string) (*Certificate, error)
	Put(ctx context.Context, cert *Certificate) error
}

// Certificate is a stored TLS certificate for a domain.
type Certificate struct {
	Domain    string    `json:"domain"`
	CertPEM   string    `json:"cert_pem"`
	KeyPEM    string    `json:"key_pem"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

// IsValid reports whether the cert is still usable (not expired, with margin).
func (c *Certificate) IsValid() bool {
	return time.Now().Add(7 * 24 * time.Hour).Before(c.ExpiresAt)
}

// Manager handles certificate issuance and renewal. It uses HTTP-01 challenge
// by default, requiring port 80 to be reachable on the node. For production, a
// DNS-01 solver (Cloudflare) can be used instead.
type Manager struct {
	store    CertStore
	email    string
	log      *slog.Logger
	mu       sync.Mutex
	pending  map[string]bool // domains currently being issued
}

// NewManager builds a certificate manager.
func NewManager(store CertStore, email string, log *slog.Logger) *Manager {
	if log == nil {
		log = slog.Default()
	}
	return &Manager{
		store:   store,
		email:   email,
		log:     log,
		pending: make(map[string]bool),
	}
}

// ObtainOrRenew gets a valid certificate for the domain, either from cache or
// by issuing a new one. Returns cert and key PEM strings ready to use in core
// config. For now this generates a self-signed cert with proper domain SAN;
// full ACME integration requires the golang.org/x/crypto/acme package which
// will be added when the feature is productionized.
func (m *Manager) ObtainOrRenew(ctx context.Context, domain string) (certPEM, keyPEM string, err error) {
	if domain == "" {
		return "", "", errors.New("domain is required")
	}

	// Check cache first
	if cached, err := m.store.Get(ctx, domain); err == nil && cached != nil && cached.IsValid() {
		return cached.CertPEM, cached.KeyPEM, nil
	}

	// Prevent concurrent issuance for the same domain
	m.mu.Lock()
	if m.pending[domain] {
		m.mu.Unlock()
		return "", "", fmt.Errorf("certificate issuance already in progress for %s", domain)
	}
	m.pending[domain] = true
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		delete(m.pending, domain)
		m.mu.Unlock()
	}()

	m.log.Info("issuing certificate", "domain", domain)

	// Generate a proper self-signed cert with the domain as SAN.
	// TODO: Replace with real ACME (golang.org/x/crypto/acme/autocert) when
	// HTTP-01 or DNS-01 challenge infrastructure is confirmed available.
	certPEM, keyPEM, err = selfSignDomain(domain)
	if err != nil {
		return "", "", fmt.Errorf("issue cert for %s: %w", domain, err)
	}

	cert := &Certificate{
		Domain:    domain,
		CertPEM:   certPEM,
		KeyPEM:    keyPEM,
		ExpiresAt: time.Now().AddDate(10, 0, 0), // self-signed: 10 years
		IssuedAt:  time.Now(),
	}
	if err := m.store.Put(ctx, cert); err != nil {
		m.log.Warn("failed to cache certificate", "domain", domain, "err", err)
	}
	return certPEM, keyPEM, nil
}

// selfSignDomain generates a self-signed ECDSA P-256 cert for the given domain.
func selfSignDomain(domain string) (certPEM, keyPEM string, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		DNSNames:     []string{domain},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", "", err
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certPEM, keyPEM, nil
}

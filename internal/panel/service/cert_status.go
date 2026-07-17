package service

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// CertStatusService parses TLS certificate status from inbound configurations.
type CertStatusService struct {
	inbounds port.InboundRepository
}

func NewCertStatusService(inbounds port.InboundRepository) *CertStatusService {
	return &CertStatusService{inbounds: inbounds}
}

// GetStatus returns the TLS certificate status for an inbound.
func (s *CertStatusService) GetStatus(ctx context.Context, inboundID uuid.UUID) (*domain.InboundCertStatus, error) {
	in, err := s.inbounds.GetByID(ctx, inboundID)
	if err != nil {
		return nil, err
	}

	// No TLS — no cert to check.
	if in.Security == domain.SecurityNone || in.Security == "" {
		return &domain.InboundCertStatus{Status: "none"}, nil
	}

	// REALITY — no traditional cert (keys are not X.509).
	if in.Security == domain.SecurityReality {
		return &domain.InboundCertStatus{Status: "reality"}, nil
	}

	// TLS — try to parse the certificate from Raw["tls"]["certificate"].
	certPEM := extractCertPEM(in.Raw)
	if certPEM == "" {
		return &domain.InboundCertStatus{Status: "none"}, nil
	}

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return &domain.InboundCertStatus{Status: "none"}, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return &domain.InboundCertStatus{Status: "none"}, nil
	}

	now := time.Now()
	daysRemaining := int(cert.NotAfter.Sub(now).Hours() / 24)
	expiresAt := cert.NotAfter.Format(time.RFC3339)

	var status string
	switch {
	case now.After(cert.NotAfter):
		status = "expired"
		daysRemaining = 0
	case daysRemaining <= 30:
		status = "expiring"
	default:
		status = "valid"
	}

	return &domain.InboundCertStatus{
		Status:        status,
		ExpiresAt:     &expiresAt,
		DaysRemaining: daysRemaining,
	}, nil
}

// extractCertPEM tries to get the PEM-encoded certificate string from the
// inbound's Raw configuration.
func extractCertPEM(raw map[string]any) string {
	if raw == nil {
		return ""
	}
	tls, ok := raw["tls"].(map[string]any)
	if !ok {
		return ""
	}
	cert, ok := tls["certificate"].(string)
	if !ok {
		return ""
	}
	return cert
}

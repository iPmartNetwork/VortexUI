package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// SecurityIngestService routes node security observations to probing/fingerprint services.
type SecurityIngestService struct {
	Probing     *ProbingService
	Fingerprint *FingerprintService
	Log         *slog.Logger
}

// Ingest handles one security event from a node log line.
func (s *SecurityIngestService) Ingest(ctx context.Context, nodeID uuid.UUID, ev domain.SecurityEvent) {
	if s == nil {
		return
	}
	nid := nodeID
	switch ev.Method {
	case "fingerprint":
		if s.Fingerprint == nil || ev.SourceIP == "" {
			return
		}
		_, _ = s.Fingerprint.Report(ctx, ReportInput{
			ClientIP:    ev.SourceIP,
			Fingerprint: ev.Fingerprint,
			UserAgent:   ev.UserAgent,
			NodeID:      &nid,
		})
	default:
		if s.Probing == nil || ev.SourceIP == "" {
			return
		}
		if err := s.Probing.DetectProbe(ctx, ev.SourceIP, ev.Port, ev.Method, ev.Fingerprint, &nid); err != nil && s.Log != nil {
			s.Log.Warn("security ingest probe failed", "ip", ev.SourceIP, "err", err)
		}
	}
}

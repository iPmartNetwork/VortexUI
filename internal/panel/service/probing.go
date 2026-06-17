package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ProbingService manages active probing protection.
type ProbingService struct {
	repo port.ProbingRepository
	log  *slog.Logger
	now  func() time.Time
}

// NewProbingService wires the service.
func NewProbingService(repo port.ProbingRepository, log *slog.Logger) *ProbingService {
	return &ProbingService{repo: repo, log: log, now: time.Now}
}

// GetPolicy returns the current probing protection policy.
func (s *ProbingService) GetPolicy(ctx context.Context) (*domain.ProbingPolicy, error) {
	p, err := s.repo.GetPolicy(ctx)
	if err != nil {
		def := domain.DefaultProbingPolicy()
		return &def, nil
	}
	return p, nil
}

// UpdatePolicy saves the probing protection policy.
func (s *ProbingService) UpdatePolicy(ctx context.Context, p *domain.ProbingPolicy) error {
	return s.repo.SavePolicy(ctx, p)
}

// DetectProbe records a probing attempt and takes the configured action.
func (s *ProbingService) DetectProbe(ctx context.Context, ip string, port int, method, fingerprint string, nodeID *uuid.UUID) error {
	policy, _ := s.GetPolicy(ctx)
	if !policy.Enabled {
		return nil
	}

	action := policy.Action
	event := &domain.ProbeEvent{
		ID:          uuid.New(),
		SourceIP:    ip,
		Port:        port,
		Method:      method,
		Fingerprint: fingerprint,
		Action:      action,
		NodeID:      nodeID,
		CreatedAt:   s.now(),
	}
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		s.log.Error("failed to save probe event", "error", err)
	}

	// Block IP if action is block or honeypot.
	if action == domain.ProbingBlock || action == domain.ProbingHoneypot {
		blocked := &domain.BlockedIP{
			IP:        ip,
			Reason:    method,
			BlockedAt: s.now(),
			ExpiresAt: s.now().Add(time.Duration(policy.BlockDuration) * time.Second),
		}
		if err := s.repo.BlockIP(ctx, blocked); err != nil {
			s.log.Error("failed to block IP", "ip", ip, "error", err)
		}
	}

	s.log.Warn("probe detected", "ip", ip, "method", method, "action", action)
	return nil
}

// ListEvents returns recent probe events.
func (s *ProbingService) ListEvents(ctx context.Context, limit, offset int) ([]*domain.ProbeEvent, int, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, limit, offset)
}

// ListBlockedIPs returns currently blocked IPs.
func (s *ProbingService) ListBlockedIPs(ctx context.Context) ([]domain.BlockedIP, error) {
	return s.repo.ListBlockedIPs(ctx)
}

// UnblockIP removes an IP from the blocklist.
func (s *ProbingService) UnblockIP(ctx context.Context, ip string) error {
	return s.repo.UnblockIP(ctx, ip)
}

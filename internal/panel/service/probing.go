package service

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ProbingService manages active probing protection.
type ProbingService struct {
	repo   port.ProbingRepository
	log    *slog.Logger
	now    func() time.Time
	resync *FleetResync
	pub    events.Publisher
}

// NewProbingService wires the service.
func NewProbingService(repo port.ProbingRepository, log *slog.Logger) *ProbingService {
	return &ProbingService{repo: repo, log: log, now: time.Now, pub: events.Nop{}}
}

// SetFleetResync triggers node resync after blocklist changes.
func (s *ProbingService) SetFleetResync(f *FleetResync) { s.resync = f }

// SetPublisher emits security probe alerts (e.g. Telegram).
func (s *ProbingService) SetPublisher(p events.Publisher) {
	if p != nil {
		s.pub = p
	}
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
	if policy == nil || !policy.Enabled {
		return nil
	}
	if probeWhitelisted(policy, ip) {
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

	shouldBlock := action == domain.ProbingBlock || action == domain.ProbingHoneypot
	if policy.MaxProbePerMin > 0 {
		since := s.now().Add(-time.Minute)
		count, err := s.repo.CountRecentByIP(ctx, ip, since)
		if err == nil && count < policy.MaxProbePerMin {
			shouldBlock = false
		}
	}
	if action == domain.ProbingLog {
		shouldBlock = false
	}

	if shouldBlock {
		blocked := &domain.BlockedIP{
			IP:        ip,
			Reason:    method,
			BlockedAt: s.now(),
			ExpiresAt: s.now().Add(time.Duration(policy.BlockDuration) * time.Second),
		}
		if err := s.repo.BlockIP(ctx, blocked); err != nil {
			s.log.Error("failed to block IP", "ip", ip, "error", err)
		} else if s.resync != nil {
			_ = s.resync.Node(ctx, nodeID)
		}
	}

	if policy.NotifyTelegram {
		msg := "Probe from " + ip + " (" + method + ")"
		if nodeID != nil {
			msg += " node=" + nodeID.String()
		}
		s.pub.Publish(events.Event{
			Type:    events.SecurityProbe,
			Time:    s.now(),
			NodeID:  nodeIDString(nodeID),
			Message: msg,
			Data: map[string]any{
				"source_ip": ip,
				"method":    method,
				"action":    string(action),
				"blocked":   shouldBlock,
			},
		})
	}

	s.log.Warn("probe detected", "ip", ip, "method", method, "action", action, "blocked", shouldBlock)
	return nil
}

func probeWhitelisted(policy *domain.ProbingPolicy, ip string) bool {
	if policy == nil || ip == "" {
		return false
	}
	parsed := net.ParseIP(ip)
	for _, w := range policy.WhitelistedIPs {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		if w == ip {
			return true
		}
		if _, cidr, err := net.ParseCIDR(w); err == nil && parsed != nil && cidr.Contains(parsed) {
			return true
		}
	}
	return false
}

func nodeIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
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

// BlockIP adds or refreshes a blocked IP using the current policy duration.
func (s *ProbingService) BlockIP(ctx context.Context, ip, reason string) error {
	if ip == "" {
		return nil
	}
	policy, _ := s.GetPolicy(ctx)
	duration := 3600
	if policy != nil && policy.BlockDuration > 0 {
		duration = policy.BlockDuration
	}
	return s.repo.BlockIP(ctx, &domain.BlockedIP{
		IP:        ip,
		Reason:    reason,
		BlockedAt: s.now(),
		ExpiresAt: s.now().Add(time.Duration(duration) * time.Second),
	})
}

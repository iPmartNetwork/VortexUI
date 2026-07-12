package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// FingerprintService manages TLS client fingerprint validation.
type FingerprintService struct {
	repo     port.FingerprintRepository
	probing  *ProbingService
	resync   *FleetResync
	log      *slog.Logger
	now      func() time.Time
}

func NewFingerprintService(repo port.FingerprintRepository) *FingerprintService {
	return &FingerprintService{repo: repo, now: time.Now}
}

// SetProbing wires probe blocking for fingerprint block actions.
func (s *FingerprintService) SetProbing(p *ProbingService) { s.probing = p }

// SetFleetResync triggers node resync after a block action.
func (s *FingerprintService) SetFleetResync(f *FleetResync) { s.resync = f }

// SetLogger attaches structured logging.
func (s *FingerprintService) SetLogger(log *slog.Logger) { s.log = log }

func (s *FingerprintService) GetPolicy(ctx context.Context) (*domain.FingerprintPolicy, error) {
	p, err := s.repo.GetPolicy(ctx)
	if err != nil {
		def := domain.DefaultFingerprintPolicy()
		return &def, nil
	}
	return p, nil
}

func (s *FingerprintService) UpdatePolicy(ctx context.Context, p *domain.FingerprintPolicy) error {
	return s.repo.SavePolicy(ctx, p)
}

func (s *FingerprintService) CreateRule(ctx context.Context, name, fingerprint, ja3, action string, priority int) (*domain.FingerprintRule, error) {
	r := &domain.FingerprintRule{
		ID:          uuid.New(),
		Name:        name,
		Fingerprint: fingerprint,
		JA3Hash:     ja3,
		Action:      domain.FingerprintAction(action),
		Priority:    priority,
		Enabled:     true,
		CreatedAt:   s.now(),
	}
	if err := s.repo.CreateRule(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *FingerprintService) ListRules(ctx context.Context) ([]*domain.FingerprintRule, error) {
	return s.repo.ListRules(ctx)
}

func (s *FingerprintService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRule(ctx, id)
}

func (s *FingerprintService) ListEvents(ctx context.Context, limit int) ([]*domain.FingerprintEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, limit)
}

// ReportInput describes a runtime fingerprint observation.
type ReportInput struct {
	ClientIP    string
	Fingerprint string
	JA3Hash     string
	UserAgent   string
	NodeID      *uuid.UUID
}

// Report evaluates a client fingerprint and optionally blocks the source IP.
func (s *FingerprintService) Report(ctx context.Context, in ReportInput) (domain.FingerprintAction, error) {
	policy, _ := s.GetPolicy(ctx)
	if policy == nil || !policy.Enabled {
		return domain.FingerprintAllow, nil
	}

	action := policy.DefaultAction
	matched := false
	rules, _ := s.repo.ListRules(ctx)
	for _, r := range rules {
		if r == nil || !r.Enabled {
			continue
		}
		if !fingerprintMatches(r, in.Fingerprint, in.JA3Hash) {
			continue
		}
		action = r.Action
		matched = true
		break
	}
	if !matched && policy.LogUnknown && action == domain.FingerprintAllow {
		action = domain.FingerprintLog
	}

	event := &domain.FingerprintEvent{
		ID:          uuid.New(),
		ClientIP:    in.ClientIP,
		Fingerprint: in.Fingerprint,
		JA3Hash:     in.JA3Hash,
		UserAgent:   in.UserAgent,
		Action:      action,
		NodeID:      in.NodeID,
		CreatedAt:   s.now(),
	}
	if err := s.repo.SaveEvent(ctx, event); err != nil && s.log != nil {
		s.log.Error("fingerprint event save failed", "error", err)
	}

	if action == domain.FingerprintBlock && in.ClientIP != "" && s.probing != nil {
		if err := s.probing.BlockIP(ctx, in.ClientIP, "fingerprint"); err != nil && s.log != nil {
			s.log.Error("fingerprint block failed", "ip", in.ClientIP, "error", err)
		}
		if s.resync != nil {
			_ = s.resync.Node(ctx, in.NodeID)
		}
	}

	return action, nil
}

func fingerprintMatches(r *domain.FingerprintRule, fp, ja3 string) bool {
	if ja3 != "" && r.JA3Hash != "" && strings.EqualFold(r.JA3Hash, ja3) {
		return true
	}
	if fp != "" && r.Fingerprint != "" && strings.EqualFold(r.Fingerprint, fp) {
		return true
	}
	return false
}

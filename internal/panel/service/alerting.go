package service

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// AlertingService evaluates threshold-based alert rules against nodes and users
// and publishes events when thresholds are crossed. Cooldown windows prevent
// repeated firing for the same condition.
type AlertingService struct {
	nodes     port.NodeRepository
	users     port.UserRepository
	inbounds  port.InboundRepository
	interval  time.Duration
	now       func() time.Time
	log       *slog.Logger
	pub       events.Publisher
	lastFired map[string]time.Time
}

// NewAlertingService wires the service. Call SetPublisher before Run.
func NewAlertingService(nodes port.NodeRepository, users port.UserRepository, inbounds port.InboundRepository, log *slog.Logger) *AlertingService {
	if log == nil {
		log = slog.Default()
	}
	return &AlertingService{
		nodes:     nodes,
		users:     users,
		inbounds:  inbounds,
		interval:  2 * time.Minute,
		now:       time.Now,
		log:       log,
		pub:       events.Nop{},
		lastFired: make(map[string]time.Time),
	}
}

// SetPublisher wires the event publisher.
func (s *AlertingService) SetPublisher(p events.Publisher) {
	if p != nil {
		s.pub = p
	}
}

// Run evaluates alert rules on a ticker until ctx is cancelled.
func (s *AlertingService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.evaluateAll(ctx)
		}
	}
}

func (s *AlertingService) evaluateAll(ctx context.Context) {
	s.evaluateNodeAlerts(ctx)
	s.evaluateUserQuotaAlerts(ctx)
	s.evaluateCertAlerts(ctx)
	s.pruneOldEntries()
}

// evaluateNodeAlerts checks CPU, memory, and offline thresholds.
func (s *AlertingService) evaluateNodeAlerts(ctx context.Context) {
	nodes, err := s.nodes.List(ctx)
	if err != nil {
		s.log.Warn("alerting: failed to list nodes", "err", err)
		return
	}
	now := s.now()
	for _, n := range nodes {
		// CPU alert: >= 90%
		if n.Health.CPUPercent >= 90 {
			key := "node_cpu_" + n.ID.String()
			if s.shouldFire(key, now, 15*time.Minute) {
				s.fire(key, now, events.Event{
					Type:     events.NodeAlert,
					NodeID:   n.ID.String(),
					NodeName: n.Name,
					Message:  fmt.Sprintf("Node %s CPU at %.0f%% (threshold: 90%%)", n.Name, n.Health.CPUPercent),
					Data:     map[string]any{"metric": "cpu", "value": n.Health.CPUPercent},
				})
			}
		}
		// Memory alert: >= 90%
		if n.Health.MemPercent >= 90 {
			key := "node_mem_" + n.ID.String()
			if s.shouldFire(key, now, 15*time.Minute) {
				s.fire(key, now, events.Event{
					Type:     events.NodeAlert,
					NodeID:   n.ID.String(),
					NodeName: n.Name,
					Message:  fmt.Sprintf("Node %s memory at %.0f%% (threshold: 90%%)", n.Name, n.Health.MemPercent),
					Data:     map[string]any{"metric": "memory", "value": n.Health.MemPercent},
				})
			}
		}
		// Offline alert: last seen > 5 minutes ago
		if n.LastSeen != nil && now.Sub(*n.LastSeen) > 5*time.Minute {
			key := "node_offline_" + n.ID.String()
			if s.shouldFire(key, now, 30*time.Minute) {
				s.fire(key, now, events.Event{
					Type:     events.NodeDown,
					NodeID:   n.ID.String(),
					NodeName: n.Name,
					Message:  fmt.Sprintf("Node %s offline for %s", n.Name, now.Sub(*n.LastSeen).Round(time.Minute)),
					Data:     map[string]any{"metric": "offline", "last_seen": n.LastSeen.Format(time.RFC3339)},
				})
			}
		}
	}
}

// evaluateUserQuotaAlerts checks users approaching their data limit (>= 80%).
func (s *AlertingService) evaluateUserQuotaAlerts(ctx context.Context) {
	users, _, err := s.users.List(ctx, port.UserFilter{Limit: 10000})
	if err != nil {
		s.log.Warn("alerting: failed to list users", "err", err)
		return
	}
	now := s.now()
	for _, u := range users {
		if u.DataLimit <= 0 {
			continue // unlimited
		}
		pct := float64(u.UsedTraffic) / float64(u.DataLimit) * 100
		if pct >= 80 {
			key := "user_quota_" + u.ID.String()
			if s.shouldFire(key, now, 6*time.Hour) {
				s.fire(key, now, events.Event{
					Type:     events.UserQuotaWarn,
					UserID:   u.ID.String(),
					Username: u.Username,
					Message:  fmt.Sprintf("User %s traffic at %.0f%% of data limit", u.Username, pct),
					Data:     map[string]any{"usage_pct": pct, "used": u.UsedTraffic, "limit": u.DataLimit},
				})
			}
		}
	}
}

// evaluateCertAlerts checks inbound TLS certificates expiring within 7 days.
func (s *AlertingService) evaluateCertAlerts(ctx context.Context) {
	if s.inbounds == nil {
		return
	}
	nodes, err := s.nodes.List(ctx)
	if err != nil {
		return
	}
	now := s.now()
	for _, n := range nodes {
		inbounds, err := s.inbounds.ListByNode(ctx, n.ID)
		if err != nil {
			continue
		}
		for _, in := range inbounds {
			if in.Security != domain.SecurityTLS {
				continue
			}
			certPEM := extractCertPEM(in.Raw)
			if certPEM == "" {
				continue
			}
			daysLeft := parseCertDaysRemaining(certPEM, now)
			if daysLeft < 0 {
				continue // parse failure
			}
			if daysLeft <= 7 {
				key := "cert_expiring_" + in.ID.String()
				if s.shouldFire(key, now, 24*time.Hour) {
					s.fire(key, now, events.Event{
						Type:     events.CertExpiring,
						NodeID:   n.ID.String(),
						NodeName: n.Name,
						Message:  fmt.Sprintf("TLS certificate on node %s (inbound %s) expires in %d days", n.Name, in.Tag, daysLeft),
						Data:     map[string]any{"inbound_id": in.ID.String(), "days_remaining": daysLeft},
					})
				}
			}
		}
	}
}

// shouldFire returns true if the alert identified by key has not fired within
// the cooldown period.
func (s *AlertingService) shouldFire(key string, now time.Time, cooldown time.Duration) bool {
	last, ok := s.lastFired[key]
	return !ok || now.Sub(last) >= cooldown
}

// fire records the firing time and publishes the event.
func (s *AlertingService) fire(key string, now time.Time, evt events.Event) {
	s.lastFired[key] = now
	s.pub.Publish(evt)
	s.log.Warn("alert fired", "type", string(evt.Type), "message", evt.Message)
}

// pruneOldEntries removes lastFired entries older than 48 hours to prevent
// unbounded memory growth in long-running panels.
func (s *AlertingService) pruneOldEntries() {
	now := s.now()
	for k, t := range s.lastFired {
		if now.Sub(t) > 48*time.Hour {
			delete(s.lastFired, k)
		}
	}
}

// parseCertDaysRemaining parses a PEM certificate and returns days until expiry,
// or -1 on any parse error.
func parseCertDaysRemaining(certPEM string, now time.Time) int {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return -1
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return -1
	}
	return int(cert.NotAfter.Sub(now).Hours() / 24)
}

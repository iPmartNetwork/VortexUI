// Package security detects probing and fingerprint anomalies from core logs.
package security

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

var (
	probeLine = regexp.MustCompile(`(?i)(reject|invalid|failed|handshake|unauthorized|denied)`)
	ipv4      = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
)

// Monitor tails core logs and emits security events.
type Monitor struct {
	driver core.CoreDriver

	mu      sync.Mutex
	seen    map[string]struct{}
	seenCap int
}

// NewMonitor wires a log-based security monitor.
func NewMonitor(driver core.CoreDriver) *Monitor {
	return &Monitor{driver: driver, seen: make(map[string]struct{}), seenCap: 4096}
}

// Run polls driver logs until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) <-chan domain.SecurityEvent {
	out := make(chan domain.SecurityEvent, 32)
	go func() {
		defer close(out)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.poll(ctx, out)
			}
		}
	}()
	return out
}

func (m *Monitor) poll(ctx context.Context, out chan<- domain.SecurityEvent) {
	if m.driver == nil {
		return
	}
	lines, err := m.driver.Logs(ctx, 200)
	if err != nil {
		return
	}
	for _, line := range lines {
		ev, ok := parseLine(line)
		if !ok || m.duplicate(ev) {
			continue
		}
		select {
		case out <- ev:
		default:
		}
	}
}

func (m *Monitor) duplicate(ev domain.SecurityEvent) bool {
	key := ev.SourceIP + "|" + ev.Method + "|" + ev.Fingerprint
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.seen[key]; ok {
		return true
	}
	if len(m.seen) >= m.seenCap {
		m.seen = make(map[string]struct{}, m.seenCap/2)
	}
	m.seen[key] = struct{}{}
	return false
}

// ParseLogLine extracts a security event from one core log line.
func ParseLogLine(line string) (domain.SecurityEvent, bool) {
	return parseLine(line)
}

func parseLine(line string) (domain.SecurityEvent, bool) {
	lower := strings.ToLower(line)
	if !probeLine.MatchString(lower) {
		return domain.SecurityEvent{}, false
	}
	ipMatch := ipv4.FindStringSubmatch(line)
	if len(ipMatch) < 2 {
		return domain.SecurityEvent{}, false
	}
	method := "tls_probe"
	switch {
	case strings.Contains(lower, "http"):
		method = "http_probe"
	case strings.Contains(lower, "fingerprint"), strings.Contains(lower, "ja3"):
		method = "fingerprint"
	}
	fp := ""
	if idx := strings.Index(lower, "ja3"); idx >= 0 {
		fp = strings.TrimSpace(line[idx:])
	}
	return domain.SecurityEvent{
		SourceIP:    ipMatch[1],
		Method:      method,
		Fingerprint: fp,
	}, true
}

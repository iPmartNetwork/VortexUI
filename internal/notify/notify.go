// Package notify delivers domain events to external destinations: an HTTP
// webhook and a Telegram chat. Each destination is a bus subscriber that runs in
// its own goroutine and performs network I/O off the panel's hot path.
package notify

import (
	"fmt"
	"time"

	"github.com/vortexui/vortexui/internal/events"
)

// describe renders an event as a short human-readable line for chat messages.
func describe(e events.Event) string {
	switch e.Type {
	case events.UserCreated:
		return fmt.Sprintf("👤 User created: %s", e.Username)
	case events.UserDeleted:
		return fmt.Sprintf("🗑 User deleted: %s", e.Username)
	case events.UserLimited:
		return fmt.Sprintf("🚫 User reached data limit: %s", e.Username)
	case events.UserExpired:
		return fmt.Sprintf("⏰ User expired: %s", e.Username)
	case events.UserReset:
		return fmt.Sprintf("🔄 User traffic reset: %s", e.Username)
	case events.UserExpiryWarning:
		days := e.Data["days_left"]
		return fmt.Sprintf("⚠️ User expiring in %v days: %s", days, e.Username)
	case events.UserIPLimit:
		ips, limit := e.Data["online_ips"], e.Data["device_limit"]
		if limited, _ := e.Data["limited"].(bool); limited {
			return fmt.Sprintf("👥 Account sharing — %s limited (%v IPs > %v devices)", e.Username, ips, limit)
		}
		return fmt.Sprintf("👥 Account sharing detected: %s (%v IPs > %v devices)", e.Username, ips, limit)
	case events.NodeDown:
		return fmt.Sprintf("🔴 Node down: %s", e.NodeName)
	case events.NodeUp:
		return fmt.Sprintf("🟢 Node up: %s", e.NodeName)
	case events.NodeDisconnectAlert:
		mins, _ := e.Data["since_minutes"].(float64)
		code, _ := e.Data["diagnostics"].(string)
		if e.Message != "" {
			return fmt.Sprintf("🔴 Node disconnected >%.0fm: %s — %s (%s)", mins, e.NodeName, e.Message, code)
		}
		return fmt.Sprintf("🔴 Node disconnected >%.0fm: %s (%s)", mins, e.NodeName, code)
	case events.AdminQuotaWarning:
		admin, _ := e.Data["admin"].(string)
		metric, _ := e.Data["metric"].(string)
		pct := e.Data["usage_pct"]
		th := e.Data["threshold"]
		return fmt.Sprintf("📊 Reseller quota alert: %s — %s at %v%% (threshold %v%%)", admin, metric, pct, th)
	case events.SecurityProbe:
		ip, _ := e.Data["source_ip"].(string)
		method, _ := e.Data["method"].(string)
		return fmt.Sprintf("🛡 Probe detected: %s (%s) on %s", ip, method, e.NodeName)
	default:
		msg := e.Message
		if msg == "" {
			msg = string(e.Type)
		}
		return fmt.Sprintf("ℹ️ %s", msg)
	}
}

// stamp formats the event time for messages.
func stamp(t time.Time) string { return t.UTC().Format("2006-01-02 15:04:05 UTC") }

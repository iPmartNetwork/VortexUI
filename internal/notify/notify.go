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
	case events.NodeDown:
		return fmt.Sprintf("🔴 Node down: %s", e.NodeName)
	case events.NodeUp:
		return fmt.Sprintf("🟢 Node up: %s", e.NodeName)
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

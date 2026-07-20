// Package events is a tiny in-process publish/subscribe bus for domain events
// (a user hitting its limit, a node going down, ...). Publishers stay decoupled
// from consumers like the webhook dispatcher and the Telegram notifier, and slow
// consumers never block the publisher: fan-out is non-blocking per subscriber.
package events

import (
	"log/slog"
	"sync"
	"time"
)

// Type identifies what happened. Values are stable strings so they can be sent
// verbatim to webhooks and rendered in messages.
type Type string

const (
	UserCreated Type = "user.created"
	UserDeleted Type = "user.deleted"
	UserLimited Type = "user.limited" // crossed the data cap
	UserExpired Type = "user.expired" // passed expire_at
	UserReset   Type = "user.reset"   // traffic counter reset
	// UserIPLimit fires when a user is online from more distinct source IPs than
	// its device limit allows — a strong signal of account sharing.
	UserIPLimit Type = "user.ip_limit"
	// UserExpiryWarning fires N days before a user's subscription expires.
	UserExpiryWarning Type = "user.expiry_warning"
	NodeDown          Type = "node.down" // became unhealthy/unreachable
	NodeUp            Type = "node.up"   // (re)connected
	NodeDisconnectAlert Type = "node.disconnect_alert" // unreachable > 5 min
	NodeAutoRecover   Type = "node.auto_recover" // panel restarted core or reset hub link
	AdminQuotaWarning Type = "admin.quota_warning"
	SecurityProbe     Type = "security.probe"
	NodeAlert         Type = "node.alert"         // resource threshold crossed
	UserQuotaWarn     Type = "user.quota_warning"  // traffic nearing data_limit
	CertExpiring      Type = "cert.expiring"       // TLS cert expires within 7 days
	ProtocolSwitch    Type = "protocol.switch"     // client switched protocol within a group
)

// Event is a single notification. Fields are flat and JSON-friendly so the
// webhook payload and message templates stay simple; Data carries any extras.
type Event struct {
	Type     Type           `json:"type"`
	Time     time.Time      `json:"time"`
	UserID   string         `json:"user_id,omitempty"`
	Username string         `json:"username,omitempty"`
	NodeID   string         `json:"node_id,omitempty"`
	NodeName string         `json:"node_name,omitempty"`
	Message  string         `json:"message,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}

// Publisher is the narrow interface producers depend on. *Bus satisfies it; a
// nil publisher is handled by Nop so callers never need a nil check.
type Publisher interface {
	Publish(e Event)
}

// Bus fans events out to every subscriber. Subscribers receive on their own
// buffered channel; if a subscriber is too slow and its buffer fills, the event
// is dropped for that subscriber (and logged) rather than stalling the panel.
type Bus struct {
	mu   sync.RWMutex
	subs []chan Event
	log  *slog.Logger
}

// New builds a Bus.
func New(log *slog.Logger) *Bus {
	if log == nil {
		log = slog.Default()
	}
	return &Bus{log: log}
}

// Subscribe registers a consumer and returns its event channel. buffer sizes the
// per-subscriber queue; pick it large enough to absorb bursts during slow I/O.
// Long-lived but transient consumers (e.g. an SSE connection) must call
// Unsubscribe when done, or the channel leaks.
func (b *Bus) Subscribe(buffer int) <-chan Event {
	if buffer <= 0 {
		buffer = 64
	}
	ch := make(chan Event, buffer)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes and closes a channel returned by Subscribe. Safe to call
// once per channel; unknown channels are ignored.
func (b *Bus) Unsubscribe(ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, c := range b.subs {
		if (<-chan Event)(c) == ch {
			b.subs = append(b.subs[:i], b.subs[i+1:]...)
			close(c)
			return
		}
	}
}

// Publish delivers e to every subscriber without blocking. The event time is
// stamped if unset. A full subscriber queue drops the event for that subscriber.
func (b *Bus) Publish(e Event) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
			b.log.Warn("event dropped: subscriber queue full", "type", string(e.Type))
		}
	}
}

// Nop is a publisher that discards events, used as a safe default when no bus is
// wired (e.g. in tests or a minimal deployment).
type Nop struct{}

// Publish implements Publisher.
func (Nop) Publish(Event) {}

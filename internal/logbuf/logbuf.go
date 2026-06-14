// Package logbuf provides an slog.Handler that keeps the most recent log records
// in a fixed-size in-memory ring buffer while still delegating to an inner
// handler (e.g. JSON to stdout). It lets the panel expose its recent logs over
// the API without a separate log store.
package logbuf

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Entry is one captured log record in an API-friendly shape. Level marshals to
// its string name ("INFO", "WARN", ...) via slog.Level's own JSON encoding.
type Entry struct {
	Time    time.Time      `json:"time"`
	Level   slog.Level     `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

// ring is a fixed-capacity circular buffer of entries, safe for concurrent use.
type ring struct {
	mu   sync.Mutex
	buf  []Entry
	next int
	full bool
}

func newRing(capacity int) *ring {
	if capacity <= 0 {
		capacity = 1000
	}
	return &ring{buf: make([]Entry, capacity)}
}

func (r *ring) add(e Entry) {
	r.mu.Lock()
	r.buf[r.next] = e
	r.next = (r.next + 1) % len(r.buf)
	if r.next == 0 {
		r.full = true
	}
	r.mu.Unlock()
}

// snapshot returns entries oldest→newest, filtered to level >= minLevel, capped
// to the most recent `limit` matches (limit <= 0 means no cap).
func (r *ring) snapshot(minLevel slog.Level, limit int) []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	var ordered []Entry
	if r.full {
		ordered = append(ordered, r.buf[r.next:]...)
		ordered = append(ordered, r.buf[:r.next]...)
	} else {
		ordered = append(ordered, r.buf[:r.next]...)
	}

	out := make([]Entry, 0, len(ordered))
	for _, e := range ordered {
		if e.Level >= minLevel {
			out = append(out, e)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

// Handler is an slog.Handler that records into a shared ring buffer and forwards
// to an inner handler. Derived handlers (WithAttrs/WithGroup) share the buffer.
type Handler struct {
	inner  slog.Handler
	ring   *ring
	attrs  []slog.Attr
	groups []string
}

// New wraps inner, capturing the most recent `capacity` records in memory.
func New(inner slog.Handler, capacity int) *Handler {
	return &Handler{inner: inner, ring: newRing(capacity)}
}

// Enabled defers to the inner handler.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle records the entry (with accumulated attrs) then forwards to inner.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	e := Entry{Time: r.Time, Level: r.Level, Message: r.Message}
	attrs := make(map[string]any, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	if len(attrs) > 0 {
		e.Attrs = attrs
	}
	h.ring.add(e)
	return h.inner.Handle(ctx, r)
}

// WithAttrs returns a handler that adds attrs, sharing the same ring buffer.
func (h *Handler) WithAttrs(as []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(as))
	merged = append(merged, h.attrs...)
	merged = append(merged, as...)
	return &Handler{inner: h.inner.WithAttrs(as), ring: h.ring, attrs: merged, groups: h.groups}
}

// WithGroup returns a handler in the named group, sharing the same ring buffer.
func (h *Handler) WithGroup(name string) slog.Handler {
	groups := make([]string, 0, len(h.groups)+1)
	groups = append(groups, h.groups...)
	groups = append(groups, name)
	return &Handler{inner: h.inner.WithGroup(name), ring: h.ring, attrs: h.attrs, groups: groups}
}

// Entries returns the recent captured records (oldest→newest) at or above
// minLevel, capped to the most recent `limit`.
func (h *Handler) Entries(minLevel slog.Level, limit int) []Entry {
	return h.ring.snapshot(minLevel, limit)
}

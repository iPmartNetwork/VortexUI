package logbuf

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func newTestLogger(capacity int) (*slog.Logger, *Handler) {
	h := New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}), capacity)
	return slog.New(h), h
}

func TestCapturesRecentEntriesOldestToNewest(t *testing.T) {
	log, h := newTestLogger(10)
	log.Info("first")
	log.Warn("second")
	log.Error("third")

	entries := h.Entries(slog.LevelDebug, 0)
	if len(entries) != 3 {
		t.Fatalf("want 3 entries, got %d", len(entries))
	}
	if entries[0].Message != "first" || entries[2].Message != "third" {
		t.Errorf("order wrong: %s ... %s", entries[0].Message, entries[2].Message)
	}
}

func TestRingEvictsOldest(t *testing.T) {
	log, h := newTestLogger(3)
	for _, m := range []string{"a", "b", "c", "d", "e"} {
		log.Info(m)
	}
	entries := h.Entries(slog.LevelDebug, 0)
	if len(entries) != 3 {
		t.Fatalf("want 3 (capacity), got %d", len(entries))
	}
	// Oldest two evicted -> c, d, e remain.
	if entries[0].Message != "c" || entries[2].Message != "e" {
		t.Errorf("ring eviction wrong: %v", []string{entries[0].Message, entries[1].Message, entries[2].Message})
	}
}

func TestLevelFilterAndLimit(t *testing.T) {
	log, h := newTestLogger(10)
	log.Info("i1")
	log.Warn("w1")
	log.Info("i2")
	log.Error("e1")

	// Only WARN+ .
	warn := h.Entries(slog.LevelWarn, 0)
	if len(warn) != 2 || warn[0].Message != "w1" || warn[1].Message != "e1" {
		t.Errorf("level filter wrong: %+v", warn)
	}
	// Limit to most recent 1 overall.
	last := h.Entries(slog.LevelDebug, 1)
	if len(last) != 1 || last[0].Message != "e1" {
		t.Errorf("limit wrong: %+v", last)
	}
}

func TestAttrsCaptured(t *testing.T) {
	log, h := newTestLogger(10)
	log.Info("with attrs", "node", "de-1", "count", 3)
	e := h.Entries(slog.LevelDebug, 0)
	if len(e) != 1 {
		t.Fatalf("want 1, got %d", len(e))
	}
	if e[0].Attrs["node"] != "de-1" {
		t.Errorf("attr node = %v, want de-1", e[0].Attrs["node"])
	}
}

func TestWithAttrsSharesBuffer(t *testing.T) {
	log, h := newTestLogger(10)
	child := log.With("component", "hub")
	child.Info("child msg")
	_ = context.Background()

	e := h.Entries(slog.LevelDebug, 0)
	if len(e) != 1 || e[0].Attrs["component"] != "hub" {
		t.Errorf("WithAttrs not captured into shared buffer: %+v", e)
	}
}

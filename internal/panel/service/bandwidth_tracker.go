package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// BandwidthSample is one per-second bandwidth measurement for a user.
type BandwidthSample struct {
	UserID    uuid.UUID `json:"user_id"`
	Upload    int64     `json:"upload"`    // bytes/sec
	Download  int64     `json:"download"`  // bytes/sec
	Timestamp time.Time `json:"timestamp"`
}

// BandwidthTracker maintains a sliding window of per-user bandwidth measurements.
// It's fed by the stats aggregator and consumed by the live dashboard.
type BandwidthTracker struct {
	mu      sync.RWMutex
	samples map[uuid.UUID][]BandwidthSample
	window  time.Duration // how long to keep samples (e.g. 5 minutes)
}

func NewBandwidthTracker(window time.Duration) *BandwidthTracker {
	if window == 0 {
		window = 5 * time.Minute
	}
	return &BandwidthTracker{
		samples: make(map[uuid.UUID][]BandwidthSample),
		window:  window,
	}
}

// Record adds a bandwidth sample for a user.
func (t *BandwidthTracker) Record(userID uuid.UUID, upload, download int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.samples[userID] = append(t.samples[userID], BandwidthSample{
		UserID: userID, Upload: upload, Download: download, Timestamp: now,
	})
	t.prune(userID, now)
}

// Get returns recent bandwidth samples for a user.
func (t *BandwidthTracker) Get(userID uuid.UUID) []BandwidthSample {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.samples[userID]
}

// Current returns the latest bandwidth for a user (bytes/sec up + down).
func (t *BandwidthTracker) Current(userID uuid.UUID) (upload, download int64) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	samples := t.samples[userID]
	if len(samples) == 0 {
		return 0, 0
	}
	latest := samples[len(samples)-1]
	return latest.Upload, latest.Download
}

func (t *BandwidthTracker) prune(userID uuid.UUID, now time.Time) {
	cutoff := now.Add(-t.window)
	samples := t.samples[userID]
	i := 0
	for i < len(samples) && samples[i].Timestamp.Before(cutoff) {
		i++
	}
	if i > 0 {
		t.samples[userID] = samples[i:]
	}
}

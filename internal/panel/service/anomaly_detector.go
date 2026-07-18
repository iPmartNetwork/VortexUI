package service

import (
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/events"
)

// AnomalyType classifies the kind of anomaly detected.
type AnomalyType string

const (
	AnomalyTrafficSpike  AnomalyType = "traffic_spike"
	AnomalyLoginBurst    AnomalyType = "login_burst"
	AnomalyNodeFlap      AnomalyType = "node_flap"
	AnomalyUserSurge     AnomalyType = "user_surge"
	AnomalyBandwidthDrop AnomalyType = "bandwidth_drop"
)

// Anomaly represents a detected abnormal pattern.
type Anomaly struct {
	Type       AnomalyType `json:"type"`
	Severity   string      `json:"severity"` // "low", "medium", "high", "critical"
	Message    string      `json:"message"`
	Value      float64     `json:"value"`     // observed value
	Threshold  float64     `json:"threshold"` // expected threshold
	DetectedAt time.Time   `json:"detected_at"`
}

// AnomalyDetector uses statistical methods to detect unusual patterns.
type AnomalyDetector struct {
	mu         sync.RWMutex
	baselines  map[string][]float64 // key → historical values for mean/stddev
	log        *slog.Logger
	pub        events.Publisher
	maxHistory int
}

func NewAnomalyDetector(log *slog.Logger) *AnomalyDetector {
	if log == nil {
		log = slog.Default()
	}
	return &AnomalyDetector{
		baselines:  make(map[string][]float64),
		log:        log,
		pub:        events.Nop{},
		maxHistory: 1000,
	}
}

func (d *AnomalyDetector) SetPublisher(p events.Publisher) {
	if p != nil {
		d.pub = p
	}
}

// Observe records a metric value and checks for anomalies using z-score.
// Returns an anomaly if the value deviates significantly from the baseline.
func (d *AnomalyDetector) Observe(key string, value float64, anomalyType AnomalyType) *Anomaly {
	d.mu.Lock()
	defer d.mu.Unlock()

	history := d.baselines[key]
	history = append(history, value)
	if len(history) > d.maxHistory {
		history = history[len(history)-d.maxHistory:]
	}
	d.baselines[key] = history

	// Need at least 10 samples for meaningful statistics.
	if len(history) < 10 {
		return nil
	}

	mean, stddev := meanStdDev(history[:len(history)-1]) // exclude current
	if stddev == 0 {
		return nil
	}

	zScore := (value - mean) / stddev
	if zScore < 3.0 { // 3 sigma threshold
		return nil
	}

	severity := "low"
	if zScore > 5 {
		severity = "high"
	} else if zScore > 4 {
		severity = "medium"
	}

	return &Anomaly{
		Type:       anomalyType,
		Severity:   severity,
		Value:      value,
		Threshold:  mean + 3*stddev,
		DetectedAt: time.Now(),
	}
}

func meanStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	sumSq := 0.0
	for _, v := range values {
		sumSq += (v - mean) * (v - mean)
	}
	stddev := math.Sqrt(sumSq / float64(len(values)))
	return mean, stddev
}

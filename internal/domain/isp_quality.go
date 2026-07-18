package domain

import "time"

// ISPQualityEntry holds connection quality metrics for an ISP at a specific hour.
type ISPQualityEntry struct {
	ISP         string    `json:"isp"`
	Hour        int       `json:"hour"`         // 0-23
	DayOfWeek   int       `json:"day_of_week"`  // 0=Sunday, 6=Saturday
	AvgLatency  float64   `json:"avg_latency"`  // ms
	SuccessRate float64   `json:"success_rate"` // 0.0-1.0
	Samples     int       `json:"samples"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ISPQualityHeatmap holds a full week of quality data for visualization.
type ISPQualityHeatmap struct {
	ISP     string            `json:"isp"`
	Entries []ISPQualityEntry `json:"entries"` // 7 days × 24 hours = up to 168 entries
}

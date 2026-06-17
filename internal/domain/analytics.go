package domain

import "time"

// GeoTrafficPoint is an aggregated traffic sample by country.
type GeoTrafficPoint struct {
	Time        time.Time `json:"time"`
	Country     string    `json:"country"`
	Connections int       `json:"connections"`
	BytesUp     int64     `json:"bytes_up"`
	BytesDown   int64     `json:"bytes_down"`
}

// UserTrafficRank represents a top user by bandwidth usage.
type UserTrafficRank struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	UsedTraffic int64  `json:"used_traffic"`
}

// PeakHour represents traffic aggregated by hour-of-day.
type PeakHour struct {
	Hour        int   `json:"hour"` // 0-23
	Connections int64 `json:"connections"`
	BytesTotal  int64 `json:"bytes_total"`
}

// AnalyticsOverview wraps the analytics response.
type AnalyticsOverview struct {
	GeoBreakdown []GeoTrafficPoint `json:"geo_breakdown"`
	TopUsers     []UserTrafficRank `json:"top_users"`
	PeakHours    []PeakHour        `json:"peak_hours"`
	TotalUp      int64             `json:"total_up"`
	TotalDown    int64             `json:"total_down"`
}

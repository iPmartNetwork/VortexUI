package domain

import (
	"time"

	"github.com/google/uuid"
)

// DailyCheckWidget aggregates the morning daily-check data:
// node health, traffic anomalies, certificate status, and diagnostic cards.
type DailyCheckWidget struct {
	NodesOnline    int                `json:"nodes_online"`
	NodesTotal     int                `json:"nodes_total"`
	TrafficAnomaly bool               `json:"traffic_anomaly"`
	CertStatus     []CertHealthStatus `json:"cert_status"`
	Diagnostics    []DiagnosticCard   `json:"diagnostics"`
}

// DiagnosticCard represents an actionable diagnostic item surfaced by
// anomaly detection (e.g., offline nodes, expiring certs, traffic spikes).
type DiagnosticCard struct {
	Severity    string   `json:"severity"` // "critical" | "warning" | "info"
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Actions     []string `json:"actions"`
}

// CertHealthStatus reports TLS certificate validity for a domain.
type CertHealthStatus struct {
	Domain    string    `json:"domain"`
	ExpiresAt time.Time `json:"expires_at"`
	Valid     bool      `json:"valid"`
}

// ISPHeatmap represents a 7-day x 24-hour quality heatmap for a single ISP.
type ISPHeatmap struct {
	ISP   string        `json:"isp"`
	Cells []HeatmapCell `json:"cells"`
}

// HeatmapCell is a single data point in the ISP heatmap grid.
type HeatmapCell struct {
	Day   int     `json:"day"`  // 0-6 (Sun-Sat)
	Hour  int     `json:"hour"` // 0-23
	Score float64 `json:"score"`
}

// GeoNode represents a node with geographic coordinates for map visualization.
type GeoNode struct {
	NodeID uuid.UUID `json:"node_id"`
	Name   string    `json:"name"`
	Lat    float64   `json:"lat"`
	Lng    float64   `json:"lng"`
	Status string    `json:"status"`
}

// RevenueReport aggregates financial data across all admins with time series.
type RevenueReport struct {
	TotalIncome  int64              `json:"total_income"`
	TotalExpense int64              `json:"total_expense"`
	Profit       int64              `json:"profit"`
	ByAdmin      []AdminRevenue     `json:"by_admin"`
	TimeSeries   []RevenueDataPoint `json:"time_series"`
}

// AdminRevenue breaks down income/expense for a specific admin (reseller).
type AdminRevenue struct {
	AdminID   uuid.UUID `json:"admin_id"`
	AdminName string    `json:"admin_name"`
	Income    int64     `json:"income"`
	Expense   int64     `json:"expense"`
}

// RevenueDataPoint is a single date entry in the revenue time series.
type RevenueDataPoint struct {
	Date    string `json:"date"` // YYYY-MM-DD
	Income  int64  `json:"income"`
	Expense int64  `json:"expense"`
}

// SubAnalyticsReport aggregates subscription fetch analytics by format, ISP, and time.
type SubAnalyticsReport struct {
	ByFormat []FormatCount `json:"by_format"`
	ByISP    []ISPCount    `json:"by_isp"`
	ByTime   []TimeCount   `json:"by_time"`
}

// FormatCount counts subscription fetches by output format.
type FormatCount struct {
	Format string `json:"format"`
	Count  int    `json:"count"`
}

// ISPCount counts subscription fetches by ISP name.
type ISPCount struct {
	ISP   string `json:"isp"`
	Count int    `json:"count"`
}

// TimeCount counts subscription fetches by hour of day.
type TimeCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// RevenueEntry represents a single financial entry (income or expense).
type RevenueEntry struct {
	ID          uuid.UUID `json:"id"`
	AdminID     uuid.UUID `json:"admin_id"`
	Type        string    `json:"type"` // "income" | "expense"
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

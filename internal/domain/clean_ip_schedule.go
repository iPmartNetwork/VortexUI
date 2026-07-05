package domain

import "time"

// CleanIPSchedule configures an unattended, recurring Clean-IP scan: when
// enabled, the panel re-runs a scan over the saved candidate IPs on the
// configured cadence, so operators don't have to remember to re-scan by
// hand. It is a single-row config (id=1 in storage).
type CleanIPSchedule struct {
	Enabled         bool       `json:"enabled"`
	IntervalMinutes int        `json:"interval_minutes"`
	Port            int        `json:"port"`
	IPs             []string   `json:"ips"`
	LastRunAt       *time.Time `json:"last_run_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// DefaultCleanIPSchedule returns the built-in defaults: disabled, 6-hour
// cadence, port 443, no candidates configured yet.
func DefaultCleanIPSchedule() CleanIPSchedule {
	return CleanIPSchedule{IntervalMinutes: 360, Port: 443}
}

package domain

import "time"

// DashboardWidgets powers the command-tower overview and sidebar badges.
type DashboardWidgets struct {
	NavBadges NavBadges          `json:"nav_badges"`
	Trends    DashboardTrends    `json:"trends"`
	Probing   ProbingWidget      `json:"probing"`
	Routing   RoutingWidget      `json:"routing"`
	Protocols []ProtocolStat     `json:"protocols"`
	Telemetry *TelemetryWidget   `json:"telemetry,omitempty"`
	NodeFleet []NodeFleetRow     `json:"node_fleet"`
	TopUsers  []OverviewUserRow  `json:"top_users"`
}

// NavBadges feeds sidebar notification counts.
type NavBadges struct {
	ActiveUsers   int `json:"active_users"`
	OpenTickets   int `json:"open_tickets"`
	PendingOrders int `json:"pending_orders"`
}

// DashboardTrends compares the last 24h window to the prior 24h.
type DashboardTrends struct {
	UsersPct     float64 `json:"users_pct"`
	BandwidthPct float64 `json:"bandwidth_pct"`
	SessionsPct  float64 `json:"sessions_pct"`
}

// ProbingWidget summarises anti-DPI / probing protection.
type ProbingWidget struct {
	Enabled         bool `json:"enabled"`
	BlockedScanners int  `json:"blocked_scanners"`
	Events24h       int  `json:"events_24h"`
}

// RoutingWidget summarises smart routing configuration.
type RoutingWidget struct {
	ActiveRules  int `json:"active_rules"`
	RoutingPacks int `json:"routing_packs"`
	Balancers    int `json:"balancers"`
	Inbounds     int `json:"inbounds"`
}

// ProtocolStat is one slice of the protocol breakdown donut.
type ProtocolStat struct {
	Label   string `json:"label"`
	Count   int    `json:"count"`
	Percent int    `json:"percent"`
}

// TelemetryWidget is the live header/fleet strip for the primary node.
type TelemetryWidget struct {
	NodeName    string  `json:"node_name"`
	Core        string  `json:"core"`
	Connections int     `json:"connections"`
	CPUPercent  float64 `json:"cpu_percent"`
	Online      bool    `json:"online"`
	Location    string  `json:"location,omitempty"`
	PingMs      int     `json:"ping_ms,omitempty"`
}

// NodeFleetRow is one node in the overview fleet telemetry panel.
type NodeFleetRow struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Core         string  `json:"core"`
	Location     string  `json:"location"`
	CountryCode  string  `json:"country_code,omitempty"`
	PingMs       int     `json:"ping_ms"`
	UsersCount   int     `json:"users_count"`
	Connections  int     `json:"connections"`
	CPUPercent   float64 `json:"cpu_percent"`
	MemPercent   float64 `json:"mem_percent"`
	Online       bool    `json:"online"`
	Status       string  `json:"status"` // active | warning | inactive
}

// OverviewUserRow is a top consumer row on the overview dashboard.
type OverviewUserRow struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	ProtocolLabel string     `json:"protocol_label"`
	ExpireAt      *time.Time `json:"expire_at,omitempty"`
	UsedTraffic   int64      `json:"used_traffic"`
	DataLimit     int64      `json:"data_limit"`
	Status        UserStatus `json:"status"`
}

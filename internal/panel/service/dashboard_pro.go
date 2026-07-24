package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NodeLiveStatus provides live node status from the Hub.
// The hub.Hub type satisfies this interface.
type NodeLiveStatus interface {
	Live(id uuid.UUID) (domain.NodeStatus, domain.NodeHealth, domain.NodeDiagnostics, bool)
}

// DashboardProService provides advanced dashboard analytics including daily
// checks, ISP heatmaps, geographic node visualization, revenue reporting,
// and subscription analytics.
type DashboardProService struct {
	nodes        port.NodeRepository
	users        port.UserRepository
	ispQuality   port.ISPQualityRepository
	subAnalytics port.SubscriptionAnalyticsRepository
	revenue      port.RevenueRepository
	hub          NodeLiveStatus // optional; enriches node status with live data
}

// NewDashboardProService constructs a DashboardProService with all required
// repository dependencies.
func NewDashboardProService(
	nodes port.NodeRepository,
	users port.UserRepository,
	ispQuality port.ISPQualityRepository,
	subAnalytics port.SubscriptionAnalyticsRepository,
	revenue port.RevenueRepository,
) *DashboardProService {
	return &DashboardProService{
		nodes:        nodes,
		users:        users,
		ispQuality:   ispQuality,
		subAnalytics: subAnalytics,
		revenue:      revenue,
	}
}

// SetHub attaches the live node status provider (Hub) after construction.
func (s *DashboardProService) SetHub(h NodeLiveStatus) {
	s.hub = h
}

// DailyCheck aggregates the morning daily-check data: node health, traffic
// anomalies, certificate status, and diagnostic cards for actionable issues.
func (s *DashboardProService) DailyCheck(ctx context.Context) (*domain.DailyCheckWidget, error) {
	nodes, err := s.nodes.List(ctx)
	if err != nil {
		return nil, err
	}

	widget := &domain.DailyCheckWidget{}
	widget.NodesTotal = len(nodes)

	var offlineNodes []*domain.Node
	var certIssues []domain.CertHealthStatus

	now := time.Now()
	for _, n := range nodes {
		// Use live status from Hub if available
		status := n.Status
		if s.hub != nil {
			if liveStatus, _, _, ok := s.hub.Live(n.ID); ok {
				status = liveStatus
			}
		}

		if status == domain.NodeConnected {
			widget.NodesOnline++
		} else {
			offlineNodes = append(offlineNodes, n)
		}

		// Check for stale nodes (no heartbeat in 5 minutes) as traffic anomaly signal.
		if n.LastSeen != nil && now.Sub(*n.LastSeen) > 5*time.Minute && status == domain.NodeConnected {
			widget.TrafficAnomaly = true
		}
	}

	widget.CertStatus = certIssues

	// Build diagnostic cards from detected issues.
	var diagnostics []domain.DiagnosticCard

	// Offline node diagnostics
	if len(offlineNodes) > 0 {
		severity := "warning"
		if len(offlineNodes) > len(nodes)/2 {
			severity = "critical"
		}
		diagnostics = append(diagnostics, domain.DiagnosticCard{
			Severity:    severity,
			Title:       "Offline Nodes Detected",
			Description: formatOfflineNodes(offlineNodes),
			Actions:     []string{"restart", "investigate"},
		})
	}

	// Traffic anomaly diagnostic
	if widget.TrafficAnomaly {
		diagnostics = append(diagnostics, domain.DiagnosticCard{
			Severity:    "warning",
			Title:       "Traffic Anomaly Detected",
			Description: "One or more nodes have stale heartbeats while marked as connected.",
			Actions:     []string{"check_connectivity", "restart_core"},
		})
	}

	widget.Diagnostics = diagnostics
	return widget, nil
}

// ISPHeatmap returns a 7-day x 24-hour quality heatmap for the specified ISP.
func (s *DashboardProService) ISPHeatmap(ctx context.Context, isp string, days int) (*domain.ISPHeatmap, error) {
	if days <= 0 {
		days = 7
	}

	cells, err := s.ispQuality.GetHeatmap(ctx, isp, days)
	if err != nil {
		return nil, err
	}

	return &domain.ISPHeatmap{
		ISP:   isp,
		Cells: cells,
	}, nil
}

// GeoMap returns node locations with live status for geographic visualization.
func (s *DashboardProService) GeoMap(ctx context.Context) ([]*domain.GeoNode, error) {
	nodes, err := s.nodes.List(ctx)
	if err != nil {
		return nil, err
	}

	var geoNodes []*domain.GeoNode
	for _, n := range nodes {
		status := "offline"
		nodeStatus := n.Status
		if s.hub != nil {
			if liveStatus, _, _, ok := s.hub.Live(n.ID); ok {
				nodeStatus = liveStatus
			}
		}
		if nodeStatus == domain.NodeConnected {
			status = "online"
		}

		lat, lng := resolveNodeLocation(n)
		geoNodes = append(geoNodes, &domain.GeoNode{
			NodeID: n.ID,
			Name:   n.Name,
			Lat:    lat,
			Lng:    lng,
			Status: status,
		})
	}

	return geoNodes, nil
}

// resolveNodeLocation determines geographic coordinates for a node.
// It checks the node's CountryCode, Region, Address, and Name fields
// for known country/city patterns.
func resolveNodeLocation(n *domain.Node) (float64, float64) {
	// Map known country codes / names to coordinates
	locations := map[string][2]float64{
		"ir":            {35.6892, 51.3890},  // Tehran, Iran
		"iran":          {35.6892, 51.3890},
		"local":         {35.6892, 51.3890},  // Local = Iran
		"fr":            {48.8566, 2.3522},   // Paris, France
		"france":        {48.8566, 2.3522},
		"us":            {37.7749, -122.4194}, // San Francisco, US
		"usa":           {37.7749, -122.4194},
		"united states": {37.7749, -122.4194},
		"de":            {52.5200, 13.4050},  // Berlin, Germany
		"germany":       {52.5200, 13.4050},
		"nl":            {52.3676, 4.9041},   // Amsterdam, Netherlands
		"uk":            {51.5074, -0.1278},  // London, UK
		"gb":            {51.5074, -0.1278},
		"fi":            {60.1699, 24.9384},  // Helsinki, Finland
		"se":            {59.3293, 18.0686},  // Stockholm, Sweden
		"tr":            {41.0082, 28.9784},  // Istanbul, Turkey
		"ae":            {25.2048, 55.2708},  // Dubai, UAE
		"sg":            {1.3521, 103.8198},  // Singapore
		"jp":            {35.6762, 139.6503}, // Tokyo, Japan
		"ca":            {43.6532, -79.3832}, // Toronto, Canada
		"au":            {-33.8688, 151.2093}, // Sydney, Australia
	}

	// Try country code first (most reliable)
	if n.CountryCode != "" {
		code := strings.ToLower(n.CountryCode)
		if loc, ok := locations[code]; ok {
			return loc[0], loc[1]
		}
	}

	// Try node name (lowercase)
	name := strings.ToLower(n.Name)
	if loc, ok := locations[name]; ok {
		return loc[0], loc[1]
	}

	// Try region field
	if n.Region != "" {
		region := strings.ToLower(n.Region)
		for key, loc := range locations {
			if strings.Contains(region, key) {
				return loc[0], loc[1]
			}
		}
	}

	// Try node address
	addr := strings.ToLower(n.Address)
	for key, loc := range locations {
		if strings.Contains(addr, key) {
			return loc[0], loc[1]
		}
	}

	// Default: Tehran, Iran
	return 35.0, 51.0
}

// Revenue returns an aggregated revenue report for the given time range.
// If adminID is non-nil, the report is scoped to that admin/reseller.
func (s *DashboardProService) Revenue(ctx context.Context, adminID *uuid.UUID, from, to time.Time) (*domain.RevenueReport, error) {
	return s.revenue.Report(ctx, adminID, from, to)
}

// SubAnalytics returns subscription fetch analytics grouped by format, ISP,
// and hour-of-day for the given time range.
func (s *DashboardProService) SubAnalytics(ctx context.Context, from, to time.Time) (*domain.SubAnalyticsReport, error) {
	return s.subAnalytics.Report(ctx, from, to)
}

// formatOfflineNodes returns a human-readable summary of offline nodes.
func formatOfflineNodes(nodes []*domain.Node) string {
	if len(nodes) == 1 {
		return "Node '" + nodes[0].Name + "' is offline."
	}
	names := ""
	for i, n := range nodes {
		if i > 0 {
			names += ", "
		}
		if i >= 3 {
			names += "..."
			break
		}
		names += n.Name
	}
	return names + " are offline."
}

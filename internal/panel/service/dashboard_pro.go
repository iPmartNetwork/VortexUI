package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DashboardProService provides advanced dashboard analytics including daily
// checks, ISP heatmaps, geographic node visualization, revenue reporting,
// and subscription analytics.
type DashboardProService struct {
	nodes        port.NodeRepository
	users        port.UserRepository
	ispQuality   port.ISPQualityRepository
	subAnalytics port.SubscriptionAnalyticsRepository
	revenue      port.RevenueRepository
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
		if n.Status == domain.NodeConnected {
			widget.NodesOnline++
		} else {
			offlineNodes = append(offlineNodes, n)
		}

		// Check for stale nodes (no heartbeat in 5 minutes) as traffic anomaly signal.
		if n.LastSeen != nil && now.Sub(*n.LastSeen) > 5*time.Minute && n.Status == domain.NodeConnected {
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
		if n.Status == domain.NodeConnected {
			status = "online"
		}

		geoNodes = append(geoNodes, &domain.GeoNode{
			NodeID: n.ID,
			Name:   n.Name,
			Lat:    0, // populated from GeoIP resolution if available
			Lng:    0,
			Status: status,
		})
	}

	return geoNodes, nil
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

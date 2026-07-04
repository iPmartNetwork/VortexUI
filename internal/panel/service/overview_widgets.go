package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// WidgetDeps supplies optional data for dashboard widgets. All fields are
// optional; missing deps degrade gracefully to zero/empty widgets.
type WidgetDeps struct {
	Counters     port.DashboardCounter
	Traffic      port.TrafficRepository
	Inbounds     port.InboundRepository
	Routing      port.RoutingRepository
	Balancers    port.BalancerRepository
	RoutingPacks port.RoutingPackRepository
	Probing      port.ProbingRepository
}

func (s *OverviewService) buildWidgets(ctx context.Context, summary NodeSummary, stats domain.UserStats, adminID *uuid.UUID, sudo bool) domain.DashboardWidgets {
	w := domain.DashboardWidgets{
		NavBadges: domain.NavBadges{ActiveUsers: stats.ByStatus[domain.UserStatusActive]},
		Protocols: []domain.ProtocolStat{},
	}
	if s.widgets == nil {
		w.Telemetry = pickTelemetry(summary)
		return w
	}
	dep := s.widgets
	now := s.now()

	if dep.Counters != nil {
		if n, err := dep.Counters.CountOpenTickets(ctx); err == nil {
			w.NavBadges.OpenTickets = n
		}
		if n, err := dep.Counters.CountPendingOrders(ctx, adminID, sudo); err == nil {
			w.NavBadges.PendingOrders = n
		}
		since := now.Add(-24 * time.Hour)
		prevFrom := now.Add(-48 * time.Hour)
		curUsers, _ := dep.Counters.CountUsersCreatedSince(ctx, since)
		prevUsers, _ := dep.Counters.CountUsersCreatedBetween(ctx, prevFrom, since)
		w.Trends.UsersPct = pctChange(int64(curUsers), int64(prevUsers))

		if blocked, err := dep.Counters.CountBlockedIPs(ctx); err == nil {
			w.Probing.BlockedScanners = blocked
		}
		if events, err := dep.Counters.CountProbeEventsSince(ctx, since); err == nil {
			w.Probing.Events24h = events
		}
	}

	if dep.Probing != nil {
		if p, err := dep.Probing.GetPolicy(ctx); err == nil && p != nil {
			w.Probing.Enabled = p.Enabled
		}
	}

	if dep.Traffic != nil {
		curFrom := now.Add(-24 * time.Hour).Unix()
		prevFrom := now.Add(-48 * time.Hour).Unix()
		curTo := now.Unix()
		curPts, _ := dep.Traffic.TotalSeries(ctx, port.SeriesQuery{FromUnix: curFrom, ToUnix: curTo, Bucket: "1h"})
		prevPts, _ := dep.Traffic.TotalSeries(ctx, port.SeriesQuery{FromUnix: prevFrom, ToUnix: curFrom, Bucket: "1h"})
		w.Trends.BandwidthPct = pctChange(sumTraffic(curPts), sumTraffic(prevPts))
	}

	w.Protocols, w.Routing.Inbounds = aggregateProtocols(ctx, dep.Inbounds, s.nodes)
	w.Routing.ActiveRules, w.Routing.Balancers = countRouting(ctx, dep.Routing, dep.Balancers, s.nodes)
	if dep.RoutingPacks != nil {
		if packs, err := dep.RoutingPacks.List(ctx); err == nil {
			w.Routing.RoutingPacks = len(packs)
		}
	}

	active := stats.ByStatus[domain.UserStatusActive]
	if stats.Total > 0 {
		w.Trends.SessionsPct = math.Round(float64(active)/float64(stats.Total)*1000) / 10
	}

	w.Telemetry = pickTelemetry(summary)
	w.NodeFleet = buildNodeFleet(summary, dep, ctx)
	if dep.Counters != nil {
		if top, err := dep.Counters.TopUsersOverview(ctx, 5, adminID, sudo); err == nil {
			w.TopUsers = top
		}
	}
	return w
}

func buildNodeFleet(summary NodeSummary, dep *WidgetDeps, ctx context.Context) []domain.NodeFleetRow {
	var usersByNode map[uuid.UUID]int
	if dep != nil && dep.Counters != nil {
		usersByNode, _ = dep.Counters.UsersCountByNode(ctx)
	}
	out := make([]domain.NodeFleetRow, 0, len(summary.Items))
	for _, n := range summary.Items {
		load := math.Max(n.Health.CPUPercent, n.Health.MemPercent)
		st := "inactive"
		if n.Online && n.Health.CoreRunning {
			if load > 75 {
				st = "warning"
			} else {
				st = "active"
			}
		}
		id, _ := uuid.Parse(n.ID)
		out = append(out, domain.NodeFleetRow{
			ID:          n.ID,
			Name:        n.Name,
			Core:        string(n.Core),
			Location:    n.Location,
			CountryCode: n.CountryCode,
			PingMs:      n.PingMs,
			UsersCount:  usersByNode[id],
			Connections: n.Health.Connections,
			CPUPercent:  n.Health.CPUPercent,
			MemPercent:  n.Health.MemPercent,
			Online:      n.Online,
			Status:      st,
		})
	}
	return out
}

func pickTelemetry(summary NodeSummary) *domain.TelemetryWidget {
	for _, n := range summary.Items {
		if n.Online {
			return &domain.TelemetryWidget{
				NodeName:    n.Name,
				Core:        string(n.Core),
				Connections: n.Health.Connections,
				CPUPercent:  n.Health.CPUPercent,
				Online:      true,
				Location:    n.Location,
				PingMs:      n.PingMs,
			}
		}
	}
	if len(summary.Items) > 0 {
		n := summary.Items[0]
		return &domain.TelemetryWidget{
			NodeName:    n.Name,
			Core:        string(n.Core),
			Connections: n.Health.Connections,
			CPUPercent:  n.Health.CPUPercent,
			Online:      n.Online,
			Location:    n.Location,
			PingMs:      n.PingMs,
		}
	}
	return nil
}

func pctChange(current, previous int64) float64 {
	if previous <= 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return math.Round(float64(current-previous)/float64(previous)*1000) / 10
}

func sumTraffic(pts []domain.TrafficPoint) int64 {
	var total int64
	for _, p := range pts {
		total += p.Up + p.Down
	}
	return total
}

func aggregateProtocols(ctx context.Context, inbounds port.InboundRepository, nodes port.NodeRepository) ([]domain.ProtocolStat, int) {
	if inbounds == nil || nodes == nil {
		return nil, 0
	}
	list, err := nodes.List(ctx)
	if err != nil {
		return nil, 0
	}
	counts := map[string]int{}
	total := 0
	for _, node := range list {
		ibs, err := inbounds.ListByNode(ctx, node.ID)
		if err != nil {
			continue
		}
		for _, ib := range ibs {
			label := protocolLabel(string(ib.Protocol), ib.Network, string(ib.Security))
			counts[label]++
			total++
		}
	}
	if total == 0 {
		return []domain.ProtocolStat{}, 0
	}
	out := make([]domain.ProtocolStat, 0, len(counts))
	for label, count := range counts {
		out = append(out, domain.ProtocolStat{
			Label:   label,
			Count:   count,
			Percent: int(math.Round(float64(count) / float64(total) * 100)),
		})
	}
	// Sort descending by count (simple bubble for tiny sets).
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Count > out[i].Count {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, total
}

func protocolLabel(protocol, transport, security string) string {
	p := strings.ToLower(protocol)
	t := strings.ToLower(transport)
	s := strings.ToLower(security)
	switch {
	case p == "vless" && strings.Contains(s, "reality"):
		return "VLESS+Reality"
	case p == "vmess" && strings.Contains(t, "ws"):
		return "VMess+WS+CDN"
	case p == "hysteria2" || p == "hysteria":
		return "Hysteria2 UDP"
	case p == "trojan" || p == "shadowsocks":
		return "Trojan / SS"
	default:
		if t != "" {
			return strings.ToUpper(p) + "+" + strings.ToUpper(t)
		}
		return strings.ToUpper(p)
	}
}

func countRouting(ctx context.Context, routing port.RoutingRepository, balancers port.BalancerRepository, nodes port.NodeRepository) (rules, bal int) {
	if nodes == nil {
		return 0, 0
	}
	list, err := nodes.List(ctx)
	if err != nil {
		return 0, 0
	}
	for _, node := range list {
		if routing != nil {
			if rs, err := routing.ListByNode(ctx, node.ID); err == nil {
				rules += len(rs)
			}
		}
		if balancers != nil {
			if bs, err := balancers.ListByNode(ctx, node.ID); err == nil {
				bal += len(bs)
			}
		}
	}
	return rules, bal
}

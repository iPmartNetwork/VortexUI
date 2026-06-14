package singbox

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	scmd "github.com/xtls/xray-core/app/stats/command"
)

// statsClient reads per-user traffic counters. Abstracted so the driver's
// streaming/delta logic is testable with a fake.
type statsClient interface {
	QueryTraffic(ctx context.Context, reset bool) ([]UserTraffic, error)
	Close() error
}

// UserTraffic is one user's up/down sample; Email is the stats key (user UUID).
type UserTraffic struct {
	Email string
	Up    int64
	Down  int64
}

// grpcStats talks to sing-box's V2Ray API. That API implements the same stats
// gRPC service as Xray, so we reuse Xray's generated client and stat-name
// convention ("user>>>EMAIL>>>traffic>>>uplink|downlink").
type grpcStats struct {
	conn  *grpc.ClientConn
	stats scmd.StatsServiceClient
}

func dialStats(addr string) (statsClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcStats{conn: conn, stats: scmd.NewStatsServiceClient(conn)}, nil
}

func (g *grpcStats) Close() error { return g.conn.Close() }

func (g *grpcStats) QueryTraffic(ctx context.Context, reset bool) ([]UserTraffic, error) {
	resp, err := g.stats.QueryStats(ctx, &scmd.QueryStatsRequest{Pattern: "user>>>", Reset_: reset})
	if err != nil {
		return nil, err
	}
	byEmail := map[string]*UserTraffic{}
	for _, st := range resp.GetStat() {
		email, dir, ok := parseUserStat(st.GetName())
		if !ok {
			continue
		}
		ut := byEmail[email]
		if ut == nil {
			ut = &UserTraffic{Email: email}
			byEmail[email] = ut
		}
		switch dir {
		case "uplink":
			ut.Up += st.GetValue()
		case "downlink":
			ut.Down += st.GetValue()
		}
	}
	out := make([]UserTraffic, 0, len(byEmail))
	for _, ut := range byEmail {
		out = append(out, *ut)
	}
	return out, nil
}

func parseUserStat(name string) (email, dir string, ok bool) {
	parts := strings.Split(name, ">>>")
	if len(parts) != 4 || parts[0] != "user" || parts[2] != "traffic" {
		return "", "", false
	}
	return parts[1], parts[3], true
}

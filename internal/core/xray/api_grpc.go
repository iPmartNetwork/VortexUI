package xray

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	hcmd "github.com/xtls/xray-core/app/proxyman/command"
	scmd "github.com/xtls/xray-core/app/stats/command"
	xproto "github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	ssproxy "github.com/xtls/xray-core/proxy/shadowsocks"
	trojanproxy "github.com/xtls/xray-core/proxy/trojan"
	vlessproxy "github.com/xtls/xray-core/proxy/vless"
	vmessproxy "github.com/xtls/xray-core/proxy/vmess"

	"github.com/vortexui/vortexui/internal/domain"
)

// grpcAPI is the live implementation of xrayAPI. It speaks Xray's own gRPC API
// (exposed by the reserved loopback inbound) to mutate users and read counters
// without restarting the core.
type grpcAPI struct {
	conn    *grpc.ClientConn
	handler hcmd.HandlerServiceClient
	stats   scmd.StatsServiceClient
}

// dialGRPC connects to the Xray API on a loopback address. The link is plaintext
// because it never leaves localhost; the node's own firewall protects the port.
func dialGRPC(addr string) (xrayAPI, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial xray api %s: %w", addr, err)
	}
	return &grpcAPI{
		conn:    conn,
		handler: hcmd.NewHandlerServiceClient(conn),
		stats:   scmd.NewStatsServiceClient(conn),
	}, nil
}

func (g *grpcAPI) Close() error { return g.conn.Close() }

// AddUser builds the protocol-specific account and adds the user to the inbound.
func (g *grpcAPI) AddUser(ctx context.Context, in domain.Inbound, u *domain.User) error {
	account, err := buildAccount(in, u)
	if err != nil {
		return err
	}
	op := &hcmd.AddUserOperation{User: &xproto.User{
		Level:   0,
		Email:   u.ID.String(), // stats key == user UUID, matching the config builder
		Account: serial.ToTypedMessage(account),
	}}
	_, err = g.handler.AlterInbound(ctx, &hcmd.AlterInboundRequest{
		Tag:       in.Tag,
		Operation: serial.ToTypedMessage(op),
	})
	return err
}

// RemoveUser removes a user (by stats email) from an inbound.
func (g *grpcAPI) RemoveUser(ctx context.Context, inboundTag, email string) error {
	op := &hcmd.RemoveUserOperation{Email: email}
	_, err := g.handler.AlterInbound(ctx, &hcmd.AlterInboundRequest{
		Tag:       inboundTag,
		Operation: serial.ToTypedMessage(op),
	})
	return err
}

// QueryTraffic reads and (optionally) resets every per-user counter. Xray names
// counters "user>>>EMAIL>>>traffic>>>uplink|downlink"; we fold them per email.
func (g *grpcAPI) QueryTraffic(ctx context.Context, reset bool) ([]UserTraffic, error) {
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

// parseUserStat splits "user>>>EMAIL>>>traffic>>>uplink" into (email, direction).
func parseUserStat(name string) (email, dir string, ok bool) {
	parts := strings.Split(name, ">>>")
	if len(parts) != 4 || parts[0] != "user" || parts[2] != "traffic" {
		return "", "", false
	}
	return parts[1], parts[3], true
}

// OnlineUsers reports how many live connections each user currently has, via
// Xray's GetAllOnlineUsers. The response lists an entry per online user; we fold
// occurrences so the value is the connection count (>=1 means online).
func (g *grpcAPI) OnlineUsers(ctx context.Context) (map[string]int, error) {
	resp, err := g.stats.GetAllOnlineUsers(ctx, &scmd.GetAllOnlineUsersRequest{})
	if err != nil {
		return nil, err
	}
	out := make(map[string]int)
	for _, email := range resp.GetUsers() {
		out[email]++
	}
	return out, nil
}

// OnlineIPs lists the distinct source IPs currently online for one user via
// Xray's GetStatsOnlineIpList, mapping each IP to its last-seen unix time. A
// missing counter (user idle/unknown) is reported as an empty map, not an error.
func (g *grpcAPI) OnlineIPs(ctx context.Context, email string) (map[string]int64, error) {
	resp, err := g.stats.GetStatsOnlineIpList(ctx, &scmd.GetStatsRequest{Name: email})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return map[string]int64{}, nil
		}
		return nil, err
	}
	return resp.GetIps(), nil
}

// buildAccount maps a domain user + inbound into the matching Xray account proto.
func buildAccount(in domain.Inbound, u *domain.User) (proto.Message, error) {
	switch in.Protocol {
	case domain.ProtoVLESS:
		return &vlessproxy.Account{
			Id:         u.Proxies.VLESSUUID.String(),
			Flow:       in.Flow,
			Encryption: "none",
		}, nil
	case domain.ProtoVMess:
		return &vmessproxy.Account{
			Id:               u.Proxies.VMessUUID.String(),
			SecuritySettings: &xproto.SecurityConfig{Type: xproto.SecurityType_AUTO},
		}, nil
	case domain.ProtoTrojan:
		return &trojanproxy.Account{Password: u.Proxies.TrojanPass}, nil
	case domain.ProtoShadowsocks:
		return &ssproxy.Account{
			Password:   u.Proxies.ShadowsocksP,
			CipherType: cipherType(u.Proxies.SSMethod),
		}, nil
	default:
		return nil, fmt.Errorf("xray api: unsupported protocol %q", in.Protocol)
	}
}

func cipherType(method string) ssproxy.CipherType {
	switch method {
	case "aes-256-gcm":
		return ssproxy.CipherType_AES_256_GCM
	case "chacha20-poly1305", "chacha20-ietf-poly1305":
		return ssproxy.CipherType_CHACHA20_POLY1305
	case "none", "plain":
		return ssproxy.CipherType_NONE
	default:
		return ssproxy.CipherType_AES_128_GCM
	}
}

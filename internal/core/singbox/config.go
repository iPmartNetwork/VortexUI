// Package singbox implements core.CoreDriver for sing-box. Unlike Xray, sing-box
// has no live "alter inbound" gRPC for adding users, so the driver keeps the
// user set in memory and rebuilds + reloads the config on every change. This
// file renders the engine-neutral core.GeneratedConfig into sing-box JSON.
package singbox

import (
	"encoding/json"
	"fmt"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// Builder renders core.GeneratedConfig into sing-box native JSON. APIPort is the
// loopback port for sing-box's V2Ray API, which exposes per-user traffic stats.
type Builder struct {
	APIPort int
}

// Build implements core.Builder.
func (b Builder) Build(cfg *core.GeneratedConfig) ([]byte, error) {
	level := cfg.LogLevel
	if level == "" {
		level = "warn"
	}

	var inbounds []map[string]any
	emailSet := map[string]struct{}{}
	for _, in := range cfg.Inbounds {
		users := cfg.UsersByInbound[in.Tag]
		built, err := buildInbound(in, users)
		if err != nil {
			return nil, fmt.Errorf("inbound %q: %w", in.Tag, err)
		}
		inbounds = append(inbounds, built)
		for _, u := range users {
			emailSet[u.ID.String()] = struct{}{}
		}
	}

	emails := make([]string, 0, len(emailSet))
	for e := range emailSet {
		emails = append(emails, e)
	}

	conf := map[string]any{
		"log":       map[string]any{"level": level},
		"inbounds":  inbounds,
		"outbounds": []map[string]any{{"type": "direct", "tag": "direct"}},
		// V2Ray API exposes per-user up/down counters; stats.users must list the
		// users to track, which is why adding a user requires a config rebuild.
		"experimental": map[string]any{
			"v2ray_api": map[string]any{
				"listen": fmt.Sprintf("127.0.0.1:%d", b.APIPort),
				"stats":  map[string]any{"enabled": true, "users": emails},
			},
		},
	}
	return json.MarshalIndent(conf, "", "  ")
}

func buildInbound(in domain.Inbound, users []*domain.User) (map[string]any, error) {
	m := map[string]any{
		"tag":         in.Tag,
		"listen":      orDefault(in.Listen, "::"),
		"listen_port": in.Port,
	}

	switch in.Protocol {
	case domain.ProtoVLESS:
		m["type"] = "vless"
		m["users"] = vlessUsers(in, users)
	case domain.ProtoVMess:
		m["type"] = "vmess"
		m["users"] = vmessUsers(users)
	case domain.ProtoTrojan:
		m["type"] = "trojan"
		m["users"] = trojanUsers(users)
	case domain.ProtoShadowsocks:
		m["type"] = "shadowsocks"
		m["method"] = ssMethod(users)
		m["users"] = ssUsers(users)
	default:
		return nil, fmt.Errorf("unsupported protocol %q for sing-box", in.Protocol)
	}

	if tls := tlsBlock(in); tls != nil {
		m["tls"] = tls
	}
	if tr := transportBlock(in); tr != nil {
		m["transport"] = tr
	}
	if in.Raw != nil {
		if extra, ok := in.Raw["singbox"].(map[string]any); ok {
			for k, v := range extra {
				m[k] = v // operator override/escape hatch
			}
		}
	}
	return m, nil
}

func vlessUsers(in domain.Inbound, users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "uuid": u.Proxies.VLESSUUID.String(), "flow": in.Flow})
	}
	return out
}

func vmessUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "uuid": u.Proxies.VMessUUID.String(), "alterId": 0})
	}
	return out
}

func trojanUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.TrojanPass})
	}
	return out
}

func ssUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.ShadowsocksP})
	}
	return out
}

// ssMethod takes the cipher from the first user; all users on an SS inbound share
// one method in sing-box.
func ssMethod(users []*domain.User) string {
	if len(users) > 0 && users[0].Proxies.SSMethod != "" {
		return users[0].Proxies.SSMethod
	}
	return "aes-128-gcm"
}

func tlsBlock(in domain.Inbound) map[string]any {
	if in.Security != domain.SecurityTLS && in.Security != domain.SecurityReality {
		return nil
	}
	tls := map[string]any{"enabled": true}
	if len(in.SNI) > 0 {
		tls["server_name"] = in.SNI[0]
	}
	if in.Security == domain.SecurityReality {
		tls["reality"] = map[string]any{"enabled": true}
	}
	return tls
}

func transportBlock(in domain.Inbound) map[string]any {
	switch in.Network {
	case "ws":
		t := map[string]any{"type": "ws"}
		if in.Path != "" {
			t["path"] = in.Path
		}
		if len(in.Host) > 0 {
			t["headers"] = map[string]any{"Host": in.Host[0]}
		}
		return t
	case "grpc":
		return map[string]any{"type": "grpc", "service_name": in.Path}
	default:
		return nil
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

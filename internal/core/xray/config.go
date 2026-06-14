// Package xray implements core.CoreDriver for Xray-core. This file renders the
// engine-neutral core.GeneratedConfig into Xray's native JSON using typed
// structs (no map[string]any soup) so the output is predictable and unit-tested.
package xray

import (
	"encoding/json"
	"fmt"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// APIInboundTag is the reserved inbound Xray exposes its gRPC API on. The driver
// dials this for runtime user mutations and stats queries.
const APIInboundTag = "vortex-api"

// Builder renders core.GeneratedConfig to Xray JSON. apiPort is where the local
// gRPC API inbound listens (loopback only).
type Builder struct {
	APIPort int
}

// Build implements core.Builder.
func (b Builder) Build(cfg *core.GeneratedConfig) ([]byte, error) {
	x := xrayConfig{
		Log:   logConf{Loglevel: orDefault(cfg.LogLevel, "warning")},
		API:   apiConf{Tag: APIInboundTag, Services: []string{"HandlerService", "StatsService", "LoggerService"}},
		Stats: struct{}{},
		Policy: policyConf{
			Levels: map[string]policyLevel{
				"0": {StatsUserUplink: true, StatsUserDownlink: true},
			},
			System: systemPolicy{StatsInboundUplink: true, StatsInboundDownlink: true},
		},
		Outbounds: []outbound{{Protocol: "freedom", Tag: "direct"}, {Protocol: "blackhole", Tag: "blocked"}},
		Routing:   routingConf{Rules: []routingRule{{Type: "field", InboundTag: []string{APIInboundTag}, OutboundTag: "api"}}},
	}

	// Reserved loopback inbound carrying the Xray API.
	x.Inbounds = append(x.Inbounds, inbound{
		Tag: APIInboundTag, Listen: "127.0.0.1", Port: b.APIPort,
		Protocol: "dokodemo-door",
		Settings: mustRaw(map[string]any{"address": "127.0.0.1"}),
	})

	for _, in := range cfg.Inbounds {
		users := cfg.UsersByInbound[in.Tag]
		built, err := b.buildInbound(in, users)
		if err != nil {
			return nil, fmt.Errorf("inbound %q: %w", in.Tag, err)
		}
		x.Inbounds = append(x.Inbounds, built)
	}

	return json.MarshalIndent(x, "", "  ")
}

func (b Builder) buildInbound(in domain.Inbound, users []*domain.User) (inbound, error) {
	settings, err := protocolSettings(in, users)
	if err != nil {
		return inbound{}, err
	}
	out := inbound{
		Tag:            in.Tag,
		Listen:         orDefault(in.Listen, "0.0.0.0"),
		Port:           in.Port,
		Protocol:       string(in.Protocol),
		Settings:       settings,
		StreamSettings: streamSettings(in),
	}
	return out, nil
}

// protocolSettings renders the protocol-specific "settings" block, including the
// per-user client list keyed off the shared credentials.
func protocolSettings(in domain.Inbound, users []*domain.User) (json.RawMessage, error) {
	switch in.Protocol {
	case domain.ProtoVLESS:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{
				"id": u.Proxies.VLESSUUID.String(), "email": u.ID.String(), "flow": in.Flow,
			})
		}
		return mustRaw(map[string]any{"clients": clients, "decryption": "none"}), nil

	case domain.ProtoVMess:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{"id": u.Proxies.VMessUUID.String(), "email": u.ID.String()})
		}
		return mustRaw(map[string]any{"clients": clients}), nil

	case domain.ProtoTrojan:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{"password": u.Proxies.TrojanPass, "email": u.ID.String()})
		}
		return mustRaw(map[string]any{"clients": clients}), nil

	case domain.ProtoShadowsocks:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{
				"password": u.Proxies.ShadowsocksP, "method": u.Proxies.SSMethod, "email": u.ID.String(),
			})
		}
		return mustRaw(map[string]any{"clients": clients}), nil

	default:
		return nil, fmt.Errorf("unsupported protocol %q for xray", in.Protocol)
	}
}

// streamSettings renders transport + security. Operators can override the whole
// block via Inbound.Raw["streamSettings"] for fields the abstraction omits.
func streamSettings(in domain.Inbound) json.RawMessage {
	if raw, ok := in.Raw["streamSettings"]; ok {
		return mustRaw(raw)
	}
	ss := map[string]any{
		"network":  orDefault(in.Network, "tcp"),
		"security": string(in.Security),
	}
	switch in.Security {
	case domain.SecurityTLS:
		tls := map[string]any{}
		if len(in.SNI) > 0 {
			tls["serverName"] = in.SNI[0]
		}
		ss["tlsSettings"] = tls
	case domain.SecurityReality:
		// Reality keys/shortIds are expected in Raw until the evasion-profile
		// layer is built; we still emit the block shape here.
		reality := map[string]any{}
		if len(in.SNI) > 0 {
			reality["serverNames"] = in.SNI
		}
		ss["realitySettings"] = reality
	}
	switch in.Network {
	case "ws":
		ws := map[string]any{}
		if in.Path != "" {
			ws["path"] = in.Path
		}
		if len(in.Host) > 0 {
			ws["headers"] = map[string]any{"Host": in.Host[0]}
		}
		ss["wsSettings"] = ws
	case "grpc":
		ss["grpcSettings"] = map[string]any{"serviceName": in.Path}
	}
	return mustRaw(ss)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func mustRaw(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil { // only unmarshalable types (chan/func) hit this; never with our inputs
		panic(fmt.Sprintf("xray: marshal settings: %v", err))
	}
	return b
}

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ShareLinkService generates protocol-specific share URIs for inbounds.
type ShareLinkService struct {
	inbounds port.InboundRepository
	nodes    port.NodeRepository
}

func NewShareLinkService(inbounds port.InboundRepository, nodes port.NodeRepository) *ShareLinkService {
	return &ShareLinkService{inbounds: inbounds, nodes: nodes}
}

// GenerateLink produces a share link for the given inbound.
// Returns empty string with error for unsupported protocols.
func (s *ShareLinkService) GenerateLink(ctx context.Context, inboundID uuid.UUID) (string, error) {
	in, err := s.inbounds.GetByID(ctx, inboundID)
	if err != nil {
		return "", err
	}
	node, err := s.nodes.GetByID(ctx, in.NodeID)
	if err != nil {
		return "", fmt.Errorf("node not found")
	}

	host := resolveShareHost(in, node)
	p := in.Port

	switch in.Protocol {
	case domain.ProtoVLESS:
		return buildVLESSLink(in, host, p), nil
	case domain.ProtoVMess:
		return buildVMessLink(in, host, p), nil
	case domain.ProtoTrojan:
		return buildTrojanLink(in, host, p), nil
	case domain.ProtoShadowsocks:
		return buildShadowsocksLink(in, host, p), nil
	case domain.ProtoHysteria2:
		return buildHysteria2Link(in, host, p), nil
	case domain.ProtoTUIC:
		return buildTUICLink(in, host, p), nil
	default:
		return "", fmt.Errorf("protocol %q does not support share links", in.Protocol)
	}
}

// resolveShareHost determines the host for the share link.
// Priority: first SNI > node endpoint > node address (host part).
func resolveShareHost(in *domain.Inbound, node *domain.Node) string {
	if len(in.SNI) > 0 && in.SNI[0] != "" {
		return in.SNI[0]
	}
	if node.Endpoint != "" {
		return node.Endpoint
	}
	return hostOf(node.Address)
}

func buildVLESSLink(in *domain.Inbound, host string, port int) string {
	// vless://uuid@host:port?params#tag
	userUUID := uuid.New().String()
	params := url.Values{}
	params.Set("encryption", "none")
	params.Set("type", orStr(in.Network, "tcp"))
	params.Set("security", string(in.Security))
	if in.Security == domain.SecurityTLS || in.Security == domain.SecurityReality {
		if len(in.SNI) > 0 {
			params.Set("sni", in.SNI[0])
		}
	}
	if in.Security == domain.SecurityReality {
		if rm, ok := in.Raw["reality"].(map[string]any); ok {
			if pk, ok := rm["public_key"].(string); ok {
				params.Set("pbk", pk)
			}
			if sids, ok := rm["short_ids"].([]any); ok && len(sids) > 0 {
				params.Set("sid", fmt.Sprint(sids[0]))
			}
		}
		params.Set("fp", "chrome")
	}
	if in.Path != "" {
		params.Set("path", in.Path)
	}
	if len(in.Host) > 0 && in.Host[0] != "" {
		params.Set("host", in.Host[0])
	}
	if in.Flow != "" {
		params.Set("flow", in.Flow)
	}
	if in.Network == "grpc" && in.Path != "" {
		params.Set("serviceName", in.Path)
	}
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s", userUUID, host, port, params.Encode(), url.PathEscape(in.Tag))
}

func buildVMessLink(in *domain.Inbound, host string, port int) string {
	// vmess://base64(json)
	obj := map[string]any{
		"v":    "2",
		"ps":   in.Tag,
		"add":  host,
		"port": port,
		"id":   uuid.New().String(),
		"aid":  0,
		"net":  orStr(in.Network, "tcp"),
		"type": "none",
		"host": "",
		"path": in.Path,
		"tls":  "",
		"sni":  "",
	}
	if in.Security == domain.SecurityTLS {
		obj["tls"] = "tls"
	}
	if len(in.SNI) > 0 {
		obj["sni"] = in.SNI[0]
	}
	if len(in.Host) > 0 {
		obj["host"] = in.Host[0]
	}
	raw, _ := json.Marshal(obj)
	return "vmess://" + base64.StdEncoding.EncodeToString(raw)
}

func buildTrojanLink(in *domain.Inbound, host string, port int) string {
	// trojan://password@host:port?params#tag
	password := uuid.New().String()
	params := url.Values{}
	params.Set("type", orStr(in.Network, "tcp"))
	params.Set("security", string(in.Security))
	if len(in.SNI) > 0 {
		params.Set("sni", in.SNI[0])
	}
	if in.Path != "" {
		params.Set("path", in.Path)
	}
	if len(in.Host) > 0 && in.Host[0] != "" {
		params.Set("host", in.Host[0])
	}
	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s", password, host, port, params.Encode(), url.PathEscape(in.Tag))
}

func buildShadowsocksLink(in *domain.Inbound, host string, port int) string {
	// ss://base64(method:password)@host:port#tag
	method := "chacha20-ietf-poly1305"
	password := uuid.New().String()
	userInfo := base64.URLEncoding.EncodeToString([]byte(method + ":" + password))
	return fmt.Sprintf("ss://%s@%s:%d#%s", userInfo, host, port, url.PathEscape(in.Tag))
}

func buildHysteria2Link(in *domain.Inbound, host string, port int) string {
	// hysteria2://auth@host:port?params#tag
	auth := uuid.New().String()
	params := url.Values{}
	if len(in.SNI) > 0 {
		params.Set("sni", in.SNI[0])
	}
	params.Set("insecure", "0")
	return fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s", auth, host, port, params.Encode(), url.PathEscape(in.Tag))
}

func buildTUICLink(in *domain.Inbound, host string, port int) string {
	// tuic://uuid:password@host:port?params#tag
	uid := uuid.New().String()
	password := "password"
	params := url.Values{}
	if len(in.SNI) > 0 {
		params.Set("sni", in.SNI[0])
	}
	params.Set("congestion_control", "bbr")
	params.Set("alpn", "h3")
	return fmt.Sprintf("tuic://%s:%s@%s:%d?%s#%s", uid, password, host, port, params.Encode(), url.PathEscape(in.Tag))
}

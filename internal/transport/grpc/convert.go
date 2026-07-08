package grpc

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vortexui/vortexui/internal/domain"
	genv1 "github.com/vortexui/vortexui/internal/transport/genv1"
)

// These converters are the only place proto wire types and domain types meet,
// keeping both layers ignorant of each other everywhere else.

func coreTypeToProto(c domain.CoreType) genv1.CoreType {
	switch c {
	case domain.CoreXray:
		return genv1.CoreType_CORE_TYPE_XRAY
	case domain.CoreSingbox:
		return genv1.CoreType_CORE_TYPE_SINGBOX
	default:
		return genv1.CoreType_CORE_TYPE_UNSPECIFIED
	}
}

func coreTypeFromProto(c genv1.CoreType) domain.CoreType {
	switch c {
	case genv1.CoreType_CORE_TYPE_XRAY:
		return domain.CoreXray
	case genv1.CoreType_CORE_TYPE_SINGBOX:
		return domain.CoreSingbox
	default:
		return ""
	}
}

// userToSpec renders a domain user into the minimal identity a core needs. The
// proto "email" is the per-inbound stats key; we use the user id for stability.
func userToSpec(u *domain.User) *genv1.UserSpec {
	return &genv1.UserSpec{
		UserId:         u.ID.String(),
		Email:          u.ID.String(),
		VmessUuid:      u.Proxies.VMessUUID.String(),
		VlessUuid:      u.Proxies.VLESSUUID.String(),
		TrojanPassword: u.Proxies.TrojanPass,
		SsPassword:     u.Proxies.ShadowsocksP,
		SsMethod:       u.Proxies.SSMethod,
	}
}

// userFromSpec is the inverse, used on the node side. Parse errors collapse to a
// zero UUID rather than failing the whole sync; the node logs and skips.
func userFromSpec(s *genv1.UserSpec) *domain.User {
	id, _ := uuid.Parse(s.GetUserId())
	vmess, _ := uuid.Parse(s.GetVmessUuid())
	vless, _ := uuid.Parse(s.GetVlessUuid())
	return &domain.User{
		ID: id,
		Proxies: domain.UserCredentials{
			VMessUUID:    vmess,
			VLESSUUID:    vless,
			TrojanPass:   s.GetTrojanPassword(),
			ShadowsocksP: s.GetSsPassword(),
			SSMethod:     s.GetSsMethod(),
		},
	}
}

func inboundToSpec(in domain.Inbound) *genv1.InboundSpec {
	spec := &genv1.InboundSpec{
		Tag:      in.Tag,
		Protocol: string(in.Protocol),
		Listen:   in.Listen,
		Port:     uint32(in.Port),
		Network:  in.Network,
		Security: string(in.Security),
		Sni:      in.SNI,
		Path:     in.Path,
		Host:     in.Host,
		Flow:     in.Flow,
	}
	if len(in.Raw) > 0 {
		if b, err := json.Marshal(in.Raw); err == nil {
			spec.Raw = b
		}
	}
	return spec
}

func inboundFromSpec(s *genv1.InboundSpec) domain.Inbound {
	in := domain.Inbound{
		Tag:      s.GetTag(),
		Protocol: domain.Protocol(s.GetProtocol()),
		Listen:   s.GetListen(),
		Port:     int(s.GetPort()),
		Network:  s.GetNetwork(),
		Security: domain.Security(s.GetSecurity()),
		SNI:      s.GetSni(),
		Path:     s.GetPath(),
		Host:     s.GetHost(),
		Flow:     s.GetFlow(),
		Enabled:  true, // only enabled inbounds are ever synced
	}
	if raw := s.GetRaw(); len(raw) > 0 {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err == nil {
			in.Raw = m
		}
	}
	return in
}

func outboundToSpec(o domain.Outbound) *genv1.OutboundSpec {
	spec := &genv1.OutboundSpec{
		Tag:      o.Tag,
		Protocol: string(o.Protocol),
		Address:  o.Address,
		Port:     uint32(o.Port),
		Uuid:     o.UUID,
		Password: o.Password,
		Username: o.Username,
		Method:   o.Method,
		Flow:     o.Flow,
		Network:  o.Network,
		Security: string(o.Security),
		Sni:      o.SNI,
		Path:     o.Path,
		Host:     o.Host,
		Enabled:  o.Enabled,
	}
	if len(o.Raw) > 0 {
		if b, err := json.Marshal(o.Raw); err == nil {
			spec.Raw = b
		}
	}
	return spec
}

func outboundFromSpec(s *genv1.OutboundSpec) domain.Outbound {
	o := domain.Outbound{
		Tag:      s.GetTag(),
		Protocol: domain.OutboundProtocol(s.GetProtocol()),
		Address:  s.GetAddress(),
		Port:     int(s.GetPort()),
		UUID:     s.GetUuid(),
		Password: s.GetPassword(),
		Username: s.GetUsername(),
		Method:   s.GetMethod(),
		Flow:     s.GetFlow(),
		Network:  s.GetNetwork(),
		Security: domain.Security(s.GetSecurity()),
		SNI:      s.GetSni(),
		Path:     s.GetPath(),
		Host:     s.GetHost(),
		Enabled:  s.GetEnabled(),
	}
	if raw := s.GetRaw(); len(raw) > 0 {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err == nil {
			o.Raw = m
		}
	}
	return o
}

func routingToSpec(r domain.RoutingRule) *genv1.RoutingRuleSpec {
	return &genv1.RoutingRuleSpec{
		Priority:    int32(r.Priority),
		Name:        r.Name,
		InboundTags: r.InboundTags,
		Domains:     r.Domains,
		Ip:          r.IP,
		Port:        r.Port,
		Protocols:   r.Protocols,
		Network:     r.Network,
		OutboundTag: r.OutboundTag,
		BalancerTag: r.BalancerTag,
		Enabled:     r.Enabled,
	}
}

func routingFromSpec(s *genv1.RoutingRuleSpec) domain.RoutingRule {
	return domain.RoutingRule{
		Priority:    int(s.GetPriority()),
		Name:        s.GetName(),
		InboundTags: s.GetInboundTags(),
		Domains:     s.GetDomains(),
		IP:          s.GetIp(),
		Port:        s.GetPort(),
		Protocols:   s.GetProtocols(),
		Network:     s.GetNetwork(),
		OutboundTag: s.GetOutboundTag(),
		BalancerTag: s.GetBalancerTag(),
		Enabled:     s.GetEnabled(),
	}
}

func balancerToSpec(b domain.Balancer) *genv1.BalancerSpec {
	return &genv1.BalancerSpec{
		Tag:           b.Tag,
		Selectors:     b.Selectors,
		Strategy:      string(b.Strategy),
		Observe:       b.Observe,
		ProbeUrl:      b.ProbeURL,
		ProbeInterval: b.ProbeInterval,
		Enabled:       b.Enabled,
	}
}

func balancerFromSpec(s *genv1.BalancerSpec) domain.Balancer {
	return domain.Balancer{
		Tag:           s.GetTag(),
		Selectors:     s.GetSelectors(),
		Strategy:      domain.BalancerStrategy(s.GetStrategy()),
		Observe:       s.GetObserve(),
		ProbeURL:      s.GetProbeUrl(),
		ProbeInterval: s.GetProbeInterval(),
		Enabled:       s.GetEnabled(),
	}
}

// trafficFromProto converts a wire delta into the domain delta, attributing it
// to the node that produced it (the panel side knows which node the stream is).
func trafficFromProto(d *genv1.TrafficDelta, nodeID uuid.UUID) domain.TrafficDelta {
	uid, _ := uuid.Parse(d.GetUserId())
	ts := time.Now()
	if d.GetTimestamp() != nil {
		ts = d.GetTimestamp().AsTime()
	}
	return domain.TrafficDelta{
		NodeID:    nodeID,
		UserID:    uid,
		Up:        int64(d.GetUp()),
		Down:      int64(d.GetDown()),
		Timestamp: ts,
	}
}

func trafficToProto(d domain.TrafficDelta) *genv1.TrafficDelta {
	return &genv1.TrafficDelta{
		UserId:    d.UserID.String(),
		Up:        uint64(d.Up),
		Down:      uint64(d.Down),
		Timestamp: timestamppb.New(d.Timestamp),
	}
}

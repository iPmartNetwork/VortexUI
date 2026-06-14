package grpc

import (
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
	case genv1.CoreType_CORE_TYPE_SINGBOX:
		return domain.CoreSingbox
	default:
		return domain.CoreXray
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
	return &genv1.InboundSpec{
		Tag:      in.Tag,
		Protocol: string(in.Protocol),
		Listen:   in.Listen,
		Port:     uint32(in.Port),
		Network:  in.Network,
		Security: string(in.Security),
	}
}

func inboundFromSpec(s *genv1.InboundSpec) domain.Inbound {
	return domain.Inbound{
		Tag:      s.GetTag(),
		Protocol: domain.Protocol(s.GetProtocol()),
		Listen:   s.GetListen(),
		Port:     int(s.GetPort()),
		Network:  s.GetNetwork(),
		Security: domain.Security(s.GetSecurity()),
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

package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// coreCapabilityDTO is the JSON shape of a single core's capability matrix.
// domain.Protocol and domain.Security are string types, so they marshal as
// plain JSON strings.
type coreCapabilityDTO struct {
	Protocols  []domain.Protocol `json:"protocols"`
	Transports []string          `json:"transports"`
	Securities []domain.Security `json:"securities"`
	UDPNative  []domain.Protocol `json:"udp_native"`
	// NoTransport lists the protocols whose constraint marks them as carrying no
	// stream transport (the UI hides the transport picker for these).
	NoTransport []domain.Protocol `json:"no_transport"`
	// ProtocolSecurities maps a protocol to its security override, present only
	// for protocols that constrain securities below the core-wide set (the UI
	// uses it to narrow the security picker per protocol).
	ProtocolSecurities map[domain.Protocol][]domain.Security `json:"protocol_securities"`
}

// capabilitiesResponse exposes the per-core capability matrix for both cores so
// the UI can filter its inbound-form options to what the selected node's core
// can actually render.
type capabilitiesResponse struct {
	Xray    coreCapabilityDTO `json:"xray"`
	Singbox coreCapabilityDTO `json:"singbox"`
}

// toCapabilityDTO normalizes nil slices to empty slices so the JSON always has
// arrays (never null) for predictable frontend consumption.
func toCapabilityDTO(c core.Capability) coreCapabilityDTO {
	dto := coreCapabilityDTO{
		Protocols:          c.Protocols,
		Transports:         c.Transports,
		Securities:         c.Securities,
		UDPNative:          c.UDPNative,
		NoTransport:        []domain.Protocol{},
		ProtocolSecurities: map[domain.Protocol][]domain.Security{},
	}
	if dto.Protocols == nil {
		dto.Protocols = []domain.Protocol{}
	}
	if dto.Transports == nil {
		dto.Transports = []string{}
	}
	if dto.Securities == nil {
		dto.Securities = []domain.Security{}
	}
	if dto.UDPNative == nil {
		dto.UDPNative = []domain.Protocol{}
	}
	// Derive the per-protocol constraint views from the matrix. Iterate the
	// Protocols slice (stable order) so NoTransport is deterministic.
	for _, proto := range dto.Protocols {
		con, ok := c.Constraints[proto]
		if !ok {
			continue
		}
		if con.NoTransport {
			dto.NoTransport = append(dto.NoTransport, proto)
		}
		if len(con.Securities) > 0 {
			dto.ProtocolSecurities[proto] = con.Securities
		}
	}
	return dto
}

// GetCapabilities returns the authoritative capability matrix for both cores
// (xray and sing-box): the protocols, transports, securities, and UDP-native
// protocols each core supports. The values come from the single source of truth
// in internal/core, so the API can never disagree with the guard or renderers.
func (h *Handlers) GetCapabilities(c echo.Context) error {
	return c.JSON(http.StatusOK, capabilitiesResponse{
		Xray:    toCapabilityDTO(core.Capabilities(domain.CoreXray)),
		Singbox: toCapabilityDTO(core.Capabilities(domain.CoreSingbox)),
	})
}

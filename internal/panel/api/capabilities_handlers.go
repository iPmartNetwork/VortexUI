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
		Protocols:  c.Protocols,
		Transports: c.Transports,
		Securities: c.Securities,
		UDPNative:  c.UDPNative,
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

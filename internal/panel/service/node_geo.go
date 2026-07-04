package service

import (
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/geoip"
)

// GeoCountryResolver resolves IPs to ISO country codes (MaxMind or disabled).
type GeoCountryResolver interface {
	Country(ip string) string
}

// ApplyNodeGeo fills country/region from the node's public endpoint when auto mode
// is enabled. Manual region always wins; country is refreshed when auto.
func ApplyNodeGeo(n *domain.Node, geo GeoCountryResolver) {
	if n == nil {
		return
	}
	host := nodePublicHost(n)
	if host == "" {
		return
	}
	if geoip.IsLocal(host) {
		if n.LocationAuto {
			if strings.TrimSpace(n.Region) == "" {
				n.Region = "Local"
			}
			n.CountryCode = ""
		}
		return
	}
	if geo != nil && n.LocationAuto {
		if cc := geo.Country(host); cc != "" {
			n.CountryCode = cc
			if strings.TrimSpace(n.Region) == "" {
				n.Region = geoip.FormatLocation("", cc)
			}
		}
	}
}

func nodePublicHost(n *domain.Node) string {
	if ep := strings.TrimSpace(n.Endpoint); ep != "" {
		return geoip.HostOnly(ep)
	}
	return geoip.HostOnly(n.Address)
}

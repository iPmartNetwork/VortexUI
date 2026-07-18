package service

import (
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// AutoConfigRecommendation describes the optimal proxy settings for a given context.
type AutoConfigRecommendation struct {
	Protocol  domain.Protocol `json:"protocol"`
	Transport string          `json:"transport"`
	Security  string          `json:"security"`
	Fragment  string          `json:"fragment,omitempty"`
	Mux       bool            `json:"mux"`
	ECH       bool            `json:"ech"`
	Reason    string          `json:"reason"`
}

// AutoConfigEngine provides rule-based "AI" configuration recommendations.
// It selects the best protocol+transport+anti-DPI settings based on ISP,
// time of day (DPI is stricter during peak hours in Iran: 18:00-24:00),
// and whether the node is behind a CDN.
type AutoConfigEngine struct {
	now func() time.Time
}

// NewAutoConfigEngine creates a new auto-config engine using real time.
func NewAutoConfigEngine() *AutoConfigEngine {
	return &AutoConfigEngine{now: time.Now}
}

// Recommend returns the best proxy configuration for the given ISP and context.
func (e *AutoConfigEngine) Recommend(isp domain.ISPPreset, hasCDN bool) AutoConfigRecommendation {
	now := e.now()
	hour := now.Hour()
	isPeakHours := hour >= 18 && hour <= 23 // Iranian DPI is strictest 18:00-23:00

	switch isp {
	case domain.ISPHamrahAval: // MCI — heavy DPI
		if isPeakHours {
			return AutoConfigRecommendation{
				Protocol: domain.ProtoVLESS, Transport: "ws", Security: "tls",
				Fragment: "10-50,10-20,tlshello", Mux: true, ECH: true,
				Reason: "MCI peak hours: max evasion with fragment+mux+ECH",
			}
		}
		if hasCDN {
			return AutoConfigRecommendation{
				Protocol: domain.ProtoVLESS, Transport: "ws", Security: "tls",
				Fragment: "10-30,10-15,tlshello", Mux: false,
				Reason: "MCI off-peak with CDN: WS+TLS+fragment sufficient",
			}
		}
		return AutoConfigRecommendation{
			Protocol: domain.ProtoVLESS, Transport: "tcp", Security: "reality",
			Reason: "MCI off-peak direct: REALITY best stealth",
		}

	case domain.ISPIrancell: // MTN — moderate DPI
		if isPeakHours {
			return AutoConfigRecommendation{
				Protocol: domain.ProtoVLESS, Transport: "grpc", Security: "tls",
				Fragment: "1-3,1-5,1-3", Mux: true,
				Reason: "Irancell peak: gRPC+small fragment+mux",
			}
		}
		return AutoConfigRecommendation{
			Protocol: domain.ProtoVLESS, Transport: "tcp", Security: "reality",
			Reason: "Irancell off-peak: REALITY sufficient",
		}

	case domain.ISPMokhaberat: // TCI — strongest DPI
		return AutoConfigRecommendation{
			Protocol: domain.ProtoVLESS, Transport: "ws", Security: "tls",
			Fragment: "50-100,20-50,tlshello", Mux: true, ECH: true,
			Reason: "Mokhaberat: always needs max evasion (fragment+mux+ECH)",
		}

	case domain.ISPShatel:
		return AutoConfigRecommendation{
			Protocol: domain.ProtoVLESS, Transport: "ws", Security: "tls",
			Fragment: "10-30,5-15,tlshello",
			Reason: "Shatel: moderate fragment sufficient",
		}

	case domain.ISPAsiatech:
		return AutoConfigRecommendation{
			Protocol: domain.ProtoVLESS, Transport: "grpc", Security: "tls",
			Fragment: "20-60,10-30,1-5",
			Reason: "Asiatech: gRPC+medium fragment",
		}

	default:
		if hasCDN {
			return AutoConfigRecommendation{
				Protocol: domain.ProtoVLESS, Transport: "ws", Security: "tls",
				Reason: "Unknown ISP with CDN: WS+TLS safe default",
			}
		}
		return AutoConfigRecommendation{
			Protocol: domain.ProtoVLESS, Transport: "tcp", Security: "reality",
			Reason: "Unknown ISP direct: REALITY default",
		}
	}
}

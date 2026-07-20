package service

import (
	"sort"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/subscription"
)

// ComputeQualityScores assigns a quality score (0-100) to each proxy based on
// ISP compatibility, protocol characteristics, and transport suitability.
// Higher scores indicate better expected performance for the detected ISP.
// After scoring, proxies are stable-sorted by score descending so the best
// proxy is first in the subscription.
func ComputeQualityScores(proxies []subscription.Proxy, isp domain.ISPPreset) {
	if len(proxies) == 0 {
		return
	}

	for i := range proxies {
		proxies[i].QualityScore = scoreProxy(proxies[i], isp)
	}

	// Stable sort: highest quality first.
	sort.SliceStable(proxies, func(a, b int) bool {
		return proxies[a].QualityScore > proxies[b].QualityScore
	})
}

// scoreProxy computes quality for a single proxy. The score combines:
// - Protocol base score (protocol inherent quality)
// - Transport bonus (how well the transport works on this ISP)
// - Security bonus (stealth level)
// - Feature bonuses (mux, fragment, CDN, etc.)
func scoreProxy(p subscription.Proxy, isp domain.ISPPreset) int {
	score := 50 // base

	// Protocol base scores (inherent quality/speed/reliability).
	switch p.Protocol {
	case domain.ProtoVLESS:
		score += 15 // lightweight, fast
	case domain.ProtoTrojan:
		score += 12 // good stealth, slight overhead
	case domain.ProtoVMess:
		score += 10 // more overhead than VLESS
	case domain.ProtoHysteria2:
		score += 18 // excellent on unstable networks (QUIC)
	case domain.ProtoTUIC:
		score += 16 // good QUIC performance
	case domain.ProtoShadowsocks:
		score += 8 // simple but less stealth
	}

	// Transport quality per ISP.
	score += transportScore(p.Network, p.Security, isp)

	// Feature bonuses.
	if p.Mux && p.MuxConfig != nil {
		score += 5 // mux improves connection reuse
		if p.MuxConfig.Padding {
			score += 3 // padding helps anti-DPI
		}
	}
	if p.Fragment != "" {
		score += 5 // fragment helps bypass DPI
	}
	if p.ECH {
		score += 4 // ECH hides SNI
	}
	if p.Padding != "" {
		score += 2 // padding helps
	}

	// CDN bonus: more reliable (CDN edge is closer, less prone to blocking).
	if p.GroupName == "" && p.HostHeader != "" {
		score += 8 // likely behind CDN
	}

	// Cap at 100.
	if score > 100 {
		score = 100
	}
	return score
}

// transportScore rates a transport+security combination for a specific ISP.
func transportScore(network, security string, isp domain.ISPPreset) int {
	// Build a key for lookup.
	key := network + "+" + security

	// Per-ISP transport rankings (empirically derived).
	switch isp {
	case domain.ISPHamrahAval:
		// MCI: WS+TLS best, REALITY detected, gRPC OK.
		switch key {
		case "ws+tls":
			return 15
		case "grpc+tls":
			return 10
		case "httpupgrade+tls":
			return 12
		case "tcp+reality":
			return 3 // detected during peak
		case "tcp+tls":
			return 5
		default:
			return 0
		}
	case domain.ISPIrancell:
		// Irancell: gRPC+TLS best, REALITY works off-peak.
		switch key {
		case "grpc+tls":
			return 15
		case "ws+tls":
			return 12
		case "tcp+reality":
			return 10
		case "httpupgrade+tls":
			return 11
		default:
			return 0
		}
	case domain.ISPMokhaberat:
		// TCI: WS+TLS only reliable option, everything else risky.
		switch key {
		case "ws+tls":
			return 15
		case "httpupgrade+tls":
			return 12
		case "grpc+tls":
			return 8
		case "tcp+reality":
			return 2 // heavily detected
		default:
			return 0
		}
	case domain.ISPShatel:
		// Shatel: most things work, REALITY fine.
		switch key {
		case "tcp+reality":
			return 14
		case "ws+tls":
			return 13
		case "grpc+tls":
			return 12
		case "httpupgrade+tls":
			return 11
		default:
			return 5
		}
	case domain.ISPAsiatech:
		switch key {
		case "ws+tls":
			return 14
		case "grpc+tls":
			return 13
		case "tcp+reality":
			return 8
		default:
			return 5
		}
	default:
		// Unknown ISP: conservative scoring.
		switch key {
		case "ws+tls":
			return 10
		case "grpc+tls":
			return 10
		case "tcp+reality":
			return 8
		default:
			return 3
		}
	}
}

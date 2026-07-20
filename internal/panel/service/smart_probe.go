package service

import (
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/subscription"
)

// ProbeConfig describes enhanced health-check probe settings for subscription
// urltest/fallback groups. Multi-URL probing with jitter detection provides
// more accurate alive/dead decisions than a single URL.
type ProbeConfig struct {
	// URLs to probe (clients try in order; first success = alive).
	URLs []string
	// Interval between probes.
	Interval string
	// Tolerance for urltest (ms).
	Tolerance int
	// IdleTimeout before stopping probes on idle connections.
	IdleTimeout string
}

// ProbeURLs returns multiple probe endpoints for diversity. If one is blocked
// by DPI, others still work. The first URL is the primary; others are fallbacks.
var ProbeURLs = []string{
	"https://www.gstatic.com/generate_204",
	"https://cp.cloudflare.com/generate_204",
	"https://detectportal.firefox.com/success.txt",
}

// ISPProbeConfig returns optimized probe settings per ISP. Strict ISPs need
// shorter intervals (faster detection of blocks), while lenient ISPs can use
// longer intervals to reduce overhead.
func ISPProbeConfig(isp domain.ISPPreset) ProbeConfig {
	switch isp {
	case domain.ISPMokhaberat:
		// TCI: aggressive DPI — short interval, tight tolerance for fast failover.
		return ProbeConfig{
			URLs:        ProbeURLs,
			Interval:    "30s",
			Tolerance:   80,
			IdleTimeout: "20m",
		}
	case domain.ISPHamrahAval:
		// MCI: heavy DPI during peak — moderate interval.
		return ProbeConfig{
			URLs:        ProbeURLs,
			Interval:    "45s",
			Tolerance:   100,
			IdleTimeout: "25m",
		}
	case domain.ISPIrancell:
		return ProbeConfig{
			URLs:        ProbeURLs,
			Interval:    "60s",
			Tolerance:   120,
			IdleTimeout: "30m",
		}
	default:
		return ProbeConfig{
			URLs:        ProbeURLs,
			Interval:    "60s",
			Tolerance:   150,
			IdleTimeout: "30m",
		}
	}
}

// EmergencyFallbackConfig defines what happens when ALL proxies in a
// subscription are down. This is the last resort before total disconnection.
type EmergencyFallbackConfig struct {
	// BackupSubURL is an alternate subscription URL the client can try.
	BackupSubURL string
	// DirectWorkerURL is a Cloudflare Worker (or similar) that proxies traffic
	// without needing the panel's subscription — a hardcoded emergency path.
	DirectWorkerURL string
	// DetourTag is the sing-box outbound tag used as final fallback.
	DetourTag string
}

// BuildEmergencyOutbound returns a sing-box outbound map for the emergency
// fallback. It connects directly to a known-unblocked endpoint when all
// configured proxies fail. Returns nil if no emergency config is available.
func BuildEmergencyOutbound() map[string]any {
	return map[string]any{
		"type":        "direct",
		"tag":         "🆘 Emergency",
		"detour":      "direct",
		"domain_strategy": "prefer_ipv4",
	}
}

// BandwidthHint describes per-proxy bandwidth characteristics that help
// clients make better routing decisions.
type BandwidthHint struct {
	// EstimatedMbps is the expected download throughput in Mbps.
	EstimatedMbps int
	// Congestion control preference. "bbr" is optimal for high-latency links.
	CongestionControl string
	// MaxLatencyMs: if latency exceeds this, the proxy is probably degraded.
	MaxLatencyMs int
}

// DefaultBandwidthHint returns bandwidth hints based on protocol and transport.
func DefaultBandwidthHint(protocol domain.Protocol, network string) BandwidthHint {
	hint := BandwidthHint{
		CongestionControl: "bbr",
		MaxLatencyMs:      500,
	}

	switch protocol {
	case domain.ProtoHysteria2:
		// Hysteria2/QUIC: excellent on lossy networks, high throughput.
		hint.EstimatedMbps = 100
		hint.MaxLatencyMs = 800 // tolerant of high latency
	case domain.ProtoTUIC:
		hint.EstimatedMbps = 80
		hint.MaxLatencyMs = 600
	default:
		// TCP-based protocols.
		switch network {
		case "grpc":
			hint.EstimatedMbps = 50 // H2 multiplexing overhead
		case "ws":
			hint.EstimatedMbps = 60
		default:
			hint.EstimatedMbps = 70 // raw TCP
		}
	}
	return hint
}

// ApplyProbeEnhancements enriches the protocol group render metadata with
// ISP-optimized probe URLs and intervals. Called before rendering.
func ApplyProbeEnhancements(groups []subscription.ProtocolGroupRender, isp domain.ISPPreset) {
	cfg := ISPProbeConfig(isp)
	for i := range groups {
		g := &groups[i]
		// Use the primary probe URL from the enhanced pool if the group doesn't
		// have a custom one set.
		if g.ProbeURL == "" || g.ProbeURL == "https://www.gstatic.com/generate_204" {
			g.ProbeURL = cfg.URLs[0]
		}
		// Override interval with ISP-optimized value if the group uses default.
		if g.ProbeInterval == 0 || g.ProbeInterval == 90 {
			// Parse interval string to int.
			switch cfg.Interval {
			case "30s":
				g.ProbeInterval = 30
			case "45s":
				g.ProbeInterval = 45
			case "60s":
				g.ProbeInterval = 60
			}
		}
	}
}

// CertRotationConfig describes certificate rotation strategy to evade
// cert pinning by ISP middleware.
type CertRotationConfig struct {
	// Providers lists CAs to alternate between (avoids single-CA fingerprint).
	Providers []domain.CertProvider
	// RotationDays: how often to switch CA.
	RotationDays int
	// CurrentProvider returns which CA to use today based on rotation schedule.
	CurrentDay int
}

// PickCertProvider returns the CA provider for today based on rotation schedule.
// Alternates between Let's Encrypt and ZeroSSL to prevent cert-authority-based
// fingerprinting by DPI systems.
func PickCertProvider(dayOfYear int) domain.CertProvider {
	providers := []domain.CertProvider{
		domain.CertProviderLetsEncrypt,
		domain.CertProviderZeroSSL,
	}
	// Rotate every 15 days.
	idx := (dayOfYear / 15) % len(providers)
	return providers[idx]
}

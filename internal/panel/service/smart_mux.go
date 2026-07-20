package service

import (
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/subscription"
)

// SmartMuxConfig returns the optimal mux settings for a proxy based on ISP,
// protocol, and transport. Mux is only useful for TCP-based protocols (not
// Hysteria2/TUIC which are QUIC-native). Returns nil when mux should be disabled.
func SmartMuxConfig(isp domain.ISPPreset, protocol domain.Protocol, network, security string) *subscription.MuxSettings {
	// QUIC-based protocols don't benefit from mux (they have native multiplexing).
	switch protocol {
	case domain.ProtoHysteria2, domain.ProtoTUIC:
		return nil
	}

	// REALITY with tcp: mux can reduce stealth — only enable for WS/gRPC.
	if security == "reality" && network == "tcp" {
		return nil
	}

	cfg := &subscription.MuxSettings{
		Protocol:    "smux",
		IdleTimeout: "30s",
		Padding:     true, // always pad mux frames for anti-DPI
		XUDP:        true, // enable UDP-over-TCP for DNS, gaming, etc.
	}

	// ISP-specific tuning.
	switch isp {
	case domain.ISPHamrahAval:
		// MCI: aggressive DPI — use h2mux to blend with HTTPS traffic,
		// lower concurrency to reduce detection surface.
		cfg.Protocol = "h2mux"
		cfg.MaxConnections = 4
		cfg.MinStreams = 2
		cfg.MaxStreams = 8
		cfg.Padding = true
		cfg.IdleTimeout = "60s"
		// MCI peak hours benefit from brutal mode (force throughput).
		cfg.BrutalMode = false

	case domain.ISPIrancell:
		// Irancell: moderate DPI — yamux with higher concurrency works well.
		cfg.Protocol = "yamux"
		cfg.MaxConnections = 6
		cfg.MinStreams = 2
		cfg.MaxStreams = 16
		cfg.Padding = true
		cfg.XUDP = true
		cfg.IdleTimeout = "45s"

	case domain.ISPMokhaberat:
		// TCI: heaviest DPI — h2mux (most stealth), low concurrency,
		// heavy padding, longer idle to avoid reconnection detection.
		cfg.Protocol = "h2mux"
		cfg.MaxConnections = 2
		cfg.MinStreams = 1
		cfg.MaxStreams = 4
		cfg.Padding = true
		cfg.IdleTimeout = "90s"

	case domain.ISPShatel:
		// Shatel: lighter DPI — standard smux is fine, higher throughput.
		cfg.Protocol = "smux"
		cfg.MaxConnections = 8
		cfg.MinStreams = 2
		cfg.MaxStreams = 16
		cfg.Padding = true
		cfg.IdleTimeout = "30s"

	case domain.ISPAsiatech:
		// Asiatech: moderate — yamux balanced settings.
		cfg.Protocol = "yamux"
		cfg.MaxConnections = 6
		cfg.MinStreams = 2
		cfg.MaxStreams = 12
		cfg.Padding = true
		cfg.IdleTimeout = "30s"

	default:
		// Unknown ISP: safe defaults.
		cfg.Protocol = "smux"
		cfg.MaxConnections = 4
		cfg.MinStreams = 1
		cfg.MaxStreams = 8
		cfg.Padding = true
		cfg.IdleTimeout = "30s"
	}

	// Transport-specific adjustments.
	switch network {
	case "grpc":
		// gRPC already multiplexes via HTTP/2 — use h2mux to align.
		cfg.Protocol = "h2mux"
	case "ws":
		// WebSocket: smux or yamux work best.
		if cfg.Protocol == "h2mux" && isp != domain.ISPMokhaberat {
			cfg.Protocol = "smux" // h2mux over WS is redundant unless stealth needed
		}
	}

	return cfg
}

// ApplySmartMux applies ISP-optimized mux settings to all proxies. Called from
// the subscription handler alongside ApplySmartProfiles.
func ApplySmartMux(proxies []subscription.Proxy, isp domain.ISPPreset) {
	for i := range proxies {
		p := &proxies[i]
		// Skip if mux is explicitly disabled (no Mux flag) and no smart profile set it.
		if !p.Mux {
			continue
		}
		// If Mux=true but no detailed config, compute optimal settings.
		if p.MuxConfig == nil {
			p.MuxConfig = SmartMuxConfig(isp, p.Protocol, p.Network, p.Security)
			// If SmartMuxConfig says no mux (e.g. QUIC protocol), disable it.
			if p.MuxConfig == nil {
				p.Mux = false
			}
		}
	}
}

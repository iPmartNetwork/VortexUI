package service

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/subscription"
)

// SmartConfigEngine selects the optimal anti-censorship profile for each proxy
// based on ISP, protocol, transport, time-of-day, and CDN presence. It replaces
// the basic AutoConfigEngine with a per-proxy granular approach that integrates
// directly into the subscription rendering pipeline.
type SmartConfigEngine struct {
	now func() time.Time
}

// NewSmartConfigEngine creates a new engine using real time.
func NewSmartConfigEngine() *SmartConfigEngine {
	return &SmartConfigEngine{now: time.Now}
}

// ProfileFor returns the optimal SmartProfile for a proxy with the given
// characteristics. The profile is additive: fields left at zero-value mean
// "don't override the proxy's existing value."
func (e *SmartConfigEngine) ProfileFor(isp domain.ISPPreset, protocol domain.Protocol, transport, security string, cdn domain.CDNProvider) domain.SmartProfile {
	hour := e.now().Hour()
	isPeak := hour >= 18 && hour <= 23 // Iran DPI peaks 18:00–23:00 IRST

	// Base severity by ISP.
	severity := e.severityFor(isp, isPeak)

	p := domain.SmartProfile{
		Fingerprint: "chrome",
		Severity:    severity,
	}

	// --- Fragment settings based on severity ---
	switch severity {
	case domain.SmartProfileSeverityAggressive:
		p.FragmentEnabled = true
		p.FragmentPackets = "tlshello"
		p.MuxEnabled = true
		p.MuxConcurrency = 8
		p.PaddingEnabled = true
		p.ECHEnabled = true
		p.ALPN = []string{"h2", "http/1.1"}
		switch isp {
		case domain.ISPMokhaberat:
			p.FragmentSize = "50-100"
			p.FragmentInterval = "20-50"
			p.PaddingSize = "200-300"
		case domain.ISPHamrahAval:
			p.FragmentSize = "10-50"
			p.FragmentInterval = "10-20"
			p.PaddingSize = "100-200"
		default:
			p.FragmentSize = "20-60"
			p.FragmentInterval = "10-30"
			p.PaddingSize = "100-200"
		}
		p.Reason = "aggressive: peak hours or strict ISP — full evasion stack"

	case domain.SmartProfileSeverityModerate:
		p.FragmentEnabled = true
		p.FragmentPackets = "tlshello"
		p.MuxEnabled = true
		p.MuxConcurrency = 4
		p.PaddingEnabled = true
		switch isp {
		case domain.ISPIrancell:
			p.FragmentSize = "1-3"
			p.FragmentInterval = "1-5"
			p.FragmentPackets = "1-3"
			p.PaddingSize = "50-100"
			p.MuxXudp = true
		case domain.ISPShatel:
			p.FragmentSize = "10-30"
			p.FragmentInterval = "5-15"
			p.PaddingSize = "100-150"
		default:
			p.FragmentSize = "10-50"
			p.FragmentInterval = "10-20"
			p.PaddingSize = "100-200"
		}
		p.Reason = "moderate: standard anti-DPI with fragment+mux+padding"

	case domain.SmartProfileSeverityLight:
		// Light: only fingerprint and small padding — don't fragment.
		p.PaddingEnabled = true
		p.PaddingSize = "50-100"
		p.Reason = "light: low-restriction ISP — fingerprint+padding sufficient"
	}

	// --- Protocol-specific adjustments ---
	switch protocol {
	case domain.ProtoHysteria2:
		// Hysteria2 is UDP/QUIC — fragment/mux don't apply.
		p.FragmentEnabled = false
		p.MuxEnabled = false
		p.PaddingEnabled = false
		p.ECHEnabled = false
		p.Reason = "hysteria2: UDP transport — no TLS tricks needed"

	case domain.ProtoTUIC:
		// TUIC is also QUIC-based.
		p.FragmentEnabled = false
		p.MuxEnabled = false
		p.PaddingEnabled = false
		p.ECHEnabled = false
		p.Reason = "tuic: QUIC transport — no TLS tricks needed"
	}

	// --- Security-specific adjustments ---
	if security == "reality" {
		// REALITY doesn't need fragment (it's already camouflaged), but benefits
		// from a good fingerprint and proper server names.
		p.FragmentEnabled = false
		p.ECHEnabled = false
		p.MuxEnabled = false // mux can reduce REALITY stealth
		p.RealityFingerprint = "chrome"
		if names, ok := domain.RealityServerNamePool[isp]; ok && len(names) > 0 {
			// Deterministic rotation: pick server name based on day-of-year.
			dayIdx := e.now().YearDay() % len(names)
			p.RealityServerNames = []string{names[dayIdx]}
		}
		p.Reason = "reality: camouflaged — fingerprint+server_name rotation only"
	}

	// --- CDN-specific adjustments ---
	if cdn != domain.CDNNone && cdn != "" {
		applySmartCDN(&p, cdn, transport)
	}

	return p
}

// severityFor determines the anti-DPI severity level for a given ISP and time.
func (e *SmartConfigEngine) severityFor(isp domain.ISPPreset, isPeak bool) string {
	switch isp {
	case domain.ISPMokhaberat:
		// TCI always aggressive — strongest DPI in Iran.
		return domain.SmartProfileSeverityAggressive
	case domain.ISPHamrahAval:
		if isPeak {
			return domain.SmartProfileSeverityAggressive
		}
		return domain.SmartProfileSeverityModerate
	case domain.ISPIrancell:
		if isPeak {
			return domain.SmartProfileSeverityModerate
		}
		return domain.SmartProfileSeverityLight
	case domain.ISPShatel:
		if isPeak {
			return domain.SmartProfileSeverityModerate
		}
		return domain.SmartProfileSeverityLight
	case domain.ISPAsiatech:
		return domain.SmartProfileSeverityModerate
	default:
		if isPeak {
			return domain.SmartProfileSeverityModerate
		}
		return domain.SmartProfileSeverityLight
	}
}

// GenerateShortID produces a deterministic ShortID for REALITY that rotates
// daily per-inbound. This prevents a static short_id from being fingerprinted.
func (e *SmartConfigEngine) GenerateShortID(inboundTag string) string {
	day := e.now().Format("2006-01-02")
	h := sha256.Sum256([]byte("reality-sid:" + inboundTag + ":" + day))
	return hex.EncodeToString(h[:4]) // 8-char hex short ID
}

// applyForCDN sets CDN-optimized settings on the profile.
func applySmartCDN(p *domain.SmartProfile, cdn domain.CDNProvider, transport string) {
	// CDN connections don't need fragment (TLS terminates at CDN edge).
	p.FragmentEnabled = false
	p.ECHEnabled = false

	switch cdn {
	case domain.CDNCloudflare:
		p.CDNHost = "" // set by SubHost.HostHeader
		p.EarlyData = true
		switch transport {
		case "ws":
			p.CDNPath = "/ws"
			p.ALPN = []string{"http/1.1"} // CF WebSocket requires HTTP/1.1
		case "grpc":
			p.ALPN = []string{"h2"}
		}
		p.Reason = "cloudflare CDN: early-data enabled, no fragment needed"

	case domain.CDNArvanCloud:
		p.EarlyData = true
		if transport == "ws" {
			p.CDNPath = "/ws"
			p.ALPN = []string{"http/1.1"}
		}
		p.Reason = "arvancloud CDN: WS optimized, no fragment"

	case domain.CDNGcore:
		if transport == "ws" {
			p.CDNPath = "/ws"
		}
		p.ALPN = []string{"h2", "http/1.1"}
		p.Reason = "gcore CDN: h2 multiplex, no fragment"

	default:
		p.Reason = "CDN detected: fragment disabled, TLS terminates at edge"
	}
}

// ApplySmartProfiles applies the SmartConfigEngine's recommendations to all
// proxies based on the detected ISP. It enriches each proxy with optimal
// anti-DPI settings without overriding explicitly-set per-inbound evasion
// profiles. Called from the subscription handler after ISP detection.
func ApplySmartProfiles(proxies []subscription.Proxy, isp domain.ISPPreset, cdn domain.CDNProvider) {
	if isp == "" && cdn == "" {
		return
	}
	engine := NewSmartConfigEngine()

	for i := range proxies {
		p := &proxies[i]

		// Skip proxies that already have a per-inbound evasion profile applied
		// (fragment is the tell-tale sign).
		if p.Fragment != "" {
			continue
		}

		profile := engine.ProfileFor(isp, p.Protocol, p.Network, p.Security, cdn)

		// Apply fragment.
		if profile.FragmentEnabled && p.Fragment == "" {
			size := profile.FragmentSize
			if size == "" {
				size = "10-50"
			}
			interval := profile.FragmentInterval
			if interval == "" {
				interval = "10-20"
			}
			packets := profile.FragmentPackets
			if packets == "" {
				packets = "tlshello"
			}
			p.Fragment = size + "," + interval + "," + packets
		}

		// Apply mux.
		if profile.MuxEnabled && !p.Mux {
			p.Mux = true
		}

		// Apply padding.
		if profile.PaddingEnabled && p.Padding == "" {
			p.Padding = profile.PaddingSize
		}

		// Apply fingerprint (only if not already set).
		if profile.Fingerprint != "" && p.Fingerprint == "" {
			p.Fingerprint = profile.Fingerprint
		}

		// Apply ECH.
		if profile.ECHEnabled && !p.ECH {
			p.ECH = true
		}

		// Apply ALPN (only if not already set by host override).
		if len(profile.ALPN) > 0 && len(p.ALPN) == 0 {
			p.ALPN = profile.ALPN
		}

		// Apply REALITY server name rotation.
		if p.Security == "reality" && len(profile.RealityServerNames) > 0 && p.SNI == "" {
			p.SNI = profile.RealityServerNames[0]
		}
	}
}

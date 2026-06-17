package domain

import (
	"github.com/google/uuid"
)

// EvasionProfile is a reusable DPI-evasion preset that can be linked to
// inbounds. It bundles fragment settings, mux configuration, and uTLS
// fingerprint selection so operators apply hardening with one click.
type EvasionProfile struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`

	// Fragment splits TLS ClientHello into smaller chunks to bypass DPI.
	FragmentEnabled  bool   `json:"fragment_enabled"`
	FragmentLength   string `json:"fragment_length,omitempty"`   // e.g. "10-30"
	FragmentInterval string `json:"fragment_interval,omitempty"` // e.g. "10-20" (ms)
	FragmentPackets  string `json:"fragment_packets,omitempty"`  // "tlshello" or "1-3"

	// Mux multiplexes connections to reduce handshake overhead and confuse DPI.
	MuxEnabled     bool   `json:"mux_enabled"`
	MuxProtocol    string `json:"mux_protocol,omitempty"`    // "smux" | "yamux" | "h2mux"
	MuxMaxStreams  int    `json:"mux_max_streams,omitempty"` // 0 = default
	MuxPadding     bool   `json:"mux_padding"`
	MuxBrutal      bool   `json:"mux_brutal"`                // brutal mode (sing-box)

	// uTLS fingerprint mimics a real browser's TLS ClientHello.
	Fingerprint string `json:"fingerprint,omitempty"` // "chrome" | "firefox" | "safari" | "random" | "randomized"

	// Noise injects random padding bytes into packets.
	NoiseEnabled bool   `json:"noise_enabled"`
	NoisePacket  string `json:"noise_packet,omitempty"` // base64-encoded noise pattern

	Enabled bool `json:"enabled"`
}

// DefaultProfiles returns built-in presets operators can use immediately.
func DefaultProfiles() []EvasionProfile {
	return []EvasionProfile{
		{
			Name:             "Iran (Fragment + Chrome)",
			Description:      "Optimized for Iranian DPI: TLS fragment + Chrome fingerprint",
			FragmentEnabled:  true,
			FragmentLength:   "10-30",
			FragmentInterval: "10-20",
			FragmentPackets:  "tlshello",
			Fingerprint:      "chrome",
			Enabled:          true,
		},
		{
			Name:            "China (Mux + Random FP)",
			Description:     "Multiplexed connections with randomized fingerprint",
			MuxEnabled:      true,
			MuxProtocol:     "h2mux",
			MuxMaxStreams:   8,
			MuxPadding:      true,
			Fingerprint:     "randomized",
			Enabled:         true,
		},
		{
			Name:             "Russia (Fragment + Firefox)",
			Description:      "Fragment with Firefox fingerprint for TSPU bypass",
			FragmentEnabled:  true,
			FragmentLength:   "1-3",
			FragmentInterval: "5-10",
			FragmentPackets:  "1-3",
			Fingerprint:      "firefox",
			Enabled:          true,
		},
	}
}

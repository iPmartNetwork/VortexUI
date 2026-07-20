package domain

// SmartProfile describes the complete set of anti-censorship optimizations to
// apply to a proxy configuration. It combines transport, TLS, fragment, padding,
// mux, and fingerprint settings into a single coherent profile that the
// subscription renderer applies to each proxy based on detected ISP and conditions.
type SmartProfile struct {
	// Transport layer recommendations.
	PreferredTransport string `json:"preferred_transport,omitempty"` // ws, grpc, httpupgrade, tcp
	PreferredSecurity  string `json:"preferred_security,omitempty"` // tls, reality, none

	// TLS Fragment — split TLS ClientHello to evade DPI.
	FragmentEnabled  bool   `json:"fragment_enabled"`
	FragmentSize     string `json:"fragment_size"`     // e.g. "10-50" bytes
	FragmentInterval string `json:"fragment_interval"` // e.g. "10-20" ms
	FragmentPackets  string `json:"fragment_packets"`  // "tlshello" | "1-3"

	// Mux — multiplex streams to mask traffic patterns.
	MuxEnabled     bool `json:"mux_enabled"`
	MuxConcurrency int  `json:"mux_concurrency"`
	MuxXudp        bool `json:"mux_xudp"`

	// uTLS fingerprint — mimic a real browser's TLS handshake.
	Fingerprint string `json:"fingerprint"` // "chrome", "firefox", "safari", "random"

	// Padding — random padding to defeat length-based DPI.
	PaddingEnabled bool   `json:"padding_enabled"`
	PaddingSize    string `json:"padding_size"` // e.g. "100-200"

	// ECH (Encrypted Client Hello) — hide SNI from middleboxes.
	ECHEnabled bool `json:"ech_enabled"`

	// ALPN override — force specific ALPN for anti-fingerprinting.
	ALPN []string `json:"alpn,omitempty"` // e.g. ["h2", "http/1.1"]

	// CDN-specific settings.
	CDNHost     string `json:"cdn_host,omitempty"`      // Host header for CDN routing
	CDNPath     string `json:"cdn_path,omitempty"`      // WebSocket/gRPC path prefix for workers
	EarlyData   bool   `json:"early_data,omitempty"`    // 0-RTT for WS early data
	NoTLSVerify bool   `json:"no_tls_verify,omitempty"` // skip cert verify for self-signed

	// REALITY-specific optimizations.
	RealityServerNames []string `json:"reality_server_names,omitempty"` // recommended SNIs
	RealityFingerprint string   `json:"reality_fingerprint,omitempty"`  // "chrome" default

	// Metadata for debugging/display.
	Reason   string `json:"reason,omitempty"`
	Severity string `json:"severity,omitempty"` // "light", "moderate", "aggressive"
}

// SmartProfileSeverityLight means basic anti-DPI (fingerprint + small fragment).
const SmartProfileSeverityLight = "light"

// SmartProfileSeverityModerate means fragment + mux + padding.
const SmartProfileSeverityModerate = "moderate"

// SmartProfileSeverityAggressive means full stack: fragment + mux + padding + ECH.
const SmartProfileSeverityAggressive = "aggressive"

// CDNProvider identifies known CDN providers for config optimization.
type CDNProvider string

const (
	CDNCloudflare  CDNProvider = "cloudflare"
	CDNArvanCloud  CDNProvider = "arvancloud"
	CDNGcore       CDNProvider = "gcore"
	CDNFastly      CDNProvider = "fastly"
	CDNNone        CDNProvider = "none"
)

// RealityServerNamePool contains known-good REALITY server names per ISP.
// These are real websites that have similar TLS fingerprints to our REALITY
// connections, making it harder for DPI to distinguish.
var RealityServerNamePool = map[ISPPreset][]string{
	ISPHamrahAval: {
		"www.google.com",
		"www.microsoft.com",
		"www.samsung.com",
		"www.speedtest.net",
		"mail.yahoo.com",
		"discord.com",
		"www.tesla.com",
	},
	ISPIrancell: {
		"www.google.com",
		"www.apple.com",
		"www.nvidia.com",
		"www.cloudflare.com",
		"www.yahoo.com",
		"account.microsoft.com",
	},
	ISPMokhaberat: {
		"www.google.com",
		"www.microsoft.com",
		"www.bing.com",
		"www.samsung.com",
		"www.nvidia.com",
		"www.speedtest.net",
	},
	ISPShatel: {
		"www.google.com",
		"www.microsoft.com",
		"www.apple.com",
		"www.cloudflare.com",
	},
	ISPAsiatech: {
		"www.google.com",
		"www.microsoft.com",
		"www.apple.com",
		"www.nvidia.com",
	},
}

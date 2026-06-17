package domain

import "time"

// DoHConfig stores the DNS-over-HTTPS server configuration.
type DoHConfig struct {
	Enabled        bool     `json:"enabled"`
	ListenAddr     string   `json:"listen_addr"`      // e.g. ":8053"
	UpstreamDNS    []string `json:"upstream_dns"`     // e.g. ["1.1.1.1", "8.8.8.8"]
	BlockAds       bool     `json:"block_ads"`
	BlockMalware   bool     `json:"block_malware"`
	CustomBlocklist []string `json:"custom_blocklist"` // domains to block
	LogQueries     bool     `json:"log_queries"`
	CacheTTL       int      `json:"cache_ttl"`        // seconds; 0 = no cache
}

// DefaultDoHConfig returns sensible defaults.
func DefaultDoHConfig() DoHConfig {
	return DoHConfig{
		Enabled:      false,
		ListenAddr:   ":8053",
		UpstreamDNS:  []string{"1.1.1.1", "8.8.8.8"},
		BlockAds:     false,
		BlockMalware: true,
		CacheTTL:     300,
		LogQueries:   false,
	}
}

// DoHQueryLog is a recorded DNS query.
type DoHQueryLog struct {
	Domain    string    `json:"domain"`
	Type      string    `json:"type"` // A, AAAA, CNAME, etc.
	ClientIP  string    `json:"client_ip"`
	Blocked   bool      `json:"blocked"`
	LatencyMS int       `json:"latency_ms"`
	Timestamp time.Time `json:"timestamp"`
}

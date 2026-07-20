package service

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/subscription"
)

// SNIPool contains high-traffic domains per ISP that are known to NOT be blocked
// and have similar TLS fingerprints. These are rotated per-connection using a
// deterministic seed so each subscription fetch yields a different but stable SNI.
var SNIPool = map[domain.ISPPreset][]string{
	domain.ISPHamrahAval: {
		"www.google.com",
		"www.googleapis.com",
		"clients4.google.com",
		"www.microsoft.com",
		"login.microsoftonline.com",
		"www.samsung.com",
		"www.speedtest.net",
		"www.nvidia.com",
		"discord.com",
		"gateway.discord.gg",
		"cdn.discordapp.com",
		"www.tesla.com",
		"store.steampowered.com",
	},
	domain.ISPIrancell: {
		"www.google.com",
		"www.apple.com",
		"www.icloud.com",
		"www.nvidia.com",
		"www.cloudflare.com",
		"one.one.one.one",
		"www.yahoo.com",
		"account.microsoft.com",
		"outlook.live.com",
		"www.amazon.com",
		"www.whatsapp.net",
	},
	domain.ISPMokhaberat: {
		"www.google.com",
		"www.gstatic.com",
		"fonts.googleapis.com",
		"www.microsoft.com",
		"www.bing.com",
		"www.samsung.com",
		"www.nvidia.com",
		"www.speedtest.net",
		"www.office.com",
		"teams.microsoft.com",
	},
	domain.ISPShatel: {
		"www.google.com",
		"www.microsoft.com",
		"www.apple.com",
		"www.cloudflare.com",
		"www.amazon.com",
		"www.github.com",
		"www.stackoverflow.com",
	},
	domain.ISPAsiatech: {
		"www.google.com",
		"www.microsoft.com",
		"www.apple.com",
		"www.nvidia.com",
		"www.cloudflare.com",
		"www.github.com",
	},
}

// DefaultSNIPool is used when the ISP is unknown.
var DefaultSNIPool = []string{
	"www.google.com",
	"www.microsoft.com",
	"www.apple.com",
	"www.cloudflare.com",
	"www.amazon.com",
	"www.nvidia.com",
}

// ApplyDynamicSNI assigns randomized SNIs from the ISP-specific pool to proxies
// that use REALITY security and don't already have a fixed SNI. The rotation is
// deterministic per proxy+day so the same subscription fetch returns consistent
// results within a day, but different proxies get different SNIs for diversity.
func ApplyDynamicSNI(proxies []subscription.Proxy, isp domain.ISPPreset) {
	pool := SNIPool[isp]
	if len(pool) == 0 {
		pool = DefaultSNIPool
	}
	if len(pool) == 0 {
		return
	}

	day := time.Now().Format("2006-01-02")

	for i := range proxies {
		p := &proxies[i]

		// Only apply to REALITY proxies — TLS proxies need their real cert domain.
		if p.Security != "reality" {
			continue
		}
		// Don't override explicitly-set SNI from inbound config.
		if p.SNI != "" {
			// But if the current SNI is in our pool, that's fine — it was already
			// set by the smart config engine. Rotate it for diversity.
			if !isInPool(p.SNI, pool) {
				continue
			}
		}

		// Deterministic rotation: hash(proxy_name + day) → index into pool.
		// This gives each proxy a different SNI, but it's stable within a day.
		p.SNI = pickSNI(pool, p.Name, day)
	}
}

// pickSNI deterministically selects an SNI from the pool based on seed inputs.
func pickSNI(pool []string, proxyName, day string) string {
	h := sha256.Sum256([]byte("sni:" + proxyName + ":" + day))
	idx := binary.BigEndian.Uint32(h[:4]) % uint32(len(pool))
	return pool[idx]
}

func isInPool(sni string, pool []string) bool {
	for _, s := range pool {
		if s == sni {
			return true
		}
	}
	return false
}

// GRPCServiceNames provides obfuscated gRPC service names that look like
// legitimate Google/Microsoft gRPC services. DPI systems inspect the gRPC
// service name in the HTTP/2 path header; using realistic names reduces detection.
var GRPCServiceNames = []string{
	"grpc.health.v1.Health",
	"google.internal.communications.v1",
	"google.pubsub.v2.Publisher",
	"com.microsoft.azure.servicebus",
	"envoy.service.discovery.v3",
	"google.cloud.bigquery.storage.v1",
}

// PickGRPCServiceName returns a deterministic obfuscated gRPC service name
// for a proxy, rotating daily.
func PickGRPCServiceName(proxyName string) string {
	day := time.Now().Format("2006-01-02")
	h := sha256.Sum256([]byte("grpc:" + proxyName + ":" + day))
	idx := binary.BigEndian.Uint32(h[:4]) % uint32(len(GRPCServiceNames))
	return GRPCServiceNames[idx]
}

// ApplyTransportOptimization enhances gRPC and HTTPUpgrade transports with
// anti-detection features: obfuscated service names, proper authority headers,
// and HTTP/2 fingerprint settings.
func ApplyTransportOptimization(proxies []subscription.Proxy, isp domain.ISPPreset) {
	for i := range proxies {
		p := &proxies[i]

		switch p.Network {
		case "grpc":
			// Obfuscate gRPC service name if it looks default/empty.
			if p.Path == "" || p.Path == "grpc" || p.Path == "GunService" {
				p.Path = PickGRPCServiceName(p.Name)
			}
			// Ensure proper authority header for H2 (looks like a real gRPC call).
			if p.HostHeader == "" && p.SNI != "" {
				p.HostHeader = p.SNI
			}

		case "httpupgrade":
			// HTTPUpgrade: ensure path looks like a legitimate web endpoint.
			if p.Path == "" {
				p.Path = "/api/v1/stream"
			}
			// Randomize User-Agent-like headers via host header.
			if p.HostHeader == "" && p.SNI != "" {
				p.HostHeader = p.SNI
			}

		case "ws":
			// WS: ensure path doesn't look suspicious (no bare /ws).
			if p.Path == "/ws" || p.Path == "" {
				// Use a more natural-looking path.
				p.Path = "/api/v2/ws"
			}
		}
	}
}

// Package config loads strongly-typed configuration from environment variables
// with sane defaults. Fail-fast validation at startup is preferred over
// discovering a missing secret at request time.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Panel holds control-plane configuration.
type Panel struct {
	HTTPAddr    string
	GRPCAddr    string // hub listener for node agents
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	JWTTTL      time.Duration

	// mTLS material for the node hub.
	TLSCert string
	TLSKey  string
	TLSCA   string

	// Optional outbound notifications (empty = disabled).
	WebhookURL     string
	WebhookSecret  string
	TelegramToken  string
	TelegramChatID string

	// Optional Cloudflare DNS automation (auto-create A records for nodes).
	CloudflareToken  string
	CloudflareZoneID string

	// Optional in-process local node: run a proxy core on the panel host itself,
	// managed in-process (no gRPC agent). Empty/false = disabled.
	LocalNode     bool
	LocalNodeName string // node name/identity in the DB (default "local")
	LocalNodeHost string // public host advertised in subscriptions (default 127.0.0.1)
	Core          string // xray | singbox
	CoreBin       string // path to the core binary
	CoreConfig    string // where the rendered core config is written
	CoreAPIPort   int    // loopback port for the core's stats/control API
	// SingboxV2RayAPI controls whether the sing-box config emits the
	// experimental.v2ray_api block (per-user stats). Default true; set false for
	// sing-box binaries built without the with_v2ray_api tag, which reject it.
	SingboxV2RayAPI bool
	// ShareAutoLimit, when true, makes the account-sharing guard actually limit
	// (deprovision) users caught online from more IPs than their device limit.
	// Default false: detection only emits an alert event.
	ShareAutoLimit bool

	// GeoIPDB is the path to a MaxMind GeoLite2-Country.mmdb database used to
	// resolve subscription-fetch IPs to countries for the "Traffic by Country"
	// analytics. Empty = feature disabled (degrades gracefully).
	GeoIPDB string
}

// Node holds node-agent configuration. The agent is a gRPC *server* (the panel
// dials it), so it needs a listen address and mTLS material plus the local core
// binary/config paths.
type Node struct {
	ListenAddr string // where the NodeService gRPC server listens
	Core       string // which engine this node runs: xray | singbox
	CoreBin    string // path to the core binary
	CoreConfig string // where the agent writes the rendered core config
	APIPort    int    // loopback port for the core's stats/control API

	// SingboxV2RayAPI controls whether the sing-box config emits the
	// experimental.v2ray_api block (per-user stats). Default true; set false for
	// sing-box binaries built without the with_v2ray_api tag, which reject it.
	SingboxV2RayAPI bool

	TLSCert string
	TLSKey  string
	TLSCA   string
}

// LoadNode reads node-agent config from the environment, validating required keys.
func LoadNode() (*Node, error) {
	core := env("VORTEX_CORE", "xray")
	c := &Node{
		ListenAddr:      env("VORTEX_NODE_LISTEN", ":50051"),
		Core:            core,
		CoreBin:         env("VORTEX_CORE_BIN", core),
		CoreConfig:      env("VORTEX_CORE_CONFIG", "/etc/vortex/core.json"),
		APIPort:         envInt("VORTEX_CORE_API_PORT", 10085),
		SingboxV2RayAPI: envBool("VORTEX_SINGBOX_V2RAY_API", true),
		TLSCert:         os.Getenv("VORTEX_TLS_CERT"),
		TLSKey:          os.Getenv("VORTEX_TLS_KEY"),
		TLSCA:           os.Getenv("VORTEX_TLS_CA"),
	}
	var errs []error
	if c.TLSCert == "" || c.TLSKey == "" || c.TLSCA == "" {
		errs = append(errs, errors.New("VORTEX_TLS_CERT, VORTEX_TLS_KEY and VORTEX_TLS_CA are required for mTLS"))
	}
	return c, errors.Join(errs...)
}

// LoadPanel reads panel config from the environment, validating required keys.
func LoadPanel() (*Panel, error) {
	c := &Panel{
		HTTPAddr:         env("VORTEX_HTTP_ADDR", ":8080"),
		GRPCAddr:         env("VORTEX_GRPC_ADDR", ":50051"),
		DatabaseURL:      os.Getenv("VORTEX_DATABASE_URL"),
		RedisURL:         env("VORTEX_REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:        os.Getenv("VORTEX_JWT_SECRET"),
		JWTTTL:           envDur("VORTEX_JWT_TTL", time.Hour),
		TLSCert:          os.Getenv("VORTEX_TLS_CERT"),
		TLSKey:           os.Getenv("VORTEX_TLS_KEY"),
		TLSCA:            os.Getenv("VORTEX_TLS_CA"),
		WebhookURL:       os.Getenv("VORTEX_WEBHOOK_URL"),
		WebhookSecret:    os.Getenv("VORTEX_WEBHOOK_SECRET"),
		TelegramToken:    os.Getenv("VORTEX_TELEGRAM_TOKEN"),
		TelegramChatID:   os.Getenv("VORTEX_TELEGRAM_CHAT_ID"),
		CloudflareToken:  os.Getenv("VORTEX_CF_API_TOKEN"),
		CloudflareZoneID: os.Getenv("VORTEX_CF_ZONE_ID"),
		LocalNode:        envBool("VORTEX_LOCAL_NODE", false),
		LocalNodeName:    env("VORTEX_LOCAL_NODE_NAME", "local"),
		LocalNodeHost:    env("VORTEX_LOCAL_NODE_HOST", "127.0.0.1"),
		Core:             env("VORTEX_CORE", "xray"),
		CoreBin:          os.Getenv("VORTEX_CORE_BIN"),
		CoreConfig:       env("VORTEX_CORE_CONFIG", "/etc/vortex/local-core.json"),
		CoreAPIPort:      envInt("VORTEX_CORE_API_PORT", 10085),
		SingboxV2RayAPI:  envBool("VORTEX_SINGBOX_V2RAY_API", true),
		ShareAutoLimit:   envBool("VORTEX_SHARE_AUTOLIMIT", false),
		GeoIPDB:          env("VORTEX_GEOIP_DB", ""),
	}
	if c.CoreBin == "" {
		c.CoreBin = c.Core // resolve from PATH by core name
	}
	var errs []error
	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("VORTEX_DATABASE_URL is required"))
	}
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("VORTEX_JWT_SECRET must be at least 32 bytes"))
	}
	return c, errors.Join(errs...)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDur(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// String renders a redacted summary safe for logs.
func (p *Panel) String() string {
	return fmt.Sprintf("Panel{http=%s grpc=%s db=%s redis=%s jwtTTL=%s}",
		p.HTTPAddr, p.GRPCAddr, redact(p.DatabaseURL), p.RedisURL, p.JWTTTL)
}

func redact(url string) string {
	if url == "" {
		return "<unset>"
	}
	return "<set>"
}

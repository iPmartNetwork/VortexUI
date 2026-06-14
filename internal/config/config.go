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

	TLSCert string
	TLSKey  string
	TLSCA   string
}

// LoadNode reads node-agent config from the environment, validating required keys.
func LoadNode() (*Node, error) {
	core := env("VORTEX_CORE", "xray")
	c := &Node{
		ListenAddr: env("VORTEX_NODE_LISTEN", ":50051"),
		Core:       core,
		CoreBin:    env("VORTEX_CORE_BIN", core),
		CoreConfig: env("VORTEX_CORE_CONFIG", "/etc/vortex/core.json"),
		APIPort:    envInt("VORTEX_CORE_API_PORT", 10085),
		TLSCert:    os.Getenv("VORTEX_TLS_CERT"),
		TLSKey:     os.Getenv("VORTEX_TLS_KEY"),
		TLSCA:      os.Getenv("VORTEX_TLS_CA"),
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
		HTTPAddr:    env("VORTEX_HTTP_ADDR", ":8080"),
		GRPCAddr:    env("VORTEX_GRPC_ADDR", ":50051"),
		DatabaseURL: os.Getenv("VORTEX_DATABASE_URL"),
		RedisURL:    env("VORTEX_REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:   os.Getenv("VORTEX_JWT_SECRET"),
		JWTTTL:      envDur("VORTEX_JWT_TTL", time.Hour),
		TLSCert:     os.Getenv("VORTEX_TLS_CERT"),
		TLSKey:      os.Getenv("VORTEX_TLS_KEY"),
		TLSCA:       os.Getenv("VORTEX_TLS_CA"),
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

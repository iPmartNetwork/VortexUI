// Package geoip resolves IP addresses to ISO 3166-1 alpha-2 country codes using
// a MaxMind GeoLite2-Country database. A disabled resolver (no DB configured)
// returns "" for every lookup, so geo features degrade gracefully.
package geoip

import (
	"net"
	"strings"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

// Resolver wraps a MaxMind database reader. The zero value (and a nil pointer)
// is a valid disabled resolver.
type Resolver struct {
	mu sync.RWMutex
	db *maxminddb.Reader
}

// Open loads the mmdb at path. An empty path returns a disabled resolver with no
// error. A non-empty path that cannot be opened returns an error so the caller
// can log it and fall back to a disabled resolver.
func Open(path string) (*Resolver, error) {
	if path == "" {
		return &Resolver{}, nil
	}
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	return &Resolver{db: db}, nil
}

// Disabled returns a resolver that always yields "".
func Disabled() *Resolver { return &Resolver{} }

// Enabled reports whether a database is loaded.
func (r *Resolver) Enabled() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.db != nil
}

// Country returns the ISO country code for ip, or "" if unknown/disabled.
func (r *Resolver) Country(ip string) string {
	if r == nil {
		return ""
	}
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()
	if db == nil {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	var rec struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	if err := db.Lookup(parsed, &rec); err != nil {
		return ""
	}
	return rec.Country.ISOCode
}

// ISP attempts to identify the Internet Service Provider for the given IP
// address. It first checks the MaxMind DB for ASN/organization data; if that is
// unavailable, it falls back to a static table of well-known Iranian IP blocks.
// Returns a short ISP identifier (e.g. "mci", "irancell") or "" if unknown.
func (r *Resolver) ISP(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}

	// Try MaxMind ASN lookup (works with GeoLite2-ASN and GeoIP2-ISP DBs).
	if r != nil {
		r.mu.RLock()
		db := r.db
		r.mu.RUnlock()
		if db != nil {
			var rec struct {
				AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
			}
			if err := db.Lookup(parsed, &rec); err == nil && rec.AutonomousSystemOrganization != "" {
				if isp := mapOrgToISP(rec.AutonomousSystemOrganization); isp != "" {
					return isp
				}
			}
		}
	}

	// Static fallback: well-known Iranian IP blocks.
	return ispFromStaticRanges(parsed)
}

// mapOrgToISP maps MaxMind organization names to short ISP identifiers.
func mapOrgToISP(org string) string {
	// Normalize to lowercase for matching.
	lower := strings.ToLower(org)
	switch {
	case strings.Contains(lower, "mobile communication") || strings.Contains(lower, "mci") || strings.Contains(lower, "hamrah"):
		return "mci"
	case strings.Contains(lower, "irancell") || strings.Contains(lower, "mtn"):
		return "irancell"
	case strings.Contains(lower, "telecommunication infrastructure") || strings.Contains(lower, "tci") || strings.Contains(lower, "mokhaberat"):
		return "mokhaberat"
	case strings.Contains(lower, "shatel"):
		return "shatel"
	case strings.Contains(lower, "asiatech"):
		return "asiatech"
	case strings.Contains(lower, "rightel"):
		return "rightel"
	case strings.Contains(lower, "pars online") || strings.Contains(lower, "parsonline"):
		return "parsonline"
	case strings.Contains(lower, "afranet"):
		return "afranet"
	case strings.Contains(lower, "mobinnet"):
		return "mobinnet"
	default:
		return ""
	}
}

// ispFromStaticRanges checks well-known Iranian ISP IP ranges. This is a
// best-effort fallback for when no ASN database is loaded. The ranges cover the
// primary RIPE-allocated blocks for major Iranian operators.
func ispFromStaticRanges(ip net.IP) string {
	for _, entry := range iranISPRanges {
		if entry.network.Contains(ip) {
			return entry.isp
		}
	}
	return ""
}

type ispRange struct {
	network *net.IPNet
	isp     string
}

// mustParseCIDR is a helper for static initialization; panics on bad input.
func mustParseCIDR(cidr string) *net.IPNet {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		panic("geoip: bad CIDR in static ISP table: " + cidr)
	}
	return n
}

// iranISPRanges maps major Iranian ISP IP blocks to short identifiers.
// Source: RIPE NCC allocations. Updated periodically.
var iranISPRanges = []ispRange{
	// MCI (Hamrah Aval) — AS197207
	{mustParseCIDR("5.106.0.0/16"), "mci"},
	{mustParseCIDR("5.107.0.0/16"), "mci"},
	{mustParseCIDR("2.176.0.0/12"), "mci"},
	{mustParseCIDR("37.255.0.0/16"), "mci"},
	// Irancell (MTN) — AS44244
	{mustParseCIDR("5.112.0.0/12"), "irancell"},
	{mustParseCIDR("100.64.0.0/15"), "irancell"},
	{mustParseCIDR("37.32.0.0/14"), "irancell"},
	// TCI / Mokhaberat — AS58224, AS12880
	{mustParseCIDR("2.144.0.0/12"), "mokhaberat"},
	{mustParseCIDR("5.200.0.0/16"), "mokhaberat"},
	{mustParseCIDR("78.38.0.0/15"), "mokhaberat"},
	{mustParseCIDR("85.185.0.0/16"), "mokhaberat"},
	// Shatel — AS31549
	{mustParseCIDR("2.180.0.0/14"), "shatel"},
	{mustParseCIDR("78.154.0.0/16"), "shatel"},
	// Asiatech — AS43754
	{mustParseCIDR("188.229.0.0/17"), "asiatech"},
	{mustParseCIDR("5.145.112.0/20"), "asiatech"},
	// Rightel — AS57218
	{mustParseCIDR("5.126.0.0/16"), "rightel"},
	{mustParseCIDR("5.127.0.0/16"), "rightel"},
}

// Close releases the database.
func (r *Resolver) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db != nil {
		err := r.db.Close()
		r.db = nil
		return err
	}
	return nil
}

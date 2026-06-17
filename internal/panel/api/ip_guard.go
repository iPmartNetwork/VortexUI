package api

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// IPGuard restricts panel access to a configurable set of IPs/CIDRs.
// When the whitelist is non-empty, only listed IPs can access the panel.
// When the blacklist is non-empty, listed IPs are blocked.
// Both can be active simultaneously (whitelist is checked first).
type IPGuard struct {
	mu        sync.RWMutex
	whitelist []*net.IPNet
	blacklist []*net.IPNet
	enabled   bool
}

// NewIPGuard builds a guard from comma-separated CIDR/IP strings.
func NewIPGuard(whitelist, blacklist string) *IPGuard {
	g := &IPGuard{}
	g.whitelist = parseCIDRs(whitelist)
	g.blacklist = parseCIDRs(blacklist)
	g.enabled = len(g.whitelist) > 0 || len(g.blacklist) > 0
	return g
}

// Middleware returns an Echo middleware that enforces the IP rules.
func (g *IPGuard) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !g.enabled {
				return next(c)
			}
			ip := net.ParseIP(c.RealIP())
			if ip == nil {
				return next(c) // can't parse → allow (fail open)
			}

			g.mu.RLock()
			defer g.mu.RUnlock()

			// Whitelist check: if set, IP must be in it.
			if len(g.whitelist) > 0 {
				allowed := false
				for _, cidr := range g.whitelist {
					if cidr.Contains(ip) {
						allowed = true
						break
					}
				}
				if !allowed {
					return echo.NewHTTPError(http.StatusForbidden, "IP not allowed")
				}
			}

			// Blacklist check: if IP is in blacklist, deny.
			for _, cidr := range g.blacklist {
				if cidr.Contains(ip) {
					return echo.NewHTTPError(http.StatusForbidden, "IP blocked")
				}
			}

			return next(c)
		}
	}
}

// SetWhitelist updates the whitelist at runtime (from Settings API).
func (g *IPGuard) SetWhitelist(cidrs string) {
	g.mu.Lock()
	g.whitelist = parseCIDRs(cidrs)
	g.enabled = len(g.whitelist) > 0 || len(g.blacklist) > 0
	g.mu.Unlock()
}

// SetBlacklist updates the blacklist at runtime.
func (g *IPGuard) SetBlacklist(cidrs string) {
	g.mu.Lock()
	g.blacklist = parseCIDRs(cidrs)
	g.enabled = len(g.whitelist) > 0 || len(g.blacklist) > 0
	g.mu.Unlock()
}

func parseCIDRs(s string) []*net.IPNet {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var nets []*net.IPNet
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// If no CIDR notation, treat as single IP.
		if !strings.Contains(entry, "/") {
			if net.ParseIP(entry) != nil {
				entry += "/32"
				if strings.Contains(entry, ":") {
					entry = strings.TrimSuffix(entry, "/32") + "/128"
				}
			}
		}
		_, cidr, err := net.ParseCIDR(entry)
		if err == nil {
			nets = append(nets, cidr)
		}
	}
	return nets
}

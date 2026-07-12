package core

import (
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// ProbeBlockRules routes traffic from blocked probe IPs to the "blocked" outbound.
func ProbeBlockRules(nodeID uuid.UUID, blockedIPs []string, inboundTags []string) []domain.RoutingRule {
	if len(blockedIPs) == 0 {
		return nil
	}
	return []domain.RoutingRule{
		{
			NodeID:      nodeID,
			Name:        "probe-block",
			Priority:    -2,
			InboundTags: inboundTags,
			IP:          blockedIPs,
			OutboundTag: "blocked",
			Enabled:     true,
		},
	}
}

// SNIRouteRules converts SNI route policies into engine-neutral routing rules.
func SNIRouteRules(nodeID uuid.UUID, inboundTag string, routes []*domain.SNIRoute) []domain.RoutingRule {
	var rules []domain.RoutingRule
	for _, r := range routes {
		if r == nil || !r.Enabled || r.SNI == "" {
			continue
		}
		rule := domain.RoutingRule{
			NodeID:      nodeID,
			Name:        "sni-" + r.SNI,
			Priority:    r.Priority,
			InboundTags: []string{inboundTag},
			Domains:     []string{r.SNI},
			Enabled:     true,
		}
		switch r.Action {
		case "reject", "decoy":
			rule.OutboundTag = "blocked"
		case "proxy":
			if r.TargetTag == "" {
				continue
			}
			rule.OutboundTag = r.TargetTag
		default:
			continue
		}
		rules = append(rules, rule)
	}
	return rules
}

// ApplyManagedTLS injects operator-managed certificates and server names into a
// TLS inbound before sync. Existing Raw["tls"] with certificate material wins.
func ApplyManagedTLS(in *domain.Inbound, domains []*domain.SNIDomain, certForDomain func(string) (*domain.SSLCertificate, error)) {
	if in == nil || in.Security != domain.SecurityTLS || certForDomain == nil {
		return
	}
	if in.Raw == nil {
		in.Raw = map[string]any{}
	}
	if hasManagedCert(in.Raw["tls"]) {
		mergeSNIDomains(in, domains)
		return
	}

	var certPEM, keyPEM string
	names := make([]string, 0, len(domains))
	for _, d := range domains {
		if d == nil || d.Domain == "" {
			continue
		}
		names = append(names, d.Domain)
		if certPEM != "" {
			continue
		}
		cert, err := certForDomain(d.Domain)
		if err != nil || cert == nil || cert.Status != domain.CertActive {
			continue
		}
		if cert.CertPEM == "" || cert.KeyPEM == "" {
			continue
		}
		certPEM, keyPEM = cert.CertPEM, cert.KeyPEM
	}
	if certPEM != "" {
		in.Raw["tls"] = map[string]any{"certificate": certPEM, "key": keyPEM}
	}
	appendSNINames(in, names...)
}

func mergeSNIDomains(in *domain.Inbound, domains []*domain.SNIDomain) {
	names := make([]string, 0, len(domains))
	for _, d := range domains {
		if d != nil && d.Domain != "" {
			names = append(names, d.Domain)
		}
	}
	appendSNINames(in, names...)
}

func appendSNINames(in *domain.Inbound, names ...string) {
	if len(names) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(in.SNI)+len(names))
	for _, s := range in.SNI {
		if s != "" {
			seen[s] = struct{}{}
		}
	}
	for _, s := range names {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		in.SNI = append(in.SNI, s)
	}
}

func hasManagedCert(v any) bool {
	t, ok := v.(map[string]any)
	if !ok {
		return false
	}
	cert, _ := t["certificate"].(string)
	return cert != ""
}

// ApplyDecoy stores decoy metadata on the inbound so renderers can wire fallbacks.
func ApplyDecoy(in *domain.Inbound, decoy *domain.DecoySite) {
	if in == nil || decoy == nil || !decoy.Enabled {
		return
	}
	if in.Raw == nil {
		in.Raw = map[string]any{}
	}
	in.Raw["decoy"] = map[string]any{
		"mode":        string(decoy.Mode),
		"target_url":  decoy.TargetURL,
		"static_html": decoy.StaticHTML,
	}
}

// DecoyFallbackDest returns an xray fallback destination from inbound Raw decoy config.
func DecoyFallbackDest(in domain.Inbound) string {
	raw, ok := in.Raw["decoy"].(map[string]any)
	if !ok {
		return ""
	}
	mode, _ := raw["mode"].(string)
	switch domain.DecoyMode(mode) {
	case domain.DecoyProxy:
		return normalizeDecoyDest(raw["target_url"])
	default:
		return ""
	}
}

func normalizeDecoyDest(v any) string {
	s, _ := v.(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Port() != "" {
		return u.Host
	}
	if u.Scheme == "http" {
		return u.Hostname() + ":80"
	}
	return u.Hostname() + ":443"
}

// ActiveBlockedIPs drops expired probe blocks.
func ActiveBlockedIPs(blocked []domain.BlockedIP, now time.Time) []string {
	out := make([]string, 0, len(blocked))
	for _, b := range blocked {
		if b.IP == "" {
			continue
		}
		if !b.ExpiresAt.IsZero() && !b.ExpiresAt.After(now) {
			continue
		}
		out = append(out, b.IP)
	}
	return out
}

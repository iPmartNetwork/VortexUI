package subscription

import (
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
)

// This file maps engine-neutral domain.RoutingRule values onto the rule shapes
// the Clash and sing-box clients understand. The built-in routing packs use a
// small token vocabulary — plain domains, `geosite:<tag>`, `geoip:<cc>`, raw
// CIDRs and (optionally) `keyword:<kw>` / `domain:<suffix>` prefixes — which map
// cleanly onto both engines. Mapping is pure and side-effect free so it can be
// unit-tested in isolation.

// domTokenKind classifies a domain matcher token after its prefix is read.
type domTokenKind int

const (
	tokSuffix  domTokenKind = iota // bare domain or "domain:" → suffix match
	tokGeosite                     // "geosite:<tag>"
	tokKeyword                     // "keyword:<kw>"
)

// domainToken splits a domain matcher into its kind and bare value.
func domainToken(d string) (domTokenKind, string) {
	switch {
	case strings.HasPrefix(d, "geosite:"):
		return tokGeosite, strings.TrimPrefix(d, "geosite:")
	case strings.HasPrefix(d, "keyword:"):
		return tokKeyword, strings.TrimPrefix(d, "keyword:")
	case strings.HasPrefix(d, "domain:"):
		return tokSuffix, strings.TrimPrefix(d, "domain:")
	default:
		return tokSuffix, d
	}
}

// geoipToken reports whether an IP matcher is a `geoip:<cc>` token and returns
// the country code; otherwise it is treated as a CIDR/IP literal.
func geoipToken(ip string) (string, bool) {
	if strings.HasPrefix(ip, "geoip:") {
		return strings.TrimPrefix(ip, "geoip:"), true
	}
	return "", false
}

// normalizeCIDR ensures a bare IP literal carries an explicit prefix length so
// it is a valid CIDR for the clients (/32 for IPv4, /128 for IPv6).
func normalizeCIDR(ip string) string {
	if strings.Contains(ip, "/") {
		return ip
	}
	if strings.Contains(ip, ":") {
		return ip + "/128"
	}
	return ip + "/32"
}

// clashTarget resolves a rule's neutral OutboundTag to a Clash rule target.
// direct → DIRECT, block/blocked/reject → REJECT, anything else routes through
// the proxy selector group (reusing the name renderClash already emits).
func clashTarget(tag, proxyGroup string) string {
	switch strings.ToLower(tag) {
	case "direct":
		return "DIRECT"
	case "block", "blocked", "reject":
		return "REJECT"
	default:
		return proxyGroup
	}
}

// singboxTarget resolves a rule's neutral OutboundTag to a sing-box outbound tag.
// direct → "direct" (always present), block/blocked/reject → "block" (appended
// by renderSingbox when needed), anything else → the selector tag.
func singboxTarget(tag, selector string) string {
	switch strings.ToLower(tag) {
	case "direct":
		return "direct"
	case "block", "blocked", "reject":
		return "block"
	default:
		return selector
	}
}

// isBlockTag reports whether a neutral OutboundTag rejects traffic.
func isBlockTag(tag string) bool {
	switch strings.ToLower(tag) {
	case "block", "blocked", "reject":
		return true
	default:
		return false
	}
}

// clashRules translates the pack's rules into Clash rule strings. proxyGroup is
// the selector group non-direct/non-block targets resolve to. Each matcher value
// becomes its own rule line (so domains and IPs combine as OR, matching Clash
// semantics). The caller appends the MATCH fallback.
func clashRules(rules []domain.RoutingRule, proxyGroup string) []string {
	var out []string
	for _, r := range rules {
		target := clashTarget(r.OutboundTag, proxyGroup)
		for _, d := range r.Domains {
			kind, val := domainToken(d)
			if val == "" {
				continue
			}
			switch kind {
			case tokGeosite:
				out = append(out, "GEOSITE,"+val+","+target)
			case tokKeyword:
				out = append(out, "DOMAIN-KEYWORD,"+val+","+target)
			default:
				out = append(out, "DOMAIN-SUFFIX,"+val+","+target)
			}
		}
		for _, ip := range r.IP {
			if ip == "" {
				continue
			}
			if cc, ok := geoipToken(ip); ok {
				out = append(out, "GEOIP,"+cc+","+target)
			} else {
				out = append(out, "IP-CIDR,"+normalizeCIDR(ip)+","+target)
			}
		}
	}
	return out
}

// singboxRules translates the pack's rules into sing-box route.rule objects.
// selector is the selector tag non-direct/non-block targets resolve to. One
// object is emitted per matcher type per rule (so types combine as OR, matching
// the per-line Clash mapping). Each object lists its matcher values plus the
// resolved outbound.
func singboxRules(rules []domain.RoutingRule, selector string) []map[string]any {
	var out []map[string]any
	for _, r := range rules {
		outbound := singboxTarget(r.OutboundTag, selector)

		var suffixes, geosites, keywords []string
		for _, d := range r.Domains {
			kind, val := domainToken(d)
			if val == "" {
				continue
			}
			switch kind {
			case tokGeosite:
				geosites = append(geosites, val)
			case tokKeyword:
				keywords = append(keywords, val)
			default:
				suffixes = append(suffixes, val)
			}
		}
		var geoips, cidrs []string
		for _, ip := range r.IP {
			if ip == "" {
				continue
			}
			if cc, ok := geoipToken(ip); ok {
				geoips = append(geoips, cc)
			} else {
				cidrs = append(cidrs, normalizeCIDR(ip))
			}
		}

		if len(suffixes) > 0 {
			out = append(out, map[string]any{"domain_suffix": suffixes, "outbound": outbound})
		}
		if len(keywords) > 0 {
			out = append(out, map[string]any{"domain_keyword": keywords, "outbound": outbound})
		}
		if len(geosites) > 0 {
			out = append(out, map[string]any{"geosite": geosites, "outbound": outbound})
		}
		if len(geoips) > 0 {
			out = append(out, map[string]any{"geoip": geoips, "outbound": outbound})
		}
		if len(cidrs) > 0 {
			out = append(out, map[string]any{"ip_cidr": cidrs, "outbound": outbound})
		}
	}
	return out
}

// singboxNeedsBlock reports whether any rule rejects traffic, so renderSingbox
// can append a block outbound the rules can reference.
func singboxNeedsBlock(rules []domain.RoutingRule) bool {
	for _, r := range rules {
		if isBlockTag(r.OutboundTag) {
			return true
		}
	}
	return false
}

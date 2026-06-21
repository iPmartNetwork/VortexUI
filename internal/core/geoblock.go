// Package core — geo-blocking routing rule generation.
package core

import "github.com/vortexui/vortexui/internal/domain"

// GeoBlockRules generates routing rules that enforce an inbound's GeoPolicy.
// For Xray: creates rules that route non-allowed traffic to "blocked" outbound.
// For sing-box: creates route rules with geoip source matching.
//
// Strategy:
// - AllowedCountries set: all traffic NOT from those countries → block
// - BlockedCountries set: traffic FROM those countries → block
func GeoBlockRules(inbound domain.Inbound) []domain.RoutingRule {
	if inbound.GeoPolicy == nil {
		return nil
	}
	gp := inbound.GeoPolicy
	var rules []domain.RoutingRule

	if len(gp.AllowedCountries) > 0 {
		// Block everything NOT from allowed countries, expressed positively as two
		// rules evaluated in priority order:
		//   1. allowed source geoip → "direct" (normal egress), highest priority
		//   2. catch-all for this inbound → "blocked"
		// We deliberately route the allow rule to the well-known "direct" outbound
		// rather than leaving OutboundTag empty: an empty-target rule is dropped by
		// both renderers (xray omits the outboundTag and the rule becomes a no-op;
		// sing-box's buildRoute skips a rule with no resolvable target), which would
		// let the catch-all block rule reject ALL traffic — including the allowed
		// countries. "direct" is guaranteed to exist in both cores' outbound sets.
		geoIPs := make([]string, len(gp.AllowedCountries))
		for i, c := range gp.AllowedCountries {
			geoIPs[i] = "geoip:" + c
		}
		rules = append(rules, domain.RoutingRule{
			Name:        "geo-allow-" + inbound.Tag,
			InboundTags: []string{inbound.Tag},
			IP:          geoIPs,   // source IPs that ARE allowed
			OutboundTag: "direct", // route allowed traffic to normal egress (must be honored, not dropped)
			Enabled:     true,
			Priority:    -1, // highest priority
		})
		// Block everything else on this inbound
		rules = append(rules, domain.RoutingRule{
			Name:        "geo-block-rest-" + inbound.Tag,
			InboundTags: []string{inbound.Tag},
			OutboundTag: "blocked",
			Enabled:     true,
			Priority:    0,
		})
	}

	if len(gp.BlockedCountries) > 0 {
		// Block traffic FROM these countries
		geoIPs := make([]string, len(gp.BlockedCountries))
		for i, c := range gp.BlockedCountries {
			geoIPs[i] = "geoip:" + c
		}
		rules = append(rules, domain.RoutingRule{
			Name:        "geo-block-" + inbound.Tag,
			InboundTags: []string{inbound.Tag},
			IP:          geoIPs,
			OutboundTag: "blocked",
			Enabled:     true,
			Priority:    -1,
		})
	}

	return rules
}

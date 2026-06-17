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
		// Block everything NOT from allowed countries.
		// Xray approach: route inbound traffic where source geoip is NOT in the
		// allowed list to "blocked". We express this as: for each allowed country,
		// add a pass-through rule; then a catch-all block for this inbound.
		// Actually simpler: negate in one rule with source IP "geoip:!XX" is not
		// natively supported, so we use the positive approach — route traffic from
		// allowed countries to their normal outbound, then block the rest.

		// Rule: traffic from allowed countries on this inbound → pass (no action needed,
		// it falls through to normal routing). So we only need the block rule:
		// "any traffic on this inbound NOT from allowed geoip → blocked"
		geoIPs := make([]string, len(gp.AllowedCountries))
		for i, c := range gp.AllowedCountries {
			geoIPs[i] = "geoip:" + c
		}
		rules = append(rules, domain.RoutingRule{
			Name:        "geo-allow-" + inbound.Tag,
			InboundTags: []string{inbound.Tag},
			IP:          geoIPs, // source IPs that ARE allowed
			OutboundTag: "", // empty = pass through (don't block these)
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

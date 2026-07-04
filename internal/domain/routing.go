package domain

import "github.com/google/uuid"

// RoutingRule is one engine-neutral routing decision on a node: when the
// matchers apply, traffic is sent to OutboundTag or balanced across BalancerTag.
// Rules are evaluated in ascending Priority order; the first match on the engine
// wins, matching Xray/sing-box routing semantics.
type RoutingRule struct {
	ID       uuid.UUID `json:"id"`
	NodeID   uuid.UUID `json:"node_id"`
	Priority int       `json:"priority"` // lower runs first
	Name     string    `json:"name,omitempty"`

	// Matchers. Within a rule they combine with AND; empty matchers are ignored.
	InboundTags []string `json:"inbound_tags,omitempty"`
	Domains     []string `json:"domains,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`      // "443", "1000-2000", or "80,443"
	Protocols   []string `json:"protocols,omitempty"` // http, tls, quic, bittorrent
	Network     string   `json:"network,omitempty"`   // tcp, udp, "tcp,udp"

	// Target. Exactly one of OutboundTag / BalancerTag must be set.
	OutboundTag string `json:"outbound_tag,omitempty"`
	BalancerTag string `json:"balancer_tag,omitempty"`

	Enabled bool `json:"enabled"`
}

// HasMatcher reports whether the rule constrains anything. A rule with no
// matcher would unconditionally capture all traffic, which is almost always a
// mistake, so the service layer rejects it.
func (r *RoutingRule) HasMatcher() bool {
	return len(r.InboundTags) > 0 || len(r.Domains) > 0 || len(r.IP) > 0 ||
		r.Port != "" || len(r.Protocols) > 0 || r.Network != ""
}

// Validate checks the rule targets exactly one destination and matches something.
func (r *RoutingRule) Validate() error {
	switch {
	case r.OutboundTag == "" && r.BalancerTag == "":
		return errInvalid("routing rule needs an outbound_tag or balancer_tag")
	case r.OutboundTag != "" && r.BalancerTag != "":
		return errInvalid("routing rule cannot target both an outbound and a balancer")
	case !r.HasMatcher():
		return errInvalid("routing rule needs at least one matcher")
	}
	return nil
}

// BalancerStrategy selects how a balancer picks among its member outbounds.
type BalancerStrategy string

const (
	BalancerRandom     BalancerStrategy = "random"
	BalancerRoundRobin BalancerStrategy = "roundRobin"
	BalancerLeastPing  BalancerStrategy = "leastPing"
	BalancerLeastLoad  BalancerStrategy = "leastLoad"
)

// Valid reports whether the strategy is supported.
func (s BalancerStrategy) Valid() bool {
	switch s {
	case BalancerRandom, BalancerRoundRobin, BalancerLeastPing, BalancerLeastLoad:
		return true
	default:
		return false
	}
}

// NeedsObservatory reports whether the strategy depends on health probing. Ping/
// load based selection is meaningless without latency measurements, so the
// builder must emit an observatory when one of these is used.
func (s BalancerStrategy) NeedsObservatory() bool {
	return s == BalancerLeastPing || s == BalancerLeastLoad
}

// Balancer distributes traffic across a set of outbounds selected by tag prefix.
// When the strategy needs latency data (or Observe is set), the builder wires an
// observatory that probes each member and feeds the selection.
type Balancer struct {
	ID        uuid.UUID        `json:"id"`
	NodeID    uuid.UUID        `json:"node_id"`
	Tag       string           `json:"tag"`       // referenced by routing rules' balancer_tag
	Selectors []string         `json:"selectors"` // outbound tag prefixes to include
	Strategy  BalancerStrategy `json:"strategy"`

	// Observatory (health probing). Implied when Strategy.NeedsObservatory().
	Observe       bool   `json:"observe"`
	ProbeURL      string `json:"probe_url,omitempty"`      // default https://www.google.com/generate_204
	ProbeInterval string `json:"probe_interval,omitempty"` // default 10s

	Enabled bool `json:"enabled"`
}

// BalancerListItem is a balancer row with its parent node name for fleet views.
type BalancerListItem struct {
	Balancer
	NodeName string `json:"node_name"`
}

func (b *Balancer) WantsObservatory() bool {
	return b.Observe || b.Strategy.NeedsObservatory()
}

// Validate checks the balancer has a tag, at least one selector, and a known
// strategy.
func (b *Balancer) Validate() error {
	if b.Tag == "" {
		return errInvalid("balancer tag is required")
	}
	if len(b.Selectors) == 0 {
		return errInvalid("balancer %q needs at least one selector", b.Tag)
	}
	if b.Strategy == "" {
		b.Strategy = BalancerRandom
	}
	if !b.Strategy.Valid() {
		return errInvalid("unknown balancer strategy %q", string(b.Strategy))
	}
	return nil
}

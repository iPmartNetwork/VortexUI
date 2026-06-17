package domain

import "github.com/google/uuid"

// RoutingTemplate is a pre-built set of routing rules that can be applied to a
// node with one click. Examples: "Iran Direct", "Block Ads", "OpenAI via WARP".
type RoutingTemplate struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"` // "regional", "security", "optimization"
	Rules       []RoutingRule `json:"rules"`
	Outbounds   []Outbound    `json:"outbounds,omitempty"` // outbounds needed by the rules
}

// BuiltinRoutingTemplates returns the default templates.
func BuiltinRoutingTemplates() []RoutingTemplate {
	return []RoutingTemplate{
		{
			Name:        "Iran Direct",
			Description: "Route Iranian domains/IPs directly, everything else via proxy",
			Category:    "regional",
			Rules: []RoutingRule{
				{Name: "iran-domains-direct", Domains: []string{"geosite:category-ir"}, OutboundTag: "direct", Priority: 1, Enabled: true},
				{Name: "iran-ips-direct", IP: []string{"geoip:ir"}, OutboundTag: "direct", Priority: 2, Enabled: true},
			},
		},
		{
			Name:        "Block Ads & Malware",
			Description: "Block advertising and malicious domains",
			Category:    "security",
			Rules: []RoutingRule{
				{Name: "block-ads", Domains: []string{"geosite:category-ads-all"}, OutboundTag: "blocked", Priority: 1, Enabled: true},
				{Name: "block-malware", Domains: []string{"geosite:malware", "geosite:phishing"}, OutboundTag: "blocked", Priority: 2, Enabled: true},
			},
		},
		{
			Name:        "OpenAI via WARP",
			Description: "Route OpenAI/ChatGPT traffic through Cloudflare WARP for clean IP",
			Category:    "optimization",
			Rules: []RoutingRule{
				{Name: "openai-warp", Domains: []string{"openai.com", "chat.openai.com", "api.openai.com", "chatgpt.com"}, OutboundTag: "warp", Priority: 1, Enabled: true},
			},
		},
		{
			Name:        "China Bypass (GFW)",
			Description: "Direct for Chinese sites, proxy for international",
			Category:    "regional",
			Rules: []RoutingRule{
				{Name: "china-direct", Domains: []string{"geosite:cn"}, OutboundTag: "direct", Priority: 1, Enabled: true},
				{Name: "china-ips-direct", IP: []string{"geoip:cn"}, OutboundTag: "direct", Priority: 2, Enabled: true},
			},
		},
		{
			Name:        "Gaming Optimization",
			Description: "Direct route for game servers (lower latency)",
			Category:    "optimization",
			Rules: []RoutingRule{
				{Name: "gaming-direct", Domains: []string{"geosite:steam", "geosite:epicgames", "geosite:xbox"}, OutboundTag: "direct", Priority: 1, Enabled: true},
			},
		},
	}
}

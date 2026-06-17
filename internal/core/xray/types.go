package xray

import "encoding/json"

// These types mirror the subset of Xray's configuration schema VortexUI emits.
// Keeping them typed (rather than building maps inline) makes the builder output
// stable and lets the compiler catch field renames.

type xrayConfig struct {
	Log         logConf          `json:"log"`
	API         apiConf          `json:"api"`
	Stats       struct{}         `json:"stats"`
	Policy      policyConf       `json:"policy"`
	Inbounds    []inbound        `json:"inbounds"`
	Outbounds   []outbound       `json:"outbounds"`
	Routing     routingConf      `json:"routing"`
	Observatory *observatoryConf `json:"observatory,omitempty"`
}

type logConf struct {
	Loglevel string `json:"loglevel"`
}

type apiConf struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services"`
}

type policyConf struct {
	Levels map[string]policyLevel `json:"levels"`
	System systemPolicy           `json:"system"`
}

type policyLevel struct {
	StatsUserUplink   bool  `json:"statsUserUplink"`
	StatsUserDownlink bool  `json:"statsUserDownlink"`
	BufferSize        int32 `json:"bufferSize,omitempty"`        // KB; controls throughput
	Uplinkonly        int32 `json:"uplinkOnly,omitempty"`        // seconds
	Downlinkonly      int32 `json:"downlinkOnly,omitempty"`      // seconds
}

type systemPolicy struct {
	StatsInboundUplink   bool `json:"statsInboundUplink"`
	StatsInboundDownlink bool `json:"statsInboundDownlink"`
}

type inbound struct {
	Tag            string          `json:"tag"`
	Listen         string          `json:"listen,omitempty"`
	Port           int             `json:"port"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings"`
	StreamSettings json.RawMessage `json:"streamSettings,omitempty"`
}

type outbound struct {
	Protocol       string          `json:"protocol"`
	Tag            string          `json:"tag"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	StreamSettings json.RawMessage `json:"streamSettings,omitempty"`
}

type routingConf struct {
	DomainStrategy string         `json:"domainStrategy,omitempty"`
	Balancers      []balancerConf `json:"balancers,omitempty"`
	Rules          []routingRule  `json:"rules"`
}

type balancerConf struct {
	Tag      string           `json:"tag"`
	Selector []string         `json:"selector"`
	Strategy balancerStrategy `json:"strategy,omitempty"`
}

type balancerStrategy struct {
	Type string `json:"type"`
}

type observatoryConf struct {
	SubjectSelector []string `json:"subjectSelector"`
	ProbeURL        string   `json:"probeUrl,omitempty"`
	ProbeInterval   string   `json:"probeInterval,omitempty"`
}

type routingRule struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	Network     string   `json:"network,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
}

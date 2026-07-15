package domain

import (
	"time"

	"github.com/google/uuid"
)

// CoreType selects which proxy engine an inbound/node runs on. Supporting both
// under one abstraction is a core differentiator of VortexUI.
type CoreType string

const (
	CoreXray    CoreType = "xray"
	CoreSingbox CoreType = "singbox"
	// CoreMulti marks a node agent running several engines via CompositeDriver.
	CoreMulti CoreType = "multi"
)

// NodeStatus reflects the live connectivity of a remote node agent.
type NodeStatus string

const (
	NodeConnected    NodeStatus = "connected"
	NodeDisconnected NodeStatus = "disconnected"
	NodeError        NodeStatus = "error"
	NodeDisabled     NodeStatus = "disabled"
)

// NodeDiagCode classifies why a node is unreachable or unhealthy for the UI.
type NodeDiagCode string

const (
	NodeDiagOK          NodeDiagCode = "ok"
	NodeDiagUnreachable NodeDiagCode = "unreachable"
	NodeDiagMTLS        NodeDiagCode = "mtls_fail"
	NodeDiagCoreDown    NodeDiagCode = "core_down"
	NodeDiagUnknown     NodeDiagCode = "unknown"
)

// NodeDiagnostics is the live connectivity snapshot surfaced by the hub/API.
type NodeDiagnostics struct {
	Code             NodeDiagCode `json:"code"`
	Message          string       `json:"message,omitempty"`
	NetworkReachable bool         `json:"network_reachable,omitempty"`
	CAMatch          bool         `json:"ca_match,omitempty"`
	CheckedAt        *time.Time   `json:"checked_at,omitempty"`
}

// NodeEnrollmentPhase tracks remote node onboarding progress in the UI.
type NodeEnrollmentPhase string

const (
	NodePhasePending   NodeEnrollmentPhase = "pending"
	NodePhaseConnected NodeEnrollmentPhase = "connected"
	NodePhaseSynced    NodeEnrollmentPhase = "synced"
)

// Node is a remote server running a CoreDriver, managed over gRPC + mTLS.
type Node struct {
	ID      uuid.UUID  `json:"id"`
	Name    string     `json:"name"`
	Address string     `json:"address"` // host:port of the node agent
	Core    CoreType   `json:"core"`    // default core for inbounds on this node
	// EnabledCores lists proxy engines active on this node. When more than one is
	// set the agent runs a CompositeDriver and inbounds may override via Core.
	EnabledCores []CoreType `json:"enabled_cores,omitempty"`
	Status  NodeStatus `json:"status"`

	// UsageRatio lets you weight a node when distributing/auto-failing-over users.
	UsageRatio float64 `json:"usage_ratio"`

	// Endpoint overrides the public host/IP used in subscription links. When set,
	// clients receive this address instead of the node's real IP — essential for
	// tunnel/CDN/relay setups where the user should connect via an intermediate.
	Endpoint string `json:"endpoint,omitempty"`

	// Region is a display label for dashboards (e.g. "Frankfurt, DE"). When
	// LocationAuto is true and Region is empty, it is derived from GeoIP.
	Region string `json:"region,omitempty"`
	// CountryCode is ISO 3166-1 alpha-2, resolved from the public endpoint host.
	CountryCode string `json:"country_code,omitempty"`
	// PingMs is the last gRPC health-check round-trip in milliseconds.
	PingMs int `json:"ping_ms,omitempty"`
	// LocationAuto when true keeps Region/CountryCode in sync with endpoint GeoIP.
	LocationAuto bool `json:"location_auto"`

	// Live health, refreshed by the hub from the agent's heartbeat.
	LastSeen  *time.Time `json:"last_seen,omitempty"`
	Health    NodeHealth `json:"health"`
	CoreVer   string     `json:"core_version,omitempty"`
	AgentVer  string     `json:"agent_version,omitempty"`
	// Diagnostics explains disconnects (mTLS mismatch, unreachable agent, core down).
	Diagnostics     *NodeDiagnostics     `json:"diagnostics,omitempty"`
	EnrollmentPhase NodeEnrollmentPhase  `json:"enrollment_phase,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
}

// NodeListItem is a node row enriched for the fleet management table.
type NodeListItem struct {
	Node
	Location   string `json:"location,omitempty"`
	UsersCount int    `json:"users_count"`
}

// NodeHealth is the resource snapshot pushed by an agent heartbeat.
type NodeHealth struct {
	CPUPercent   float64 `json:"cpu_percent"`
	MemPercent   float64 `json:"mem_percent"`
	DiskPercent  float64 `json:"disk_percent"`
	CoreRunning  bool    `json:"core_running"`
	Connections  int     `json:"connections"`
	CoreVersion  string  `json:"core_version,omitempty"`
	AgentVersion string  `json:"agent_version,omitempty"`
	PingMs       int     `json:"ping_ms,omitempty"`
}

// IsHealthy is the gate used by failover: an unhealthy node should not receive users.
func (n *Node) IsHealthy() bool {
	return n.Status == NodeConnected && n.Health.CoreRunning
}

// Live decides, from persisted health alone (no live hub access), whether a node
// should still be advertised in subscriptions. The rule prunes only nodes we
// have positive evidence are down — a never-polled node (LastSeen nil) is given
// the benefit of the doubt so a freshly started panel doesn't hand out empty
// configs before its first health sweep.
func (n *Node) Live(now time.Time, staleAfter time.Duration) bool {
	if n.LastSeen == nil {
		return true
	}
	if !n.Health.CoreRunning {
		return false
	}
	return now.Sub(*n.LastSeen) <= staleAfter
}

// NormalizedEnabledCores returns the active engines for this node, falling back
// to Core (then xray) when enabled_cores is unset in older records.
func (n *Node) NormalizedEnabledCores() []CoreType {
	if len(n.EnabledCores) > 0 {
		return n.EnabledCores
	}
	if n.Core != "" {
		return []CoreType{n.Core}
	}
	return []CoreType{CoreXray}
}

// IsMultiCore reports whether the node runs more than one proxy engine.
func (n *Node) IsMultiCore() bool {
	return len(n.NormalizedEnabledCores()) > 1
}

// SyncCoreType is the core value sent over gRPC during full sync.
func (n *Node) SyncCoreType() CoreType {
	if n.IsMultiCore() {
		return CoreMulti
	}
	return n.NormalizedEnabledCores()[0]
}

// CoreEnabled reports whether ct is listed in the node's enabled engines.
func (n *Node) CoreEnabled(ct CoreType) bool {
	for _, c := range n.NormalizedEnabledCores() {
		if c == ct {
			return true
		}
	}
	return false
}

// ResolveInboundCore picks the engine for an inbound: explicit override, else
// the node's default Core, else xray.
func (n *Node) ResolveInboundCore(override CoreType) CoreType {
	if override != "" {
		return override
	}
	if n.Core != "" {
		return n.Core
	}
	return CoreXray
}

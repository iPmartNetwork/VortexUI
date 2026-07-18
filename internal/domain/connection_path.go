package domain

import "github.com/google/uuid"

// ConnectionHop represents one hop in a multi-hop proxy chain.
type ConnectionHop struct {
	Type     string `json:"type"`     // "client", "cdn", "relay", "node", "target"
	Label    string `json:"label"`    // human-readable label
	IP       string `json:"ip,omitempty"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Latency  int    `json:"latency_ms,omitempty"`
}

// ConnectionPath describes the full path a user's traffic takes.
type ConnectionPath struct {
	UserID uuid.UUID       `json:"user_id"`
	NodeID uuid.UUID       `json:"node_id"`
	Hops   []ConnectionHop `json:"hops"`
}

// BuildConnectionPath constructs the connection path for visualization.
func BuildConnectionPath(username string, nodeEndpoint string, protocol string, transport string, hasCDN bool) ConnectionPath {
	var hops []ConnectionHop
	hops = append(hops, ConnectionHop{Type: "client", Label: username})
	if hasCDN {
		hops = append(hops, ConnectionHop{Type: "cdn", Label: "CDN (Cloudflare)", Protocol: transport})
	}
	hops = append(hops, ConnectionHop{Type: "node", Label: nodeEndpoint, Protocol: protocol + "/" + transport})
	hops = append(hops, ConnectionHop{Type: "target", Label: "Internet"})
	return ConnectionPath{Hops: hops}
}

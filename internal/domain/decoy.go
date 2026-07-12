package domain

import (
	"time"

	"github.com/google/uuid"
)

// DecoyMode determines how the decoy site is served.
type DecoyMode string

const (
	DecoyProxy  DecoyMode = "proxy"  // reverse-proxy to a real website
	DecoyStatic DecoyMode = "static" // serve uploaded static HTML
)

// DefaultDecoyListen is the loopback address xray fallbacks use for static/honeypot pages.
const DefaultDecoyListen = "127.0.0.1:18080"

// DecoySite configures the decoy/fallback website served when invalid
// connections probe the server (active probing protection).
type DecoySite struct {
	ID         uuid.UUID  `json:"id"`
	NodeID     *uuid.UUID `json:"node_id,omitempty"` // nil = global default
	Mode       DecoyMode  `json:"mode"`
	TargetURL  string     `json:"target_url,omitempty"`  // for proxy mode
	StaticHTML string     `json:"static_html,omitempty"` // for static mode
	Enabled    bool       `json:"enabled"`
	CreatedAt  time.Time  `json:"created_at"`
}

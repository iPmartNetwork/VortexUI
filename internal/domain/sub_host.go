package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// HostSecurity is the TLS-layer override a subscription host applies on top of
// its inbound. "inbound_default" leaves the inbound's own security untouched;
// the others force the emitted proxy to use the named security.
type HostSecurity string

const (
	// HostSecurityInboundDefault inherits the bound inbound's security.
	HostSecurityInboundDefault HostSecurity = "inbound_default"
	// HostSecurityNone forces plaintext (no TLS) for the emitted proxy.
	HostSecurityNone HostSecurity = "none"
	// HostSecurityTLS forces standard TLS for the emitted proxy.
	HostSecurityTLS HostSecurity = "tls"
	// HostSecurityReality forces REALITY for the emitted proxy.
	HostSecurityReality HostSecurity = "reality"
)

// SubHost is a Marzban-style per-inbound host override. At subscription-render
// time an enabled host is projected into one share link, overlaying the
// inbound's own endpoint fields (address/port/sni/host/path/...) and supporting
// template variables in Remark and Address. Hosts are ordered by Priority.
type SubHost struct {
	ID            uuid.UUID    `json:"id"`
	InboundID     uuid.UUID    `json:"inbound_id"`
	Remark        string       `json:"remark"`         // supports template vars
	Address       string       `json:"address"`        // supports template vars; "" = inbound host
	Port          *int         `json:"port,omitempty"` // nil = inbound port
	SNI           string       `json:"sni,omitempty"`
	HostHeader    string       `json:"host,omitempty"`
	Path          string       `json:"path,omitempty"`
	ALPN          string       `json:"alpn,omitempty"`        // "", "h2", "h3,h2", ...
	Fingerprint   string       `json:"fingerprint,omitempty"` // uTLS
	Security      HostSecurity `json:"security"`              // default inbound_default
	AllowInsecure bool         `json:"allow_insecure"`
	MuxEnable     bool         `json:"mux_enable"`
	Fragment      string       `json:"fragment,omitempty"` // "length,interval,packet"
	Priority      int          `json:"priority"`           // ascending sort
	Enabled       bool         `json:"enabled"`
	CreatedAt     time.Time    `json:"created_at"`
}

// fragmentPattern matches a Marzban-style fragment setting: three comma-joined
// fields whose values are positive integers or hyphenated integer ranges
// (e.g. "100-200,10-20,tlshello" is rejected — packets must be numeric here as
// "length,interval,packet" counts). A leading/trailing space is tolerated.
var fragmentPattern = regexp.MustCompile(`^\d+(-\d+)?,\d+(-\d+)?,\d+(-\d+)?$`)

// ValidHostSecurity reports whether s is one of the four allowed values.
func ValidHostSecurity(s HostSecurity) bool {
	switch s {
	case HostSecurityInboundDefault, HostSecurityNone, HostSecurityTLS, HostSecurityReality:
		return true
	default:
		return false
	}
}

// Validate checks the host's invariants: a non-empty remark, a valid security
// enum, and a fragment that is either empty or "n,n,n"-shaped.
func (h *SubHost) Validate() error {
	if strings.TrimSpace(h.Remark) == "" {
		return errors.New("remark is required")
	}
	if !ValidHostSecurity(h.Security) {
		return errors.New("invalid security (want inbound_default|none|tls|reality)")
	}
	if f := strings.TrimSpace(h.Fragment); f != "" && !fragmentPattern.MatchString(f) {
		return errors.New("invalid fragment (want \"length,interval,packet\")")
	}
	return nil
}

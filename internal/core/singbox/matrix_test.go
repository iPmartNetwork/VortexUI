package singbox

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// Fake PEM material — the renderer only splits these into line arrays, never
// parses them, so bogus PEM is enough to satisfy a TLS inbound.
const (
	fakeCertPEM = "-----BEGIN CERTIFICATE-----\nMIIBmatrixfake\n-----END CERTIFICATE-----"
	fakeKeyPEM  = "-----BEGIN PRIVATE KEY-----\nMIIBmatrixfake\n-----END PRIVATE KEY-----"
)

// matrixUser returns a user with every per-protocol credential populated so it
// is renderable on any protocol the matrix throws at it.
func matrixUser() *domain.User {
	return &domain.User{
		ID:       uuid.New(),
		Username: "matrix-user",
		Proxies: domain.UserCredentials{
			VMessUUID:    uuid.New(),
			VLESSUUID:    uuid.New(),
			TrojanPass:   "trojan-pass",
			ShadowsocksP: "ss-password-0123456789abcdef",
			SSMethod:     "aes-128-gcm",
		},
	}
}

// networkLabel renders the (possibly empty) network value for sub-test names.
func networkLabel(n string) string {
	if n == "" {
		return "udp-native"
	}
	return n
}

// mandatesTLS reports protocols that cannot start without a TLS layer. The guard
// admits them with security=none (security is a free matrix axis), but in
// production provisionSecurity forces TLS onto them before they reach the
// renderer; hysteria/anytls would otherwise be skipped by inboundUsable. The
// matrix test mirrors that coercion so it exercises the real, renderable combo.
func mandatesTLS(p domain.Protocol) bool {
	switch p {
	case domain.ProtoHysteria, domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoAnyTLS:
		return true
	}
	return false
}

// matrixInbound builds the minimal-but-usable inbound for one matrix combo,
// supplying the security material each combo needs (reality key / TLS cert /
// ShadowTLS handshake target / WireGuard server key) and coercing mandatory-TLS
// protocols away from security=none.
func matrixInbound(proto domain.Protocol, network string, security domain.Security) domain.Inbound {
	in := domain.Inbound{
		Tag:      fmt.Sprintf("%s-%s-%s", proto, networkLabel(network), security),
		Protocol: proto,
		Listen:   "::",
		Port:     20000,
		Network:  network,
		Security: security,
		SNI:      []string{"matrix.example.com"},
		Path:     "/matrix",
		Raw:      map[string]any{},
	}

	if mandatesTLS(proto) && security == domain.SecurityNone {
		in.Security = domain.SecurityTLS
		security = domain.SecurityTLS
	}

	switch security {
	case domain.SecurityReality:
		in.Raw["reality"] = map[string]any{
			"private_key":  "MATRIX_FAKE_PRIVATE_KEY",
			"server_names": []any{"matrix.example.com"},
			"dest":         "matrix.example.com:443",
		}
	case domain.SecurityTLS:
		in.Raw["tls"] = map[string]any{"certificate": fakeCertPEM, "key": fakeKeyPEM}
	}

	switch proto {
	case domain.ProtoShadowTLS:
		// ShadowTLS fronts a real TLS handshake target; without it inboundUsable
		// skips the inbound. It carries no tls block of its own (tlsBlock returns
		// nil for shadowtls), so the security axis above is inert for it.
		in.Raw["shadowtls"] = map[string]any{
			"handshake_server": "matrix.example.com",
			"handshake_port":   443,
		}
	case domain.ProtoWireGuard:
		// WireGuard renders as a top-level `endpoints` entry, not an inbound; give
		// it a server key/subnet so the endpoint is materially complete.
		in.Raw["wireguard"] = map[string]any{
			"private_key": "MATRIX_FAKE_WG_KEY",
			"subnet":      "10.7.0.0/24",
			"listen_port": 20000,
		}
	}
	return in
}

// TestBuilder_MatrixRendersEveryAcceptedCombo iterates the sing-box capability
// matrix (Protocols × Transports × Securities, with UDP-native protocols pinned
// to an empty transport) and asserts EVERY combination the guard accepts renders
// into exactly one rendered entry for its tag — as an inbound for regular
// protocols, or a top-level endpoint for WireGuard. This proves there is no
// save-then-reject gap between the matrix and the renderer. Data-driven off
// core.Capabilities so later additions are covered automatically.
func TestBuilder_MatrixRendersEveryAcceptedCombo(t *testing.T) {
	const coreType = domain.CoreSingbox
	caps := core.Capabilities(coreType)

	for _, proto := range caps.Protocols {
		networks := caps.Transports
		if core.SkipsTransport(coreType, proto) {
			networks = []string{""} // no stream transport: network is irrelevant
		}
		for _, network := range networks {
			for _, security := range core.AllowedSecurities(coreType, proto) {
				proto, network, security := proto, network, security
				name := fmt.Sprintf("%s_%s_%s", proto, networkLabel(network), security)
				t.Run(name, func(t *testing.T) {
					if err := core.Supports(coreType, proto, network, security); err != nil {
						t.Fatalf("combo is not matrix-accepted (test setup bug): %v", err)
					}
					in := matrixInbound(proto, network, security)
					raw, err := Builder{APIPort: 9090}.Build(&core.GeneratedConfig{
						Inbounds:       []domain.Inbound{in},
						UsersByInbound: map[string][]*domain.User{in.Tag: {matrixUser()}},
					})
					if err != nil {
						t.Fatalf("Build returned error for %s: %v", name, err)
					}
					asEndpoint := proto == domain.ProtoWireGuard
					assertRendered(t, raw, in.Tag, string(proto), asEndpoint)
				})
			}
		}
	}
}

// assertRendered parses the rendered sing-box config and asserts the given tag
// appears exactly once — in `endpoints` when asEndpoint is true (WireGuard),
// otherwise in `inbounds` — and that its "type" matches.
func assertRendered(t *testing.T, raw []byte, tag, wantType string, asEndpoint bool) {
	t.Helper()
	var parsed struct {
		Inbounds []struct {
			Tag  string `json:"tag"`
			Type string `json:"type"`
		} `json:"inbounds"`
		Endpoints []struct {
			Tag  string `json:"tag"`
			Type string `json:"type"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("generated config is not valid JSON: %v\n%s", err, raw)
	}

	where := "inbounds"
	type entry struct{ Tag, Type string }
	var entries []entry
	if asEndpoint {
		where = "endpoints"
		for _, e := range parsed.Endpoints {
			entries = append(entries, entry{e.Tag, e.Type})
		}
	} else {
		for _, e := range parsed.Inbounds {
			entries = append(entries, entry{e.Tag, e.Type})
		}
	}

	var types []string
	for _, e := range entries {
		if e.Tag == tag {
			types = append(types, e.Type)
		}
	}
	if len(types) != 1 {
		t.Fatalf("tag %q rendered %d times in %s, want exactly 1 (silently skipped or duplicated?)", tag, len(types), where)
	}
	if types[0] != wantType {
		t.Errorf("tag %q type = %q, want %q", tag, types[0], wantType)
	}
}

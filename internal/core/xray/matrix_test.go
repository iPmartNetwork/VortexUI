package xray

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// Fake PEM material — the renderer only splits these into line arrays, it never
// parses them, so syntactically-bogus PEM is enough to make a TLS inbound
// "usable" (inboundUsable requires a non-empty certificate+key pair).
const (
	fakeCertPEM = "-----BEGIN CERTIFICATE-----\nMIIBmatrixfake\n-----END CERTIFICATE-----"
	fakeKeyPEM  = "-----BEGIN PRIVATE KEY-----\nMIIBmatrixfake\n-----END PRIVATE KEY-----"
)

// matrixUser returns a user with every per-protocol credential populated, so the
// same user is renderable on any protocol the matrix throws at it.
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

// networkLabel renders the (possibly empty) network value into a readable
// sub-test name. An empty network means a UDP-native protocol.
func networkLabel(n string) string {
	if n == "" {
		return "udp-native"
	}
	return n
}

// matrixInbound builds the minimal-but-usable inbound for one matrix combo: the
// security material (reality private key / TLS cert) needed so inboundUsable
// accepts it, plus a tag/port/SNI/path. xray has no UDP-native or mandatory-TLS
// protocols, so no protocol-specific coercion is needed here.
func matrixInbound(proto domain.Protocol, network string, security domain.Security) domain.Inbound {
	in := domain.Inbound{
		Tag:      fmt.Sprintf("%s-%s-%s", proto, networkLabel(network), security),
		Protocol: proto,
		Listen:   "0.0.0.0",
		Port:     20000,
		Network:  network,
		Security: security,
		SNI:      []string{"matrix.example.com"},
		Path:     "/matrix",
		Raw:      map[string]any{},
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
	return in
}

// TestBuilder_MatrixRendersEveryAcceptedCombo iterates the xray capability matrix
// (Protocols × Transports × Securities) and asserts that EVERY combination the
// guard would accept actually renders into exactly one xray inbound for its tag.
// This proves there is no save-then-reject gap: anything core.Supports admits is
// renderable by the core (no silent skip, no error). Data-driven off
// core.Capabilities so protocols/transports added later are covered automatically.
func TestBuilder_MatrixRendersEveryAcceptedCombo(t *testing.T) {
	const coreType = domain.CoreXray
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
					// Sanity: the combo we render must be one the guard accepts.
					if err := core.Supports(coreType, proto, network, security); err != nil {
						t.Fatalf("combo is not matrix-accepted (test setup bug): %v", err)
					}
					in := matrixInbound(proto, network, security)
					raw, err := Builder{APIPort: 10085}.Build(&core.GeneratedConfig{
						Inbounds:       []domain.Inbound{in},
						UsersByInbound: map[string][]*domain.User{in.Tag: {matrixUser()}},
					})
					if err != nil {
						t.Fatalf("Build returned error for %s: %v", name, err)
					}
					// dokodemo renders under its xray WIRE name "dokodemo-door";
					// every other protocol renders under its value verbatim.
					want := string(proto)
					if proto == domain.ProtoDokodemo {
						want = "dokodemo-door"
					}
					assertSingleInbound(t, raw, in.Tag, want)
				})
			}
		}
	}
}

// assertSingleInbound parses the rendered xray config and asserts the given tag
// appears exactly once among the inbounds (it was not silently skipped) and that
// its "protocol" matches.
func assertSingleInbound(t *testing.T, raw []byte, tag, wantProto string) {
	t.Helper()
	var parsed struct {
		Inbounds []struct {
			Tag      string `json:"tag"`
			Protocol string `json:"protocol"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("generated config is not valid JSON: %v\n%s", err, raw)
	}
	var protos []string
	for _, in := range parsed.Inbounds {
		if in.Tag == tag {
			protos = append(protos, in.Protocol)
		}
	}
	if len(protos) != 1 {
		t.Fatalf("tag %q rendered %d times, want exactly 1 (silently skipped or duplicated?)", tag, len(protos))
	}
	if protos[0] != wantProto {
		t.Errorf("tag %q protocol = %q, want %q", tag, protos[0], wantProto)
	}
}

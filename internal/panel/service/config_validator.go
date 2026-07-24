package service

import (
	"fmt"
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
)

// ConfigValidator validates inbound configuration JSON for protocol/transport/security
// combinations and provides sensible defaults.
type ConfigValidator struct{}

// NewConfigValidator creates a new ConfigValidator.
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate checks config data against the protocol+network+security schema.
// It returns a list of field-level errors (empty if valid).
func (v *ConfigValidator) Validate(protocol domain.Protocol, network string, security domain.Security, config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError

	// Protocol-specific validation
	switch protocol {
	case domain.ProtoVMess, domain.ProtoVLESS:
		errs = append(errs, v.validateXRayProxy(protocol, network, security, config)...)
	case domain.ProtoTrojan:
		errs = append(errs, v.validateTrojan(network, security, config)...)
	case domain.ProtoShadowsocks:
		errs = append(errs, v.validateShadowsocks(config)...)
	case domain.ProtoHysteria2:
		errs = append(errs, v.validateHysteria2(config)...)
	case domain.ProtoTUIC:
		errs = append(errs, v.validateTUIC(config)...)
	case domain.ProtoWireGuard:
		errs = append(errs, v.validateWireGuard(config)...)
	case domain.ProtoShadowTLS:
		errs = append(errs, v.validateShadowTLS(config)...)
	case domain.ProtoNaive:
		errs = append(errs, v.validateNaive(security, config)...)
	default:
		// Unknown protocols pass through without schema validation.
	}

	// Transport validation (common to stream-based protocols)
	if isStreamProtocol(protocol) {
		errs = append(errs, v.validateTransport(network, config)...)
	}

	// Security/TLS validation
	errs = append(errs, v.validateSecurity(security, config)...)

	return errs
}

// DefaultsFor returns valid default config for the given protocol/transport/security combination.
func (v *ConfigValidator) DefaultsFor(protocol domain.Protocol, network string, security domain.Security) domain.ConfigDefaults {
	config := make(map[string]any)

	switch protocol {
	case domain.ProtoVMess:
		config["alterId"] = 0
	case domain.ProtoVLESS:
		if security == domain.SecurityReality {
			config["flow"] = "xtls-rprx-vision"
		}
	case domain.ProtoTrojan:
		// Trojan defaults require TLS
	case domain.ProtoShadowsocks:
		config["method"] = "2022-blake3-aes-128-gcm"
	case domain.ProtoHysteria2:
		config["up_mbps"] = 100
		config["down_mbps"] = 100
	case domain.ProtoTUIC:
		config["congestion_control"] = "bbr"
	case domain.ProtoWireGuard:
		config["mtu"] = 1420
		config["subnet"] = "10.7.0.0/24"
		config["dns"] = "1.1.1.1"
	case domain.ProtoShadowTLS:
		config["version"] = 3
		config["handshake_server"] = "www.microsoft.com"
		config["handshake_port"] = 443
	case domain.ProtoNaive:
		// Naive mandates TLS, no extra defaults needed
	}

	// Transport defaults
	switch network {
	case "ws":
		config["path"] = "/"
		config["host"] = ""
	case "grpc":
		config["serviceName"] = "vortex"
		config["multiMode"] = false
	case "httpupgrade":
		config["path"] = "/"
		config["host"] = ""
	case "xhttp":
		config["path"] = "/"
		config["mode"] = "auto"
	case "tcp":
		config["headerType"] = "none"
	}

	// Security defaults
	switch security {
	case domain.SecurityTLS:
		config["sni"] = ""
		config["alpn"] = []string{"h2", "http/1.1"}
		config["fingerprint"] = "chrome"
	case domain.SecurityReality:
		config["sni"] = ""
		config["fingerprint"] = "chrome"
		config["publicKey"] = ""
		config["shortId"] = ""
		config["spiderX"] = ""
	}

	return domain.ConfigDefaults{
		Protocol: protocol,
		Network:  network,
		Security: security,
		Config:   config,
	}
}

// --- protocol-specific validators ---

func (v *ConfigValidator) validateXRayProxy(protocol domain.Protocol, network string, security domain.Security, config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError

	if protocol == domain.ProtoVMess {
		if alter, ok := config["alterId"]; ok {
			switch val := alter.(type) {
			case float64:
				if val < 0 {
					errs = append(errs, domain.ConfigValidationError{Field: "alterId", Message: "alterId must be non-negative"})
				}
			case int:
				if val < 0 {
					errs = append(errs, domain.ConfigValidationError{Field: "alterId", Message: "alterId must be non-negative"})
				}
			}
		}
	}

	if protocol == domain.ProtoVLESS && security == domain.SecurityReality {
		if flow, _ := config["flow"].(string); flow != "" && flow != "xtls-rprx-vision" {
			errs = append(errs, domain.ConfigValidationError{Field: "flow", Message: "VLESS Reality flow must be 'xtls-rprx-vision' or empty"})
		}
	}

	return errs
}

func (v *ConfigValidator) validateTrojan(network string, security domain.Security, config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	if security == domain.SecurityNone && network != "ws" {
		errs = append(errs, domain.ConfigValidationError{Field: "security", Message: "Trojan requires TLS or Reality (or WebSocket for relay)"})
	}
	return errs
}

func (v *ConfigValidator) validateShadowsocks(config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	validMethods := map[string]bool{
		"2022-blake3-aes-128-gcm":       true,
		"2022-blake3-aes-256-gcm":       true,
		"2022-blake3-chacha20-poly1305": true,
		"aes-128-gcm":                   true,
		"aes-256-gcm":                   true,
		"chacha20-ietf-poly1305":        true,
		"xchacha20-ietf-poly1305":       true,
		"none":                           true,
	}
	if method, _ := config["method"].(string); method != "" && !validMethods[method] {
		errs = append(errs, domain.ConfigValidationError{Field: "method", Message: fmt.Sprintf("unsupported encryption method: %s", method)})
	}
	return errs
}

func (v *ConfigValidator) validateHysteria2(config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	if up, ok := config["up_mbps"]; ok {
		if val, _ := toFloat(up); val <= 0 {
			errs = append(errs, domain.ConfigValidationError{Field: "up_mbps", Message: "up_mbps must be positive"})
		}
	}
	if down, ok := config["down_mbps"]; ok {
		if val, _ := toFloat(down); val <= 0 {
			errs = append(errs, domain.ConfigValidationError{Field: "down_mbps", Message: "down_mbps must be positive"})
		}
	}
	return errs
}

func (v *ConfigValidator) validateTUIC(config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	validCC := map[string]bool{"cubic": true, "bbr": true, "new_reno": true}
	if cc, _ := config["congestion_control"].(string); cc != "" && !validCC[cc] {
		errs = append(errs, domain.ConfigValidationError{Field: "congestion_control", Message: "congestion_control must be cubic, bbr, or new_reno"})
	}
	return errs
}

func (v *ConfigValidator) validateWireGuard(config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	if mtu, ok := config["mtu"]; ok {
		if val, _ := toFloat(mtu); val < 1280 || val > 1500 {
			errs = append(errs, domain.ConfigValidationError{Field: "mtu", Message: "MTU must be between 1280 and 1500"})
		}
	}
	return errs
}

func (v *ConfigValidator) validateShadowTLS(config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	if ver, ok := config["version"]; ok {
		if val, _ := toFloat(ver); val < 1 || val > 3 {
			errs = append(errs, domain.ConfigValidationError{Field: "version", Message: "ShadowTLS version must be 1, 2, or 3"})
		}
	}
	if hs, _ := config["handshake_server"].(string); hs == "" {
		errs = append(errs, domain.ConfigValidationError{Field: "handshake_server", Message: "handshake_server is required"})
	}
	return errs
}

func (v *ConfigValidator) validateNaive(security domain.Security, config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError
	if security != domain.SecurityTLS {
		errs = append(errs, domain.ConfigValidationError{Field: "security", Message: "NaiveProxy requires TLS"})
	}
	return errs
}

// --- transport validation ---

func (v *ConfigValidator) validateTransport(network string, config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError

	switch network {
	case "ws":
		if path, _ := config["path"].(string); path != "" && !strings.HasPrefix(path, "/") {
			errs = append(errs, domain.ConfigValidationError{Field: "path", Message: "WebSocket path must start with /"})
		}
	case "grpc":
		if sn, _ := config["serviceName"].(string); sn != "" {
			for _, c := range sn {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
					errs = append(errs, domain.ConfigValidationError{Field: "serviceName", Message: "gRPC serviceName contains invalid characters"})
					break
				}
			}
		}
	case "httpupgrade":
		if path, _ := config["path"].(string); path != "" && !strings.HasPrefix(path, "/") {
			errs = append(errs, domain.ConfigValidationError{Field: "path", Message: "HTTPUpgrade path must start with /"})
		}
	case "xhttp":
		validModes := map[string]bool{"auto": true, "packet-up": true, "stream-up": true}
		if mode, _ := config["mode"].(string); mode != "" && !validModes[mode] {
			errs = append(errs, domain.ConfigValidationError{Field: "mode", Message: "XHTTP mode must be auto, packet-up, or stream-up"})
		}
	case "tcp":
		validHeaders := map[string]bool{"none": true, "http": true}
		if ht, _ := config["headerType"].(string); ht != "" && !validHeaders[ht] {
			errs = append(errs, domain.ConfigValidationError{Field: "headerType", Message: "TCP headerType must be none or http"})
		}
	}

	return errs
}

// --- security validation ---

func (v *ConfigValidator) validateSecurity(security domain.Security, config map[string]any) []domain.ConfigValidationError {
	var errs []domain.ConfigValidationError

	switch security {
	case domain.SecurityTLS:
		if alpn, ok := config["alpn"]; ok {
			switch val := alpn.(type) {
			case []any:
				validALPN := map[string]bool{"h2": true, "http/1.1": true, "h3": true}
				for _, a := range val {
					if s, ok := a.(string); ok && !validALPN[s] {
						errs = append(errs, domain.ConfigValidationError{Field: "alpn", Message: fmt.Sprintf("unsupported ALPN value: %s", s)})
					}
				}
			case []string:
				validALPN := map[string]bool{"h2": true, "http/1.1": true, "h3": true}
				for _, s := range val {
					if !validALPN[s] {
						errs = append(errs, domain.ConfigValidationError{Field: "alpn", Message: fmt.Sprintf("unsupported ALPN value: %s", s)})
					}
				}
			}
		}
	case domain.SecurityReality:
		if sid, _ := config["shortId"].(string); sid != "" {
			if len(sid) > 16 {
				errs = append(errs, domain.ConfigValidationError{Field: "shortId", Message: "shortId must be at most 16 hex characters"})
			}
			for _, c := range sid {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					errs = append(errs, domain.ConfigValidationError{Field: "shortId", Message: "shortId must be hexadecimal"})
					break
				}
			}
		}
	}

	return errs
}

// --- helpers ---

func isStreamProtocol(p domain.Protocol) bool {
	switch p {
	case domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks:
		return true
	default:
		return false
	}
}

func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

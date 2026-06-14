// Package reality generates the cryptographic material an Xray/sing-box REALITY
// inbound needs: an X25519 key pair (the private key goes in the server config,
// the public key goes to clients) and short IDs. It is pure and dependency-light
// so it can be unit-tested and reused by both the config builder and the API.
package reality

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// Keypair is a REALITY X25519 key pair, base64 raw-url encoded exactly as
// Xray/sing-box expect them on the command line and in config.
type Keypair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// GenerateKeypair creates a fresh X25519 key pair. The private key is clamped
// per RFC 7748 so the derived public key matches what `xray x25519` produces,
// keeping our output interchangeable with the reference tooling.
func GenerateKeypair() (Keypair, error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return Keypair{}, fmt.Errorf("reality: read random: %w", err)
	}
	// Clamp the scalar (RFC 7748 §5): the standard X25519 private-key form.
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pub, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return Keypair{}, fmt.Errorf("reality: derive public key: %w", err)
	}
	enc := base64.RawURLEncoding
	return Keypair{
		PrivateKey: enc.EncodeToString(priv[:]),
		PublicKey:  enc.EncodeToString(pub),
	}, nil
}

// PublicKeyFor derives the public key from an existing base64 raw-url private
// key, so a stored inbound can hand clients the matching public key without
// persisting it separately.
func PublicKeyFor(privateKey string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("reality: decode private key: %w", err)
	}
	if len(raw) != 32 {
		return "", fmt.Errorf("reality: private key must be 32 bytes, got %d", len(raw))
	}
	pub, err := curve25519.X25519(raw, curve25519.Basepoint)
	if err != nil {
		return "", fmt.Errorf("reality: derive public key: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(pub), nil
}

// ShortID returns a random REALITY short ID: a hex string of the given byte
// length (0 < n <= 8). REALITY accepts short IDs of 0–8 bytes; 8 is the common
// default. The result is 2*n hex characters.
func ShortID(n int) (string, error) {
	if n <= 0 || n > 8 {
		return "", fmt.Errorf("reality: short id length must be 1..8 bytes, got %d", n)
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("reality: read random: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Params is the engine-neutral REALITY server material that both the Xray and
// sing-box builders translate into their native shapes. Storing one neutral
// representation (under Inbound.Raw["reality"]) keeps the panel core-agnostic.
type Params struct {
	PrivateKey  string   // server private key (base64 raw-url)
	PublicKey   string   // client-facing public key; not used server-side
	ShortIDs    []string // accepted short IDs
	ServerNames []string // SNIs the inbound impersonates
	Dest        string   // handshake target, "host:port"
}

// ParseParams extracts REALITY params from the value stored at
// Inbound.Raw["reality"]. It is tolerant: a nil or wrongly-typed value yields a
// zero Params, and each field is read defensively from the decoded JSON map.
func ParseParams(v any) Params {
	m, ok := v.(map[string]any)
	if !ok {
		return Params{}
	}
	return Params{
		PrivateKey:  asString(m["private_key"]),
		PublicKey:   asString(m["public_key"]),
		ShortIDs:    asStrings(m["short_ids"]),
		ServerNames: asStrings(m["server_names"]),
		Dest:        asString(m["dest"]),
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asStrings(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

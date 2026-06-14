package reality

import (
	"encoding/base64"
	"testing"
)

func TestGenerateKeypairIsValidAndDeterministicallyDerived(t *testing.T) {
	kp, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}

	// Both keys must be 32-byte raw-url base64.
	priv, err := base64.RawURLEncoding.DecodeString(kp.PrivateKey)
	if err != nil || len(priv) != 32 {
		t.Fatalf("private key not 32-byte raw-url base64: %v len=%d", err, len(priv))
	}
	pub, err := base64.RawURLEncoding.DecodeString(kp.PublicKey)
	if err != nil || len(pub) != 32 {
		t.Fatalf("public key not 32-byte raw-url base64: %v len=%d", err, len(pub))
	}

	// The public key must be reproducible from the private key.
	derived, err := PublicKeyFor(kp.PrivateKey)
	if err != nil {
		t.Fatalf("PublicKeyFor: %v", err)
	}
	if derived != kp.PublicKey {
		t.Errorf("derived public key %q != generated %q", derived, kp.PublicKey)
	}

	// The scalar must be clamped (RFC 7748): low 3 bits clear, bit 254 set, bit
	// 255 clear.
	if priv[0]&0b111 != 0 || priv[31]&0b1000_0000 != 0 || priv[31]&0b0100_0000 == 0 {
		t.Errorf("private key not clamped: first=%08b last=%08b", priv[0], priv[31])
	}
}

func TestGenerateKeypairIsRandom(t *testing.T) {
	a, _ := GenerateKeypair()
	b, _ := GenerateKeypair()
	if a.PrivateKey == b.PrivateKey || a.PublicKey == b.PublicKey {
		t.Error("two generated keypairs must differ")
	}
}

func TestPublicKeyForRejectsBadInput(t *testing.T) {
	if _, err := PublicKeyFor("not-base64!!"); err == nil {
		t.Error("expected error for invalid base64")
	}
	if _, err := PublicKeyFor(base64.RawURLEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Error("expected error for wrong-length key")
	}
}

func TestShortID(t *testing.T) {
	for _, n := range []int{1, 4, 8} {
		id, err := ShortID(n)
		if err != nil {
			t.Fatalf("ShortID(%d): %v", n, err)
		}
		if len(id) != 2*n {
			t.Errorf("ShortID(%d) length = %d, want %d hex chars", n, len(id), 2*n)
		}
	}
	if _, err := ShortID(0); err == nil {
		t.Error("ShortID(0) should error")
	}
	if _, err := ShortID(9); err == nil {
		t.Error("ShortID(9) should error (max 8 bytes)")
	}
	// Successive short IDs should differ.
	a, _ := ShortID(8)
	b, _ := ShortID(8)
	if a == b {
		t.Error("two short ids must differ")
	}
}

package wireguard

import (
	"encoding/base64"
	"testing"
)

func TestGenerateKeypairProducesDistinctNonEmptyKeys(t *testing.T) {
	priv, pub, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	if priv == "" || pub == "" {
		t.Fatalf("keys must be non-empty: priv=%q pub=%q", priv, pub)
	}
	if priv == pub {
		t.Fatal("private and public keys must differ")
	}
	for name, k := range map[string]string{"private": priv, "public": pub} {
		b, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			t.Errorf("%s key is not valid base64: %v", name, err)
		}
		if len(b) != 32 {
			t.Errorf("%s key decodes to %d bytes, want 32", name, len(b))
		}
	}

	// A second call must yield a different keypair.
	priv2, pub2, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair (2): %v", err)
	}
	if priv == priv2 || pub == pub2 {
		t.Error("consecutive keypairs must be random/distinct")
	}
}

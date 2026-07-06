package service

import "testing"

func TestSniProbeTLSConfigVerifiesChain(t *testing.T) {
	cfg := sniProbeTLSConfig("example.com")
	if cfg.InsecureSkipVerify {
		t.Fatal("SNI probe must verify the certificate chain")
	}
	if cfg.ServerName != "example.com" {
		t.Errorf("ServerName = %q, want example.com", cfg.ServerName)
	}
	if cfg.RootCAs == nil {
		t.Error("expected system or fallback root CAs")
	}
}

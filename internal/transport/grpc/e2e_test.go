package grpc

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/pki"
)

// writeChain materializes a dev PKI to disk and returns the server (node) and
// client (panel) TLS file sets, since ServerCreds/ClientCreds load from paths.
func writeChain(t *testing.T, chain *pki.PKI) (server, client TLSFiles) {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, data []byte) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	caPath := write("ca.crt", chain.CA.CertPEM)
	server = TLSFiles{Cert: write("node.crt", chain.Server.CertPEM), Key: write("node.key", chain.Server.KeyPEM), CA: caPath}
	client = TLSFiles{Cert: write("panel.crt", chain.Client.CertPEM), Key: write("panel.key", chain.Client.KeyPEM), CA: caPath}
	return server, client
}

// startMTLSNode boots a NodeServer over real mTLS on a real TCP port and returns
// its address. The server stops on test cleanup.
func startMTLSNode(t *testing.T, drv core.CoreDriver, serverFiles TLSFiles) string {
	t.Helper()
	creds, err := ServerCreds(serverFiles)
	if err != nil {
		t.Fatalf("server creds: %v", err)
	}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := NewNodeServer(drv, "e2e")
	go func() { _ = srv.Serve(lis, grpc.Creds(creds)) }()
	t.Cleanup(srv.GracefulStop)
	return lis.Addr().String()
}

func TestE2E_MTLSNodeContract(t *testing.T) {
	chain, err := pki.Generate([]string{"localhost", "127.0.0.1"})
	if err != nil {
		t.Fatalf("pki: %v", err)
	}
	serverFiles, clientFiles := writeChain(t, chain)

	uid := uuid.New()
	drv := &fakeDriver{deltas: []domain.TrafficDelta{{UserID: uid, Up: 11, Down: 22, Timestamp: time.Now()}}}
	addr := startMTLSNode(t, drv, serverFiles)

	// The node cert's SAN includes "localhost"; that is the name we verify.
	creds, err := ClientCreds(clientFiles, "localhost")
	if err != nil {
		t.Fatalf("client creds: %v", err)
	}
	client, err := Dial(uuid.New(), addr, creds)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A full handshake + RPC round-trip proves the mTLS link end to end.
	if err := client.Sync(ctx, &core.GeneratedConfig{LogLevel: "warn"}, domain.CoreXray); err != nil {
		t.Fatalf("sync over mTLS: %v", err)
	}
	if err := client.AddUser(ctx, "in-1", &domain.User{ID: uid}); err != nil {
		t.Fatalf("add user over mTLS: %v", err)
	}
	if h, err := client.Health(ctx); err != nil || !h.CoreRunning {
		t.Fatalf("health over mTLS: %+v err=%v", h, err)
	}

	got := make(chan domain.TrafficDelta, 1)
	go func() { _ = client.ConsumeTraffic(ctx, func(d domain.TrafficDelta) { got <- d }) }()
	select {
	case d := <-got:
		if d.Up != 11 || d.Down != 22 {
			t.Errorf("delta = %d/%d, want 11/22", d.Up, d.Down)
		}
	case <-ctx.Done():
		t.Fatal("no traffic delta over mTLS stream")
	}
}

func TestE2E_RejectsUntrustedClient(t *testing.T) {
	// Server trusts CA #1; client presents a cert from CA #2 -> handshake fails.
	serverChain, _ := pki.Generate([]string{"localhost"})
	foreignChain, _ := pki.Generate([]string{"localhost"})
	serverFiles, _ := writeChain(t, serverChain)
	_, foreignClient := writeChain(t, foreignChain)

	addr := startMTLSNode(t, &fakeDriver{}, serverFiles)

	creds, err := ClientCreds(foreignClient, "localhost")
	if err != nil {
		t.Fatalf("client creds: %v", err)
	}
	client, err := Dial(uuid.New(), addr, creds)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.Health(ctx); err == nil {
		t.Fatal("expected mTLS handshake to reject a client signed by an untrusted CA")
	}
}

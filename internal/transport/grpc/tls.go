// Package grpc wires the generated NodeService stubs to real transport: a node
// agent serves NodeService, and the panel dials it. All links are mutually
// authenticated (mTLS): both ends present certificates signed by a shared CA,
// so a leaked node cert alone cannot impersonate the panel and vice versa.
package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// TLSFiles points at the PEM material for one side of an mTLS link.
type TLSFiles struct {
	Cert string // this side's certificate
	Key  string // this side's private key
	CA   string // CA that signed the *peer's* certificate
}

func (f TLSFiles) validate() error {
	if f.Cert == "" || f.Key == "" || f.CA == "" {
		return errors.New("mTLS requires cert, key, and ca paths")
	}
	return nil
}

func loadCAPool(caPath string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read CA %q: %w", caPath, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no valid certificates found in CA %q", caPath)
	}
	return pool, nil
}

// ServerCreds builds transport credentials for the node-side gRPC server that
// require and verify a client (panel) certificate.
func ServerCreds(f TLSFiles) (credentials.TransportCredentials, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(f.Cert, f.Key)
	if err != nil {
		return nil, fmt.Errorf("load server keypair: %w", err)
	}
	pool, err := loadCAPool(f.CA)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS13,
	}), nil
}

// ClientCreds builds transport credentials for the panel dialing a node. serverName
// must match the node certificate's SAN.
func ClientCreds(f TLSFiles, serverName string) (credentials.TransportCredentials, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(f.Cert, f.Key)
	if err != nil {
		return nil, fmt.Errorf("load client keypair: %w", err)
	}
	pool, err := loadCAPool(f.CA)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}), nil
}

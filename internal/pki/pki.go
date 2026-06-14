// Package pki generates a small development mTLS trust chain: a self-signed CA
// plus a server (node) and client (panel) certificate it signs. It is meant for
// local/dev setup and tests — production should mint certs with a real CA — but
// the certs it issues are ordinary X.509 and work identically to any other.
package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// KeyPair is a PEM-encoded certificate and its private key.
type KeyPair struct {
	CertPEM []byte
	KeyPEM  []byte
}

// PKI is a complete dev trust chain: the CA and the two leaf certs it signs.
type PKI struct {
	CA     KeyPair // trust anchor both sides verify the peer against
	Server KeyPair // node agent (gRPC server)
	Client KeyPair // panel (gRPC client)
}

// Generate builds a fresh CA and issues a server cert valid for the given SANs
// (hostnames and/or IPs) and a client cert. Certificates are valid for ~10 years
// — fine for dev, and avoids surprise expiry mid-demo.
func Generate(serverSANs []string) (*PKI, error) {
	caCert, caKey, caPEM, caKeyPEM, err := newCA()
	if err != nil {
		return nil, err
	}

	dns, ips := splitSANs(serverSANs)
	server, err := issue(caCert, caKey, "vortex-node", dns, ips, x509.ExtKeyUsageServerAuth)
	if err != nil {
		return nil, err
	}
	client, err := issue(caCert, caKey, "vortex-panel", nil, nil, x509.ExtKeyUsageClientAuth)
	if err != nil {
		return nil, err
	}

	return &PKI{
		CA:     KeyPair{CertPEM: caPEM, KeyPEM: caKeyPEM},
		Server: *server,
		Client: *client,
	}, nil
}

func newCA() (*x509.Certificate, *ecdsa.PrivateKey, []byte, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial(),
		Subject:               pkix.Name{CommonName: "VortexUI Dev CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	keyPEM, err := encodeKey(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return cert, key, encodeCert(der), keyPEM, nil
}

func issue(ca *x509.Certificate, caKey *ecdsa.PrivateKey, cn string, dns []string, ips []net.IP, eku x509.ExtKeyUsage) (*KeyPair, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial(),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{eku},
		DNSNames:     dns,
		IPAddresses:  ips,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	keyPEM, err := encodeKey(key)
	if err != nil {
		return nil, err
	}
	return &KeyPair{CertPEM: encodeCert(der), KeyPEM: keyPEM}, nil
}

func splitSANs(sans []string) (dns []string, ips []net.IP) {
	for _, s := range sans {
		if ip := net.ParseIP(s); ip != nil {
			ips = append(ips, ip)
		} else {
			dns = append(dns, s)
		}
	}
	return dns, ips
}

func serial() *big.Int {
	n, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	return n
}

func encodeCert(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func encodeKey(key *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), nil
}

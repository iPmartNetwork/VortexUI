package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnrollmentService builds node enrollment bundles from the panel's mTLS cert dir.
type EnrollmentService struct {
	CAPath string // path to ca.crt; node.crt and node.key are read from the same dir
}

// EnrollmentBundle is returned by the panel API for the node wizard.
type EnrollmentBundle struct {
	Bundle        string `json:"bundle"`
	CAFingerprint string `json:"ca_fingerprint"`
	CertDir       string `json:"cert_dir"`
}

// Bundle reads ca.crt, node.crt, node.key and returns a base64 tarball plus the CA fingerprint.
func (s *EnrollmentService) Bundle() (*EnrollmentBundle, error) {
	if s.CAPath == "" {
		return nil, errors.New("CA path not configured")
	}
	dir := filepath.Dir(s.CAPath)
	files := []string{"ca.crt", "node.crt", "node.key"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return nil, fmt.Errorf("missing %s in %s", f, dir)
		}
	}
	fp, err := caFingerprint(s.CAPath)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(data))}); err != nil {
			return nil, err
		}
		if _, err := tw.Write(data); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return &EnrollmentBundle{
		Bundle:        base64.StdEncoding.EncodeToString(buf.Bytes()),
		CAFingerprint: fp,
		CertDir:       dir,
	}, nil
}

func caFingerprint(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return "", fmt.Errorf("invalid PEM in %s", path)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(cert.Raw)
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = strings.ToUpper(fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, ":"), nil
}

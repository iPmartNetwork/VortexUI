package acme

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/acme"
)

// CloudflareDNS holds credentials for DNS-01 via Cloudflare.
type CloudflareDNS struct {
	Token  string
	ZoneID string
}

// obtainViaLetsEncrypt issues a cert using ACME DNS-01 (Cloudflare) when configured.
func (m *Manager) obtainViaLetsEncrypt(ctx context.Context, domain string, cf CloudflareDNS) (certPEM, keyPEM string, err error) {
	if m.email == "" {
		return "", "", errors.New("acme email required")
	}
	if cf.Token == "" || cf.ZoneID == "" {
		return "", "", errors.New("cloudflare credentials required for DNS-01")
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	client := &acme.Client{
		Key:          key,
		DirectoryURL: acme.LetsEncryptURL,
	}
	acc := &acme.Account{Contact: []string{"mailto:" + m.email}}
	if _, err := client.Register(ctx, acc, acme.AcceptTOS); err != nil {
		return "", "", fmt.Errorf("acme register: %w", err)
	}

	order, err := client.AuthorizeOrder(ctx, []acme.AuthzID{{Type: "dns", Value: domain}})
	if err != nil {
		return "", "", fmt.Errorf("acme order: %w", err)
	}

	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(ctx, authzURL)
		if err != nil {
			return "", "", err
		}
		if authz.Status == acme.StatusValid {
			continue
		}
		var chal *acme.Challenge
		for _, c := range authz.Challenges {
			if c.Type == "dns-01" {
				chal = c
				break
			}
		}
		if chal == nil {
			return "", "", errors.New("dns-01 challenge not offered")
		}
		val, err := client.DNS01ChallengeRecord(chal.Token)
		if err != nil {
			return "", "", err
		}
		recordName := "_acme-challenge." + domain
		if err := cf.upsertTXT(ctx, recordName, val); err != nil {
			return "", "", err
		}
		defer func() { _ = cf.deleteTXT(ctx, recordName) }()

		time.Sleep(3 * time.Second)
		if _, err := client.Accept(ctx, chal); err != nil {
			return "", "", fmt.Errorf("accept challenge: %w", err)
		}
		if _, err := client.WaitAuthorization(ctx, authz.URI); err != nil {
			return "", "", fmt.Errorf("wait authorization: %w", err)
		}
	}

	order, err = client.WaitOrder(ctx, order.URI)
	if err != nil {
		return "", "", fmt.Errorf("wait order: %w", err)
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		DNSNames: []string{domain},
	}, key)
	if err != nil {
		return "", "", err
	}
	derChain, _, err := client.CreateOrderCert(ctx, order.FinalizeURL, csr, true)
	if err != nil {
		return "", "", fmt.Errorf("finalize order: %w", err)
	}
	if len(derChain) == 0 {
		return "", "", errors.New("empty certificate chain")
	}

	var certBuf bytes.Buffer
	for _, der := range derChain {
		_ = pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", "", err
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return certBuf.String(), keyPEM, nil
}

func (cf CloudflareDNS) deleteTXT(ctx context.Context, name string) error {
	id, err := cf.findTXT(ctx, name)
	if err != nil || id == "" {
		return err
	}
	return cf.api(ctx, http.MethodDelete, "/dns_records/"+id, nil)
}

func (cf CloudflareDNS) upsertTXT(ctx context.Context, name, value string) error {
	if id, _ := cf.findTXT(ctx, name); id != "" {
		return cf.api(ctx, http.MethodPut, "/dns_records/"+id, map[string]any{
			"type":    "TXT",
			"name":    name,
			"content": value,
			"ttl":     120,
		})
	}
	return cf.api(ctx, http.MethodPost, "/dns_records", map[string]any{
		"type":    "TXT",
		"name":    name,
		"content": value,
		"ttl":     120,
	})
}

func (cf CloudflareDNS) findTXT(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=TXT&name=%s", cf.ZoneID, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cf.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if len(out.Result) == 0 {
		return "", nil
	}
	return out.Result[0].ID, nil
}

func (cf CloudflareDNS) api(ctx context.Context, method, path string, payload any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s%s", cf.ZoneID, path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cf.Token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("cloudflare %s %s: %s", method, path, strings.TrimSpace(string(raw)))
	}
	return nil
}

// Package dns provides programmatic DNS record management via cloud provider
// APIs (Cloudflare, etc.) so VortexUI can auto-create A records when nodes are
// added or endpoints change. This eliminates the manual "point DNS to this IP"
// step that trips up new users.
package dns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Provider is the interface for DNS record management.
type Provider interface {
	// UpsertA creates or updates an A record for the given FQDN.
	UpsertA(ctx context.Context, fqdn, ip string, proxied bool) error
	// DeleteA removes the A record for the given FQDN.
	DeleteA(ctx context.Context, fqdn string) error
}

// CloudflareConfig holds Cloudflare API credentials and zone info.
type CloudflareConfig struct {
	APIToken string // API token with DNS:Edit permission
	ZoneID   string // zone ID (from the Cloudflare dashboard)
}

// Cloudflare implements Provider for Cloudflare DNS.
type Cloudflare struct {
	cfg    CloudflareConfig
	client *http.Client
	log    *slog.Logger
}

// NewCloudflare builds a Cloudflare DNS provider.
func NewCloudflare(cfg CloudflareConfig, log *slog.Logger) *Cloudflare {
	if log == nil {
		log = slog.Default()
	}
	return &Cloudflare{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
		log:    log,
	}
}

// UpsertA creates or updates an A record. If a record with the same name exists,
// it is updated; otherwise a new one is created.
func (c *Cloudflare) UpsertA(ctx context.Context, fqdn, ip string, proxied bool) error {
	// Check if record exists
	existing, err := c.findRecord(ctx, fqdn, "A")
	if err != nil {
		return fmt.Errorf("lookup %s: %w", fqdn, err)
	}

	body := map[string]any{
		"type":    "A",
		"name":    fqdn,
		"content": ip,
		"ttl":     1, // auto
		"proxied": proxied,
	}

	if existing != "" {
		// Update existing record
		return c.apiCall(ctx, http.MethodPut, "/dns_records/"+existing, body)
	}
	// Create new record
	return c.apiCall(ctx, http.MethodPost, "/dns_records", body)
}

// DeleteA removes the A record for fqdn.
func (c *Cloudflare) DeleteA(ctx context.Context, fqdn string) error {
	id, err := c.findRecord(ctx, fqdn, "A")
	if err != nil || id == "" {
		return err
	}
	return c.apiCall(ctx, http.MethodDelete, "/dns_records/"+id, nil)
}

// findRecord looks up a DNS record by name and type, returning its ID or "".
func (c *Cloudflare) findRecord(ctx context.Context, name, typ string) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=%s", c.cfg.ZoneID, name, typ)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Result) > 0 {
		return result.Result[0].ID, nil
	}
	return "", nil
}

func (c *Cloudflare) apiCall(ctx context.Context, method, path string, body any) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s%s", c.cfg.ZoneID, path)
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare %s %s: %d %s", method, path, resp.StatusCode, string(respBody))
	}
	return nil
}

// NopProvider is a no-op DNS provider for when DNS automation is disabled.
type NopProvider struct{}

func (NopProvider) UpsertA(context.Context, string, string, bool) error { return nil }
func (NopProvider) DeleteA(context.Context, string) error               { return nil }

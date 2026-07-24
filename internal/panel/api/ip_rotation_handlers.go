package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// IPRotationProvider describes a cloud provider for IP rotation.
type IPRotationProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// supportedProviders is the list of cloud providers that support IP rotation.
var supportedProviders = []IPRotationProvider{
	{ID: "digitalocean", Name: "DigitalOcean"},
	{ID: "vultr", Name: "Vultr"},
	{ID: "hetzner", Name: "Hetzner"},
	{ID: "linode", Name: "Linode"},
}

// IPRotationHandler handles IP rotation operations across cloud providers.
type IPRotationHandler struct {
	client *http.Client
}

// NewIPRotationHandler creates a new IPRotationHandler with a default HTTP client.
func NewIPRotationHandler() *IPRotationHandler {
	return &IPRotationHandler{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Register mounts IP rotation routes on the given Echo group.
func (h *IPRotationHandler) Register(g *echo.Group) {
	ipr := g.Group("/ip-rotation")
	ipr.GET("/providers", h.ListProviders)
	ipr.POST("/rotate", h.Rotate)
	ipr.GET("/detect", h.Detect)
}

// ListProviders handles GET /api/v2/ip-rotation/providers — list supported providers.
func (h *IPRotationHandler) ListProviders(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"providers": supportedProviders})
}

// rotateRequest is the request body for POST /api/v2/ip-rotation/rotate.
type rotateRequest struct {
	Provider  string `json:"provider"`
	Token     string `json:"token"`
	CurrentIP string `json:"current_ip"`
	// Optional fields for provider-specific resource identification.
	DropletID int64  `json:"droplet_id,omitempty"` // DigitalOcean
	InstanceID string `json:"instance_id,omitempty"` // Vultr
	ServerID   int64  `json:"server_id,omitempty"`   // Hetzner
	LinodeID   int64  `json:"linode_id,omitempty"`   // Linode
}

// rotateResponse is the response for a successful IP rotation.
type rotateResponse struct {
	NewIP    string `json:"new_ip"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

// Rotate handles POST /api/v2/ip-rotation/rotate.
// Takes {provider, token, current_ip} and calls the provider API to allocate a new IP.
func (h *IPRotationHandler) Rotate(c echo.Context) error {
	var req rotateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Provider == "" || req.Token == "" || req.CurrentIP == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "provider, token, and current_ip are required")
	}

	ctx := c.Request().Context()

	var newIP string
	var err error

	switch req.Provider {
	case "digitalocean":
		newIP, err = h.rotateDigitalOcean(ctx, req)
	case "vultr":
		newIP, err = h.rotateVultr(ctx, req)
	case "hetzner":
		newIP, err = h.rotateHetzner(ctx, req)
	case "linode":
		newIP, err = h.rotateLinode(ctx, req)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported provider: "+req.Provider)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("rotation failed: %v", err))
	}

	return c.JSON(http.StatusOK, rotateResponse{
		NewIP:    newIP,
		Provider: req.Provider,
		Status:   "allocated",
	})
}

func (h *IPRotationHandler) rotateDigitalOcean(ctx interface{ Done() <-chan struct{} }, req rotateRequest) (string, error) {
	// POST /v2/floating_ips — create a reserved IP and assign to droplet.
	body := map[string]interface{}{
		"region":     "auto",
		"droplet_id": req.DropletID,
	}
	jsonBody, _ := json.Marshal(body)

	httpReq, err := http.NewRequest(http.MethodPost, "https://api.digitalocean.com/v2/floating_ips", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DigitalOcean API error %d: %s", resp.StatusCode, string(errBody))
	}

	var result struct {
		FloatingIP struct {
			IP string `json:"ip"`
		} `json:"floating_ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.FloatingIP.IP, nil
}

func (h *IPRotationHandler) rotateVultr(ctx interface{ Done() <-chan struct{} }, req rotateRequest) (string, error) {
	// POST /v2/reserved-ips — create and attach.
	body := map[string]interface{}{
		"region":      "auto",
		"ip_type":     "v4",
		"instance_id": req.InstanceID,
	}
	jsonBody, _ := json.Marshal(body)

	httpReq, err := http.NewRequest(http.MethodPost, "https://api.vultr.com/v2/reserved-ips", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Vultr API error %d: %s", resp.StatusCode, string(errBody))
	}

	var result struct {
		ReservedIP struct {
			Subnet string `json:"subnet"`
		} `json:"reserved_ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ReservedIP.Subnet, nil
}

func (h *IPRotationHandler) rotateHetzner(ctx interface{ Done() <-chan struct{} }, req rotateRequest) (string, error) {
	// POST /v1/floating_ips — create floating IP for server.
	body := map[string]interface{}{
		"type":   "ipv4",
		"server": req.ServerID,
	}
	jsonBody, _ := json.Marshal(body)

	httpReq, err := http.NewRequest(http.MethodPost, "https://api.hetzner.cloud/v1/floating_ips", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Hetzner API error %d: %s", resp.StatusCode, string(errBody))
	}

	var result struct {
		FloatingIP struct {
			IP string `json:"ip"`
		} `json:"floating_ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.FloatingIP.IP, nil
}

func (h *IPRotationHandler) rotateLinode(ctx interface{ Done() <-chan struct{} }, req rotateRequest) (string, error) {
	// POST /v4/networking/ips — allocate IP for linode.
	body := map[string]interface{}{
		"type":     "ipv4",
		"public":   true,
		"linode_id": req.LinodeID,
	}
	jsonBody, _ := json.Marshal(body)

	httpReq, err := http.NewRequest(http.MethodPost, "https://api.linode.com/v4/networking/ips", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Linode API error %d: %s", resp.StatusCode, string(errBody))
	}

	var result struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Address, nil
}

// Detect handles GET /api/v2/ip-rotation/detect.
// Auto-detects the cloud provider from the server's public IP using BGP/ASN lookup.
func (h *IPRotationHandler) Detect(c echo.Context) error {
	ctx := c.Request().Context()

	// Use ip-api.com to detect ASN information for the server's outbound IP.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://ip-api.com/json/?fields=query,as,org,isp", nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create detection request")
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "IP detection service unreachable")
	}
	defer resp.Body.Close()

	var info struct {
		Query string `json:"query"`
		AS    string `json:"as"`
		Org   string `json:"org"`
		ISP   string `json:"isp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse IP info")
	}

	// Match ASN/org to known providers.
	provider := detectProvider(info.AS, info.Org, info.ISP)

	return c.JSON(http.StatusOK, echo.Map{
		"ip":       info.Query,
		"as":       info.AS,
		"org":      info.Org,
		"provider": provider,
	})
}

// detectProvider maps ASN/org strings to known cloud provider IDs.
func detectProvider(as, org, isp string) string {
	combined := as + " " + org + " " + isp

	// Simple substring matching for known providers.
	patterns := map[string][]string{
		"digitalocean": {"DIGITALOCEAN", "DigitalOcean", "DO-13"},
		"vultr":        {"VULTR", "Vultr", "AS-CHOOPA", "Choopa"},
		"hetzner":      {"HETZNER", "Hetzner"},
		"linode":       {"LINODE", "Linode", "Akamai Connected Cloud"},
	}

	for provider, keywords := range patterns {
		for _, kw := range keywords {
			if containsCI(combined, kw) {
				return provider
			}
		}
	}
	return "unknown"
}

// containsCI checks if s contains substr (case-insensitive).
func containsCI(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			pc := substr[j]
			// Lowercase both.
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if pc >= 'A' && pc <= 'Z' {
				pc += 32
			}
			if sc != pc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

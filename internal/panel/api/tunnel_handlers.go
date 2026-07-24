package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// TunnelBackend represents a supported tunnel backend type.
type TunnelBackend string

const (
	TunnelBackendBackhaul TunnelBackend = "backhaul"
	TunnelBackendRathole  TunnelBackend = "rathole"
	TunnelBackendWstunnel TunnelBackend = "wstunnel"
)

// TunnelTransport represents the transport layer for tunnels.
type TunnelTransport string

const (
	TransportTCP       TunnelTransport = "tcp"
	TransportWS        TunnelTransport = "ws"
	TransportWSS       TunnelTransport = "wss"
	TransportTCPMux    TunnelTransport = "tcpmux"
	TransportWSMux     TunnelTransport = "wsmux"
	TransportWSSMux    TunnelTransport = "wssmux"
)

// TunnelConfig represents a configured tunnel entry.
type TunnelConfig struct {
	ID        string          `json:"id"`
	Backend   TunnelBackend   `json:"backend"`
	Port      int             `json:"port"`
	Transport TunnelTransport `json:"transport"`
	Secret    string          `json:"secret"`
	NodeIP    string          `json:"node_ip"`
	IranIP    string          `json:"iran_ip"`
	Enabled   bool            `json:"enabled"`
	Status    string          `json:"status"` // online, offline, unknown
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TunnelHandlers serves Iran Bridge Tunnel management endpoints.
type TunnelHandlers struct {
	mu      sync.RWMutex
	tunnels map[string]*TunnelConfig
	// bridgeScriptURL is the public URL where vortex-bridge.sh is hosted.
	BridgeScriptURL string
}

// NewTunnelHandlers creates a new TunnelHandlers instance.
func NewTunnelHandlers(bridgeScriptURL string) *TunnelHandlers {
	return &TunnelHandlers{
		tunnels:         make(map[string]*TunnelConfig),
		BridgeScriptURL: bridgeScriptURL,
	}
}

// Register mounts all tunnel routes on the provided authenticated group.
func (h *TunnelHandlers) Register(g *echo.Group) {
	tunnels := g.Group("/tunnels")
	tunnels.GET("", h.ListTunnels)
	tunnels.POST("", h.CreateTunnel)
	tunnels.DELETE("/:id", h.DeleteTunnel)
	tunnels.GET("/bridge-command/:id", h.GetBridgeCommand)
	tunnels.POST("/:id/status", h.CheckStatus)
}

// ─── Request/Response Types ──────────────────────────────────────────────────

type createTunnelRequest struct {
	Backend   TunnelBackend   `json:"backend"`
	Port      int             `json:"port"`
	Transport TunnelTransport `json:"transport"`
	Secret    string          `json:"secret"`
	NodeIP    string          `json:"node_ip"`
	IranIP    string          `json:"iran_ip"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// ListTunnels returns all configured tunnels.
func (h *TunnelHandlers) ListTunnels(c echo.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	list := make([]*TunnelConfig, 0, len(h.tunnels))
	for _, t := range h.tunnels {
		list = append(list, t)
	}
	return c.JSON(http.StatusOK, echo.Map{"tunnels": list})
}

// CreateTunnel creates or updates a tunnel configuration.
func (h *TunnelHandlers) CreateTunnel(c echo.Context) error {
	var req createTunnelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	// Validate required fields.
	if req.Backend == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "backend is required")
	}
	if req.Port < 1 || req.Port > 65535 {
		return echo.NewHTTPError(http.StatusBadRequest, "port must be 1-65535")
	}
	switch req.Backend {
	case TunnelBackendBackhaul, TunnelBackendRathole, TunnelBackendWstunnel:
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported backend: "+string(req.Backend))
	}

	// Default transport if not specified.
	if req.Transport == "" {
		req.Transport = TransportTCP
	}

	now := time.Now()
	tunnel := &TunnelConfig{
		ID:        uuid.New().String(),
		Backend:   req.Backend,
		Port:      req.Port,
		Transport: req.Transport,
		Secret:    req.Secret,
		NodeIP:    req.NodeIP,
		IranIP:    req.IranIP,
		Enabled:   true,
		Status:    "unknown",
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.mu.Lock()
	h.tunnels[tunnel.ID] = tunnel
	h.mu.Unlock()

	return c.JSON(http.StatusCreated, echo.Map{"tunnel": tunnel})
}

// DeleteTunnel removes a tunnel by ID.
func (h *TunnelHandlers) DeleteTunnel(c echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.tunnels[id]; !ok {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}
	delete(h.tunnels, id)
	return c.NoContent(http.StatusNoContent)
}

// GetBridgeCommand generates the one-liner bash command for installing the
// bridge on an Iran VPS.
func (h *TunnelHandlers) GetBridgeCommand(c echo.Context) error {
	id := c.Param("id")

	h.mu.RLock()
	tunnel, ok := h.tunnels[id]
	h.mu.RUnlock()
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}

	// Generate backend-specific config.
	config := h.generateConfig(tunnel)
	configB64 := base64.StdEncoding.EncodeToString([]byte(config))

	// Build the one-liner.
	scriptURL := h.BridgeScriptURL
	if scriptURL == "" {
		// Derive from request host.
		scheme := "https"
		if c.Request().TLS == nil {
			scheme = "http"
		}
		scriptURL = fmt.Sprintf("%s://%s/scripts/vortex-bridge.sh", scheme, c.Request().Host)
	}

	command := fmt.Sprintf(
		`bash <(curl -fsSL %s) --backend %s --config-b64 %s --port %d`,
		scriptURL, tunnel.Backend, configB64, tunnel.Port,
	)

	return c.JSON(http.StatusOK, echo.Map{
		"command": command,
		"config":  config,
	})
}

// CheckStatus updates/checks the tunnel status.
func (h *TunnelHandlers) CheckStatus(c echo.Context) error {
	id := c.Param("id")

	h.mu.Lock()
	defer h.mu.Unlock()

	tunnel, ok := h.tunnels[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "tunnel not found")
	}

	var req updateStatusRequest
	if err := c.Bind(&req); err == nil && req.Status != "" {
		tunnel.Status = req.Status
		tunnel.UpdatedAt = time.Now()
	}

	return c.JSON(http.StatusOK, echo.Map{"tunnel": tunnel})
}

// ─── Config Generation (private) ────────────────────────────────────────────

func (h *TunnelHandlers) generateConfig(t *TunnelConfig) string {
	switch t.Backend {
	case TunnelBackendBackhaul:
		return h.generateBackhaulConfig(t)
	case TunnelBackendRathole:
		return h.generateRatholeConfig(t)
	case TunnelBackendWstunnel:
		return h.generateWstunnelConfig(t)
	default:
		return ""
	}
}

func (h *TunnelHandlers) generateBackhaulConfig(t *TunnelConfig) string {
	return fmt.Sprintf(`[server]
bind_addr = "0.0.0.0:%d"
transport = "%s"
token = "%s"
keepalive_period = 75
nodelay = true
heartbeat = 40
channel_size = 2048
mux_con = 8
mux_version = 1
mux_framesize = 32768

[server.sniffer]
enabled = false
`, t.Port, t.Transport, t.Secret)
}

func (h *TunnelHandlers) generateRatholeConfig(t *TunnelConfig) string {
	return fmt.Sprintf(`[server]
bind_addr = "0.0.0.0:%d"

[server.transport]
type = "%s"

[server.transport.tcp]
nodelay = true
keepalive_secs = 20
keepalive_interval = 8

[server.services.vortex]
type = "%s"
token = "%s"
bind_addr = "0.0.0.0:%d"
`, t.Port, t.Transport, t.Transport, t.Secret, t.Port)
}

func (h *TunnelHandlers) generateWstunnelConfig(t *TunnelConfig) string {
	proto := "ws"
	if t.Transport == TransportWSS || t.Transport == TransportWSSMux {
		proto = "wss"
	}
	return fmt.Sprintf(`# wstunnel server config
# Run: wstunnel server %s://0.0.0.0:%d --restrict-to 127.0.0.1:443
[server]
bind = "%s://0.0.0.0:%d"
restrict_to = "127.0.0.1:443"
secret = "%s"
`, proto, t.Port, proto, t.Port, t.Secret)
}

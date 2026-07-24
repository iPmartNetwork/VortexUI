package api

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// AWGConfig holds AmneziaWG obfuscation (junk) parameters.
type AWGConfig struct {
	Enabled bool   `json:"enabled"`
	Preset  string `json:"preset"` // "light", "medium", "heavy", or "custom"
	Jc      int    `json:"jc"`
	Jmin    int    `json:"jmin"`
	Jmax    int    `json:"jmax"`
	S1      int    `json:"s1"`
	S2      int    `json:"s2"`
	H1      int    `json:"h1"`
	H2      int    `json:"h2"`
	H3      int    `json:"h3"`
	H4      int    `json:"h4"`
}

// AWGHandler manages AmneziaWG obfuscated WireGuard settings.
type AWGHandler struct {
	mu     sync.RWMutex
	config AWGConfig
}

// NewAWGHandler creates an AWGHandler with default (disabled) config.
func NewAWGHandler() *AWGHandler {
	return &AWGHandler{
		config: AWGConfig{Preset: "light"},
	}
}

// Register mounts AWG routes on the given Echo group.
func (h *AWGHandler) Register(g *echo.Group) {
	awg := g.Group("/awg")
	awg.GET("/config", h.GetConfig)
	awg.PUT("/config", h.UpdateConfig)
}

// awgPresets defines built-in AmneziaWG junk parameter presets.
var awgPresets = map[string]AWGConfig{
	"light": {
		Preset: "light",
		Jc:     3, Jmin: 50, Jmax: 1000,
		S1: 15, S2: 30,
		H1: 1, H2: 2, H3: 3, H4: 4,
	},
	"medium": {
		Preset: "medium",
		Jc:     5, Jmin: 100, Jmax: 1500,
		S1: 30, S2: 60,
		H1: 5, H2: 10, H3: 15, H4: 20,
	},
	"heavy": {
		Preset: "heavy",
		Jc:     10, Jmin: 200, Jmax: 2000,
		S1: 50, S2: 100,
		H1: 10, H2: 20, H3: 30, H4: 40,
	},
}

// GetConfig handles GET /api/v2/awg/config — returns current AWG junk settings.
func (h *AWGHandler) GetConfig(c echo.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return c.JSON(http.StatusOK, echo.Map{
		"config":  h.config,
		"presets": awgPresets,
	})
}

// updateAWGRequest is the request body for PUT /api/v2/awg/config.
type updateAWGRequest struct {
	Enabled bool   `json:"enabled"`
	Preset  string `json:"preset"` // "light", "medium", "heavy", "custom"
	Jc      *int   `json:"jc"`
	Jmin    *int   `json:"jmin"`
	Jmax    *int   `json:"jmax"`
	S1      *int   `json:"s1"`
	S2      *int   `json:"s2"`
	H1      *int   `json:"h1"`
	H2      *int   `json:"h2"`
	H3      *int   `json:"h3"`
	H4      *int   `json:"h4"`
}

// UpdateConfig handles PUT /api/v2/awg/config — update AWG settings (preset or manual).
func (h *AWGHandler) UpdateConfig(c echo.Context) error {
	var req updateAWGRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.Enabled = req.Enabled

	// If a known preset is selected, apply its values.
	if preset, ok := awgPresets[req.Preset]; ok {
		h.config = preset
		h.config.Enabled = req.Enabled
		return c.JSON(http.StatusOK, echo.Map{"config": h.config})
	}

	// Custom mode: apply individual fields if provided.
	h.config.Preset = "custom"
	if req.Jc != nil {
		h.config.Jc = *req.Jc
	}
	if req.Jmin != nil {
		h.config.Jmin = *req.Jmin
	}
	if req.Jmax != nil {
		h.config.Jmax = *req.Jmax
	}
	if req.S1 != nil {
		h.config.S1 = *req.S1
	}
	if req.S2 != nil {
		h.config.S2 = *req.S2
	}
	if req.H1 != nil {
		h.config.H1 = *req.H1
	}
	if req.H2 != nil {
		h.config.H2 = *req.H2
	}
	if req.H3 != nil {
		h.config.H3 = *req.H3
	}
	if req.H4 != nil {
		h.config.H4 = *req.H4
	}

	return c.JSON(http.StatusOK, echo.Map{"config": h.config})
}

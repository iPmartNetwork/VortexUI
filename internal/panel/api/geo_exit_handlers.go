package api

import (
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/labstack/echo/v4"
)

// GeoExit represents a configured Tor or Psiphon geographic exit.
type GeoExit struct {
	ID      int64  `json:"id"`
	Service string `json:"service"` // "tor" or "psiphon"
	Country string `json:"country"` // 2-letter ISO country code
}

// GeoExitHandler manages Psiphon/Tor geographic exit configurations.
type GeoExitHandler struct {
	mu    sync.RWMutex
	exits []GeoExit
	seq   atomic.Int64
}

// NewGeoExitHandler creates a GeoExitHandler with an empty exit list.
func NewGeoExitHandler() *GeoExitHandler {
	return &GeoExitHandler{}
}

// Register mounts geo-exit routes on the given Echo group.
func (h *GeoExitHandler) Register(g *echo.Group) {
	ge := g.Group("/geo-exits")
	ge.GET("", h.List)
	ge.POST("", h.Create)
	ge.DELETE("/:id", h.Delete)
	ge.GET("/regions/:service", h.ListRegions)
}

// Supported country codes for Tor exit nodes.
var torCountries = []string{
	"at", "be", "bg", "ca", "ch", "cz", "de", "dk", "ee", "es",
	"fi", "fr", "gb", "hu", "ie", "is", "it", "lt", "lu", "lv",
	"nl", "no", "pl", "pt", "ro", "rs", "se", "si", "sk", "tr",
	"ua", "us",
}

// Supported country codes for Psiphon exit nodes.
var psiphonCountries = []string{
	"at", "be", "bg", "ca", "ch", "cz", "de", "dk", "es", "fi",
	"fr", "gb", "hu", "ie", "it", "nl", "no", "pl", "ro", "rs",
	"se", "sk", "tr", "ua", "us",
}

// List handles GET /api/v2/geo-exits — list all configured geo exits.
func (h *GeoExitHandler) List(c echo.Context) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	exits := h.exits
	if exits == nil {
		exits = []GeoExit{}
	}
	return c.JSON(http.StatusOK, echo.Map{"exits": exits})
}

// createGeoExitRequest is the request body for POST /api/v2/geo-exits.
type createGeoExitRequest struct {
	Service string `json:"service"` // "tor" or "psiphon"
	Country string `json:"country"` // 2-letter ISO code
}

// Create handles POST /api/v2/geo-exits — add a new geo exit.
func (h *GeoExitHandler) Create(c echo.Context) error {
	var req createGeoExitRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Service != "tor" && req.Service != "psiphon" {
		return echo.NewHTTPError(http.StatusBadRequest, "service must be 'tor' or 'psiphon'")
	}
	if len(req.Country) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "country must be a 2-letter ISO code")
	}

	// Validate the country is supported for the given service.
	var supported []string
	if req.Service == "tor" {
		supported = torCountries
	} else {
		supported = psiphonCountries
	}
	valid := false
	for _, cc := range supported {
		if cc == req.Country {
			valid = true
			break
		}
	}
	if !valid {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported country for "+req.Service)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check for duplicate.
	for _, e := range h.exits {
		if e.Service == req.Service && e.Country == req.Country {
			return echo.NewHTTPError(http.StatusConflict, "geo exit already exists")
		}
	}

	exit := GeoExit{
		ID:      h.seq.Add(1),
		Service: req.Service,
		Country: req.Country,
	}
	h.exits = append(h.exits, exit)

	return c.JSON(http.StatusCreated, echo.Map{"exit": exit})
}

// Delete handles DELETE /api/v2/geo-exits/:id — remove a geo exit.
func (h *GeoExitHandler) Delete(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}

	var id int64
	for _, ch := range idParam {
		if ch < '0' || ch > '9' {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
		}
		id = id*10 + int64(ch-'0')
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for i, e := range h.exits {
		if e.ID == id {
			h.exits = append(h.exits[:i], h.exits[i+1:]...)
			return c.JSON(http.StatusOK, echo.Map{"deleted": true})
		}
	}

	return echo.NewHTTPError(http.StatusNotFound, "geo exit not found")
}

// ListRegions handles GET /api/v2/geo-exits/regions/:service — list available regions.
func (h *GeoExitHandler) ListRegions(c echo.Context) error {
	svc := c.Param("service")

	var countries []string
	switch svc {
	case "tor":
		countries = make([]string, len(torCountries))
		copy(countries, torCountries)
	case "psiphon":
		countries = make([]string, len(psiphonCountries))
		copy(countries, psiphonCountries)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "service must be 'tor' or 'psiphon'")
	}

	sort.Strings(countries)
	return c.JSON(http.StatusOK, echo.Map{"service": svc, "countries": countries})
}

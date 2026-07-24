package api

import (
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// ConfigManagementHandler serves config validation, versioning, diff, rollback,
// import/export, and auto-update endpoints.
type ConfigManagementHandler struct {
	svc *service.ConfigManagementService
}

// NewConfigManagementHandler creates the handler with its service dependency.
func NewConfigManagementHandler(svc *service.ConfigManagementService) *ConfigManagementHandler {
	return &ConfigManagementHandler{svc: svc}
}

// Register mounts all config management routes on the given Echo group.
func (h *ConfigManagementHandler) Register(g *echo.Group) {
	inbounds := g.Group("/inbounds/:id")
	inbounds.POST("/validate", h.ValidateConfig)
	inbounds.GET("/defaults", h.GetDefaults)
	inbounds.GET("/diff", h.GetDiff)
	inbounds.POST("/rollback", h.Rollback)
	inbounds.GET("/export", h.ExportConfig)
	inbounds.GET("/versions", h.ListVersions)

	g.POST("/inbounds/import", h.ImportConfig)
	g.POST("/nodes/:id/auto-update", h.AutoUpdate)
}

// --- request/response types ---

type validateConfigRequest struct {
	Protocol string         `json:"protocol"`
	Network  string         `json:"network"`
	Security string         `json:"security"`
	Config   map[string]any `json:"config"`
}

type validateConfigResponse struct {
	Valid  bool                           `json:"valid"`
	Errors []domain.ConfigValidationError `json:"errors,omitempty"`
}

type getDefaultsParams struct {
	Protocol string `query:"protocol"`
	Network  string `query:"network"`
	Security string `query:"security"`
}

type diffParams struct {
	OldVersion int `query:"old_version"`
	NewVersion int `query:"new_version"`
}

type rollbackRequest struct {
	Version int    `json:"version"`
	Comment string `json:"comment,omitempty"`
}

type importConfigRequest struct {
	InboundID string `json:"inbound_id"`
	Protocol  string `json:"protocol"`
	Network   string `json:"network"`
	Security  string `json:"security"`
}

type autoUpdateRequest struct {
	CoreType    string `json:"core_type"`
	DownloadURL string `json:"download_url,omitempty"`
}

// --- handlers ---

// ValidateConfig handles POST /api/v2/inbounds/:id/validate.
func (h *ConfigManagementHandler) ValidateConfig(c echo.Context) error {
	var req validateConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	errs := h.svc.ValidateConfig(
		domain.Protocol(req.Protocol),
		req.Network,
		domain.Security(req.Security),
		req.Config,
	)

	return c.JSON(http.StatusOK, validateConfigResponse{
		Valid:  len(errs) == 0,
		Errors: errs,
	})
}

// GetDefaults handles GET /api/v2/inbounds/:id/defaults?protocol=&network=&security=.
func (h *ConfigManagementHandler) GetDefaults(c echo.Context) error {
	var params getDefaultsParams
	if err := c.Bind(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query params")
	}

	defaults := h.svc.GetDefaults(
		domain.Protocol(params.Protocol),
		params.Network,
		domain.Security(params.Security),
	)

	return c.JSON(http.StatusOK, defaults)
}

// GetDiff handles GET /api/v2/inbounds/:id/diff?old_version=&new_version=.
func (h *ConfigManagementHandler) GetDiff(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound ID")
	}

	oldVer, _ := strconv.Atoi(c.QueryParam("old_version"))
	newVer, _ := strconv.Atoi(c.QueryParam("new_version"))
	if newVer == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "new_version is required")
	}

	diff, err := h.svc.Diff(c.Request().Context(), inboundID, oldVer, newVer)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, diff)
}

// Rollback handles POST /api/v2/inbounds/:id/rollback.
func (h *ConfigManagementHandler) Rollback(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound ID")
	}

	var req rollbackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Version < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "version must be >= 1")
	}

	// Extract admin ID from context (set by auth middleware)
	adminID := extractAdminIDPtr(c)

	version, err := h.svc.Rollback(c.Request().Context(), inboundID, req.Version, adminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, version)
}

// ListVersions handles GET /api/v2/inbounds/:id/versions.
func (h *ConfigManagementHandler) ListVersions(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound ID")
	}

	versions, err := h.svc.ListVersions(c.Request().Context(), inboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, versions)
}

// ExportConfig handles GET /api/v2/inbounds/:id/export.
func (h *ConfigManagementHandler) ExportConfig(c echo.Context) error {
	inboundID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound ID")
	}

	data, err := h.svc.ExportConfig(c.Request().Context(), inboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename=config-export.json")
	return c.Blob(http.StatusOK, "application/json", data)
}

// ImportConfig handles POST /api/v2/inbounds/import.
func (h *ConfigManagementHandler) ImportConfig(c echo.Context) error {
	var meta importConfigRequest
	if err := c.Bind(&meta); err != nil {
		// Try reading raw body for file upload
		body, readErr := io.ReadAll(c.Request().Body)
		if readErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		// Assume the entire body is the config JSON
		return h.handleRawImport(c, body)
	}

	inboundID, err := uuid.Parse(meta.InboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot read body")
	}

	adminID := extractAdminIDPtr(c)

	version, validErrs, err := h.svc.ImportConfig(
		c.Request().Context(),
		inboundID,
		domain.Protocol(meta.Protocol),
		meta.Network,
		domain.Security(meta.Security),
		body,
		adminID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(validErrs) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, validateConfigResponse{
			Valid:  false,
			Errors: validErrs,
		})
	}

	return c.JSON(http.StatusOK, version)
}

// AutoUpdate handles POST /api/v2/nodes/:id/auto-update.
func (h *ConfigManagementHandler) AutoUpdate(c echo.Context) error {
	nodeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node ID")
	}

	var req autoUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.CoreType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "core_type is required")
	}

	result, err := h.svc.AutoUpdate(c.Request().Context(), service.AutoUpdateRequest{
		NodeID:      nodeID,
		CoreType:    req.CoreType,
		DownloadURL: req.DownloadURL,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	status := http.StatusOK
	if !result.Success {
		status = http.StatusBadGateway
	}
	return c.JSON(status, result)
}

// --- helpers ---

func (h *ConfigManagementHandler) handleRawImport(c echo.Context, body []byte) error {
	// Fallback: body is the full import JSON with embedded metadata
	return c.JSON(http.StatusBadRequest, map[string]string{
		"error": "request must include inbound_id, protocol, network, and security fields",
	})
}

func extractAdminIDPtr(c echo.Context) *uuid.UUID {
	if val := c.Get("admin_id"); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return &v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return &id
			}
		}
	}
	return nil
}

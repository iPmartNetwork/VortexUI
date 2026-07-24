package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// SubSettingsHandlers serves the panel-configurable subscription settings.
type SubSettingsHandlers struct {
	Svc *service.SubSettingsService
}

// GetSubSettings returns the current subscription settings.
func (h *SubSettingsHandlers) GetSubSettings(c echo.Context) error {
	cfg, err := h.Svc.GetConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

// updateSubSettingsRequest allows partial updates — omitted fields retain
// current values so legacy clients that only send update_interval don't
// accidentally clear the template fields.
type updateSubSettingsRequest struct {
	UpdateInterval       *int    `json:"update_interval"`
	ProfileTitleTemplate *string `json:"profile_title_template"`
	RemarkTemplate       *string `json:"remark_template"`
	AddressTemplate      *string `json:"address_template"`
}

// UpdateSubSettings saves the subscription settings. It merges the request with
// existing settings so that only supplied fields are changed.
func (h *SubSettingsHandlers) UpdateSubSettings(c echo.Context) error {
	var req updateSubSettingsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	// Load current settings as base for merge.
	current, err := h.Svc.GetConfig(c.Request().Context())
	if err != nil {
		def := domain.DefaultSubSettings()
		current = &def
	}

	// Apply only provided fields.
	if req.UpdateInterval != nil {
		current.UpdateInterval = *req.UpdateInterval
	}
	if req.ProfileTitleTemplate != nil {
		current.ProfileTitleTemplate = *req.ProfileTitleTemplate
	}
	if req.RemarkTemplate != nil {
		current.RemarkTemplate = *req.RemarkTemplate
	}
	if req.AddressTemplate != nil {
		current.AddressTemplate = *req.AddressTemplate
	}

	if err := h.Svc.UpdateConfig(c.Request().Context(), current); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": current})
}

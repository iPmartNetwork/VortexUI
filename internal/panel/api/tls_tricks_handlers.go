package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// TLSTricksHandlers serves Fragment/TLS Tricks endpoints.
type TLSTricksHandlers struct {
	Tricks *service.TLSTricksService
}

func (h *TLSTricksHandlers) ListProfiles(c echo.Context) error {
	profiles, err := h.Tricks.ListProfiles(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"profiles": profiles})
}

func (h *TLSTricksHandlers) CreateProfile(c echo.Context) error {
	var p domain.TLSTrickProfile
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	result, err := h.Tricks.CreateProfile(c.Request().Context(), &p)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"profile": result})
}

type createFromPresetRequest struct {
	ISP string `json:"isp"`
}

func (h *TLSTricksHandlers) CreateFromPreset(c echo.Context) error {
	var req createFromPresetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p, err := h.Tricks.CreateFromPreset(c.Request().Context(), domain.ISPPreset(req.ISP))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"profile": p})
}

func (h *TLSTricksHandlers) UpdateProfile(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p domain.TLSTrickProfile
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p.ID = id
	result, err := h.Tricks.UpdateProfile(c.Request().Context(), &p)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"profile": result})
}

func (h *TLSTricksHandlers) DeleteProfile(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Tricks.DeleteProfile(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *TLSTricksHandlers) GetPresets(c echo.Context) error {
	presets := h.Tricks.GetAvailablePresets()
	var out []echo.Map
	for _, isp := range presets {
		p := domain.ISPPresetDefaults(isp)
		out = append(out, echo.Map{"isp": isp, "name": p.Name, "defaults": p})
	}
	return c.JSON(http.StatusOK, echo.Map{"presets": out})
}

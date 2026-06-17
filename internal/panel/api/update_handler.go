package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/updater"
)

// UpdateChecker is optionally wired to check for new releases.
type UpdateChecker interface {
	Check(ctx echo.Context) (*updater.CheckResult, error)
}

// CheckUpdate returns whether a new version is available.
func (h *Handlers) CheckUpdate(c echo.Context) error {
	if h.Updater == nil {
		return c.JSON(http.StatusOK, echo.Map{"available": false, "current": "unknown"})
	}
	result, err := h.Updater.Check(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"available": false, "error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

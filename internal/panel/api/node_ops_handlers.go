package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
)

func enrichNode(n *domain.Node, h *Handlers) {
	if n == nil || h.Hub == nil {
		return
	}
	st, hl, d, ok := h.Hub.Live(n.ID)
	if !ok {
		return
	}
	n.Status = st
	n.Health = hl
	dd := d
	n.Diagnostics = &dd
}

// GetNodeEnrollment returns the mTLS enrollment bundle for adding a remote node.
func (h *Handlers) GetNodeEnrollment(c echo.Context) error {
	if h.Enrollment == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "enrollment not configured")
	}
	b, err := h.Enrollment.Bundle()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, b)
}

// TestNodeConnection performs an on-demand dial + health check for one node.
func (h *Handlers) TestNodeConnection(c echo.Context) error {
	if h.Hub == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "hub not configured")
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	n, err := h.Nodes.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}
	d := h.Hub.TestConnect(c.Request().Context(), n)
	return c.JSON(http.StatusOK, echo.Map{"diagnostics": d})
}

// GetNodeDebugBundle returns a copy-friendly debug summary for support.
func (h *Handlers) GetNodeDebugBundle(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	n, err := h.Nodes.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}
	enrichNode(n, h)
	var caFP string
	if h.Enrollment != nil {
		if b, berr := h.Enrollment.Bundle(); berr == nil {
			caFP = b.CAFingerprint
		}
	}
	diag := "unknown"
	if n.Diagnostics != nil {
		diag = string(n.Diagnostics.Code)
	}
	return c.JSON(http.StatusOK, echo.Map{
		"node":          n.Name,
		"address":       n.Address,
		"status":        n.Status,
		"diagnostics":   n.Diagnostics,
		"panel_ca_fp":   caFP,
		"last_seen":     n.LastSeen,
		"core_running":  n.Health.CoreRunning,
		"agent_version": n.AgentVer,
		"core_version":  n.CoreVer,
		"debug_text":    formatNodeDebug(n, caFP, diag),
	})
}

func formatNodeDebug(n *domain.Node, caFP, diag string) string {
	msg := ""
	if n.Diagnostics != nil {
		msg = n.Diagnostics.Message
	}
	return "VortexUI node debug\n" +
		"  name: " + n.Name + "\n" +
		"  address: " + n.Address + "\n" +
		"  status: " + string(n.Status) + "\n" +
		"  diagnostics: " + diag + "\n" +
		"  message: " + msg + "\n" +
		"  panel CA fingerprint: " + caFP + "\n" +
		"  core_running: " + boolStr(n.Health.CoreRunning) + "\n"
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

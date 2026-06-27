package api

import (
	"net/http"
	"time"

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
	n.EnrollmentPhase = enrollmentPhase(n)
}

func enrollmentPhase(n *domain.Node) domain.NodeEnrollmentPhase {
	if n.Diagnostics == nil {
		return domain.NodePhasePending
	}
	switch n.Diagnostics.Code {
	case domain.NodeDiagOK:
		if n.LastSeen != nil && time.Since(*n.LastSeen) <= 90*time.Second {
			return domain.NodePhaseSynced
		}
		return domain.NodePhaseConnected
	case domain.NodeDiagCoreDown:
		return domain.NodePhaseConnected
	case domain.NodeDiagMTLS:
		if n.Diagnostics.NetworkReachable {
			return domain.NodePhasePending
		}
	}
	if n.Diagnostics.CAMatch && n.Diagnostics.NetworkReachable {
		return domain.NodePhaseConnected
	}
	if n.Status == domain.NodeConnected {
		return domain.NodePhaseConnected
	}
	return domain.NodePhasePending
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
	var caFP string
	if h.Enrollment != nil {
		if b, berr := h.Enrollment.Bundle(); berr == nil {
			caFP = b.CAFingerprint
		}
	}
	phase := domain.NodePhasePending
	if d.Code == domain.NodeDiagOK {
		phase = domain.NodePhaseSynced
	} else if d.CAMatch || d.Code == domain.NodeDiagCoreDown {
		phase = domain.NodePhaseConnected
	}
	return c.JSON(http.StatusOK, echo.Map{
		"diagnostics":         d,
		"panel_ca_fingerprint": caFP,
		"ca_match":            d.CAMatch,
		"enrollment_phase":    phase,
	})
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
	netOK := false
	caMatch := false
	if n.Diagnostics != nil {
		netOK = n.Diagnostics.NetworkReachable
		caMatch = n.Diagnostics.CAMatch
	}
	return c.JSON(http.StatusOK, echo.Map{
		"node":               n.Name,
		"address":            n.Address,
		"status":             n.Status,
		"enrollment_phase":   n.EnrollmentPhase,
		"diagnostics":        n.Diagnostics,
		"panel_ca_fp":        caFP,
		"last_seen":          n.LastSeen,
		"core_running":       n.Health.CoreRunning,
		"agent_version":      n.AgentVer,
		"core_version":       n.CoreVer,
		"debug_text":         formatNodeDebug(n, caFP, diag, netOK, caMatch),
	})
}

func formatNodeDebug(n *domain.Node, caFP, diag string, netOK, caMatch bool) string {
	msg := ""
	if n.Diagnostics != nil {
		msg = n.Diagnostics.Message
	}
	return "VortexUI node debug\n" +
		"  name: " + n.Name + "\n" +
		"  address: " + n.Address + "\n" +
		"  status: " + string(n.Status) + "\n" +
		"  enrollment: " + string(n.EnrollmentPhase) + "\n" +
		"  diagnostics: " + diag + "\n" +
		"  message: " + msg + "\n" +
		"  network_reachable: " + boolStr(netOK) + "\n" +
		"  ca_match: " + boolStr(caMatch) + "\n" +
		"  panel CA fingerprint: " + caFP + "\n" +
		"  core_running: " + boolStr(n.Health.CoreRunning) + "\n"
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

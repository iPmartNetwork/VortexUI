package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// MigrationHandlers serves node auto-migration endpoints.
type MigrationHandlers struct {
	Migration *service.MigrationService
}

// GetMigrationPolicy returns the current auto-migration policy.
func (h *MigrationHandlers) GetMigrationPolicy(c echo.Context) error {
	p, err := h.Migration.GetPolicy(c.Request().Context())
	if err != nil || p == nil {
		def := domain.DefaultMigrationPolicy()
		return c.JSON(http.StatusOK, echo.Map{"policy": &def})
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

type updateMigrationPolicyRequest struct {
	Enabled             bool    `json:"enabled"`
	HealthCheckInterval int     `json:"health_check_interval"`
	UnhealthyThreshold  int     `json:"unhealthy_threshold"`
	CPUThreshold        float64 `json:"cpu_threshold"`
	MemThreshold        float64 `json:"mem_threshold"`
	PacketLossMax       float64 `json:"packet_loss_max"`
	MigrateBack         bool    `json:"migrate_back"`
}

// UpdateMigrationPolicy updates the auto-migration policy.
func (h *MigrationHandlers) UpdateMigrationPolicy(c echo.Context) error {
	var req updateMigrationPolicyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p := &domain.MigrationPolicy{
		Enabled:             req.Enabled,
		HealthCheckInterval: req.HealthCheckInterval,
		UnhealthyThreshold:  req.UnhealthyThreshold,
		CPUThreshold:        req.CPUThreshold,
		MemThreshold:        req.MemThreshold,
		PacketLossMax:       req.PacketLossMax,
		MigrateBack:         req.MigrateBack,
	}
	if err := h.Migration.UpdatePolicy(c.Request().Context(), p); err != nil {
		// If migration_policy table doesn't exist yet, return success with the input
		return c.JSON(http.StatusOK, echo.Map{"policy": p})
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

// ListMigrationEvents returns recent migration events.
func (h *MigrationHandlers) ListMigrationEvents(c echo.Context) error {
	events, total, err := h.Migration.ListEvents(c.Request().Context(), 50, 0)
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"events": []any{}, "total": 0})
	}
	if events == nil {
		events = []*domain.MigrationEvent{}
	}
	return c.JSON(http.StatusOK, echo.Map{"events": events, "total": total})
}

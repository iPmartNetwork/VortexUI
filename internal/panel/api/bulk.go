package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// BulkHandler serves bulk user operation endpoints.
type BulkHandler struct {
	svc *service.BulkService
}

// NewBulkHandler creates a new BulkHandler with the given service dependency.
func NewBulkHandler(svc *service.BulkService) *BulkHandler {
	return &BulkHandler{svc: svc}
}

// Register mounts all bulk operation routes on the given Echo group.
func (h *BulkHandler) Register(g *echo.Group) {
	bulk := g.Group("/bulk")
	bulk.POST("/preview", h.Preview)
	bulk.POST("/execute", h.Execute)
	bulk.GET("/history", h.History)
}

type bulkRequest struct {
	OperationType string         `json:"operation_type"`
	Parameters    map[string]any `json:"parameters"`
	Filters       domain.BulkFilter `json:"filters"`
}

// Preview handles POST /api/v2/bulk/preview.
// Performs a dry-run of the bulk operation, returning affected count and summary.
func (h *BulkHandler) Preview(c echo.Context) error {
	var req bulkRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.OperationType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "operation_type is required")
	}

	result, err := h.svc.Preview(
		c.Request().Context(),
		domain.BulkOperationType(req.OperationType),
		req.Parameters,
		req.Filters,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"preview": result})
}

// Execute handles POST /api/v2/bulk/execute.
// Applies the bulk operation and records it in history.
func (h *BulkHandler) Execute(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req bulkRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.OperationType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "operation_type is required")
	}

	result, err := h.svc.Execute(
		c.Request().Context(),
		claims.AdminID,
		domain.BulkOperationType(req.OperationType),
		req.Parameters,
		req.Filters,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"operation": result})
}

// History handles GET /api/v2/bulk/history.
// Returns paginated bulk operation history.
func (h *BulkHandler) History(c echo.Context) error {
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	// Sudo admins can see all history; non-sudo see only their own.
	var filterAdminID *uuid.UUID
	if !claims.Sudo {
		filterAdminID = &claims.AdminID
	}

	operations, total, err := h.svc.History(c.Request().Context(), filterAdminID, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list history")
	}
	if operations == nil {
		operations = []*domain.BulkOperation{}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"operations": operations,
		"total":      total,
	})
}

package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// AuditEventHandlers handles audit event API endpoints
type AuditEventHandlers struct {
	eventRepo port.AuditEventRepository
	log       *slog.Logger
}

// NewAuditEventHandlers creates new audit event handlers
func NewAuditEventHandlers(eventRepo port.AuditEventRepository, log *slog.Logger) *AuditEventHandlers {
	if log == nil {
		log = slog.Default()
	}
	return &AuditEventHandlers{
		eventRepo: eventRepo,
		log:       log,
	}
}

// GetEvent retrieves a specific audit event
func (h *AuditEventHandlers) GetEvent(c echo.Context) error {
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid event ID"})
	}

	event, err := h.eventRepo.GetEvent(c.Request().Context(), eventID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "event not found"})
		}
		h.log.Error("failed to get audit event", "error", err, "event_id", eventID)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get event"})
	}

	return c.JSON(http.StatusOK, event)
}

// ListEvents retrieves audit events with filtering
func (h *AuditEventHandlers) ListEvents(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var adminID *uuid.UUID
	if adminIDStr := c.QueryParam("admin_id"); adminIDStr != "" {
		parsed, err := uuid.Parse(adminIDStr)
		if err == nil {
			adminID = &parsed
		}
	}

	eventType := c.QueryParam("event_type")
	eventTypePtr := &eventType
	if eventType == "" {
		eventTypePtr = nil
	}

	severity := c.QueryParam("severity")
	severityPtr := &severity
	if severity == "" {
		severityPtr = nil
	}

	var startDate, endDate *time.Time
	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &t
		}
	}

	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &t
		}
	}

	events, total, err := h.eventRepo.ListEvents(c.Request().Context(), adminID, eventTypePtr, severityPtr, startDate, endDate, limit, offset)
	if err != nil {
		h.log.Error("failed to list audit events", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list events"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// SearchEvents performs full-text search on audit events
func (h *AuditEventHandlers) SearchEvents(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "search query required"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	events, total, err := h.eventRepo.SearchEvents(c.Request().Context(), query, limit, offset)
	if err != nil {
		h.log.Error("failed to search audit events", "error", err, "query", query)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to search events"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEventsByAdmin retrieves all events for a specific admin
func (h *AuditEventHandlers) GetEventsByAdmin(c echo.Context) error {
	adminIDStr := c.Param("admin_id")
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid admin ID"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	events, total, err := h.eventRepo.GetEventsByAdmin(c.Request().Context(), adminID, limit, offset)
	if err != nil {
		h.log.Error("failed to get admin events", "error", err, "admin_id", adminID)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get admin events"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEventCount retrieves total count of audit events
func (h *AuditEventHandlers) GetEventCount(c echo.Context) error {
	count, err := h.eventRepo.CountEvents(c.Request().Context())
	if err != nil {
		h.log.Error("failed to count audit events", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to count events"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total": count,
	})
}

// DeleteOldEvents deletes audit events older than retention days (admin-only)
func (h *AuditEventHandlers) DeleteOldEvents(c echo.Context) error {
	var req struct {
		RetentionDays int `json:"retention_days"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.RetentionDays < 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "retention_days must be at least 1"})
	}

	deleted, err := h.eventRepo.DeleteOldEvents(c.Request().Context(), req.RetentionDays)
	if err != nil {
		h.log.Error("failed to delete old events", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete events"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"deleted": deleted,
	})
}

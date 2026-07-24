package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// JobsHandler serves background job status and listing endpoints.
type JobsHandler struct {
	queue *service.JobQueue
}

// NewJobsHandler creates the handler.
func NewJobsHandler(queue *service.JobQueue) *JobsHandler {
	return &JobsHandler{queue: queue}
}

// Register mounts job routes on the given Echo group.
func (h *JobsHandler) Register(g *echo.Group) {
	g.GET("/jobs/:id", h.GetJobStatus)
	g.GET("/jobs", h.ListJobs)
}

// GetJobStatus handles GET /api/v2/jobs/:id.
func (h *JobsHandler) GetJobStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job ID")
	}

	job, err := h.queue.GetStatus(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, job)
}

// ListJobs handles GET /api/v2/jobs.
func (h *JobsHandler) ListJobs(c echo.Context) error {
	jobs, err := h.queue.ListRecent(c.Request().Context(), 50)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, jobs)
}

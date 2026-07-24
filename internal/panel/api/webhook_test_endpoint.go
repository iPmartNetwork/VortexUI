package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// WebhookTestHandler provides a test endpoint for webhook delivery verification.
type WebhookTestHandler struct {
	client *http.Client
}

// NewWebhookTestHandler creates the handler.
func NewWebhookTestHandler() *WebhookTestHandler {
	return &WebhookTestHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Register mounts webhook test routes.
func (h *WebhookTestHandler) Register(g *echo.Group) {
	g.POST("/webhooks/test", h.TestWebhook)
}

type webhookTestRequest struct {
	URL     string         `json:"url"`
	Secret  string         `json:"secret,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

type webhookTestResponse struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"status_code"`
	Response   string `json:"response,omitempty"`
	Error      string `json:"error,omitempty"`
	Duration   string `json:"duration"`
}

// TestWebhook handles POST /api/v2/webhooks/test.
// Sends a sample payload to the configured URL and returns the delivery result.
func (h *WebhookTestHandler) TestWebhook(c echo.Context) error {
	var req webhookTestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "url is required")
	}

	// Build sample payload
	payload := req.Payload
	if payload == nil {
		payload = map[string]any{
			"event": "test",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"message": "This is a test webhook delivery from VortexUI",
		}
	}

	body, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	start := time.Now()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.URL, bytes.NewReader(body))
	if err != nil {
		return c.JSON(http.StatusOK, webhookTestResponse{
			Success:  false,
			Error:    "invalid URL: " + err.Error(),
			Duration: time.Since(start).String(),
		})
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "VortexUI-Webhook/1.0")

	if req.Secret != "" {
		httpReq.Header.Set("X-Webhook-Secret", req.Secret)
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return c.JSON(http.StatusOK, webhookTestResponse{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).String(),
		})
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	return c.JSON(http.StatusOK, webhookTestResponse{
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		StatusCode: resp.StatusCode,
		Response:   string(respBody),
		Duration:   time.Since(start).String(),
	})
}

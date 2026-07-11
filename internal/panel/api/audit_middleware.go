package api

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// AuditMiddleware logs audit events for admin actions
func AuditMiddleware(auditService *service.AuditService, log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Only audit mutating requests
			if !isAuditableMethod(req.Method) {
				return next(c)
			}

			// Read request body for audit
			bodyBytes, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Capture response
			recorder := &responseRecorder{
				ResponseWriter: res.Writer,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}
			res.Writer = recorder

			// Call next handler
			err := next(c)

			// Extract admin info
			adminID, ok := c.Get("admin_id").(uuid.UUID)
			if !ok {
				return err
			}

			// Log audit asynchronously to not block request
			go func() {
				ctx, cancel := c.Request().Context(), func() {}
				defer cancel()

				// Extract resource info from request path
				action := getActionFromMethod(req.Method)
				resourceType := extractResourceType(req.URL.Path)

				if err := auditService.LogAction(ctx, adminID, action, resourceType, nil, string(bodyBytes), nil); err != nil {
					log.Error("failed to log audit", "err", err)
				}
			}()

			return err
		}
	}
}

// responseRecorder records response details
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// isAuditableMethod checks if the HTTP method should be audited
func isAuditableMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// getActionFromMethod converts HTTP method to audit action
func getActionFromMethod(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// extractResourceType extracts resource type from API path
func extractResourceType(path string) string {
	if bytes.Contains([]byte(path), []byte("/users")) {
		return "user"
	}
	if bytes.Contains([]byte(path), []byte("/nodes")) {
		return "node"
	}
	if bytes.Contains([]byte(path), []byte("/plans")) {
		return "plan"
	}
	if bytes.Contains([]byte(path), []byte("/admins")) {
		return "admin"
	}
	if bytes.Contains([]byte(path), []byte("/inbounds")) {
		return "inbound"
	}
	if bytes.Contains([]byte(path), []byte("/subscriptions")) {
		return "subscription"
	}
	return "unknown"
}

// SessionMiddleware validates session and extracts session info
func SessionMiddleware(sessionService *service.SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract session token from header
			sessionToken := c.Request().Header.Get("X-Session-Token")
			if sessionToken == "" {
				// Try from Authorization header
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					sessionToken = authHeader[7:]
				}
			}

			// Validate session
			if sessionToken != "" {
				session, err := sessionService.ValidateSession(c.Request().Context(), sessionToken)
				if err == nil && session.IsActive() {
					c.Set("session_id", session.ID)
					c.Set("admin_id", session.AdminID)
					c.Set("ip_address", session.IPAddress)
					c.Set("user_agent", session.UserAgent)
					return next(c)
				}
			}

			// If no valid session, continue (will be handled by auth middleware)
			return next(c)
		}
	}
}

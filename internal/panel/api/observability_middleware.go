package api

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// MetricsMiddleware collects request metrics
func MetricsMiddleware(metricsService *service.MetricsCollectorService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Capture request body size
			var bodyBytes []byte
			if req.Body != nil {
				bodyBytes, _ = io.ReadAll(req.Body)
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// Custom response writer to capture response size
			responseWriter := &responseCapture{ResponseWriter: res.Writer}
			res.Writer = responseWriter

			// Execute handler
			err := next(c)

			// Record metrics
			duration := time.Since(start)
			
			// Extract trace ID from context
			traceID := ""
			if trace := c.Request().Context().Value("trace_id"); trace != nil {
				if str, ok := trace.(string); ok {
					traceID = str
				}
			}

			metric := &domain.RequestMetrics{
				Method:    req.Method,
				Path:      req.RequestURI,
				Status:    res.Status,
				Latency:   duration.Milliseconds(),
				BytesIn:   int64(len(bodyBytes)),
				BytesOut:  int64(responseWriter.written),
				Timestamp: start,
				TraceID:   traceID,
			}

			go func() {
				_ = metricsService.RecordRequestMetric(c.Request().Context(), metric)
			}()

			return err
		}
	}
}

// TraceMiddleware injects trace context into requests
func TraceMiddleware(traceService *service.TraceManagerService, log *slog.Logger) echo.MiddlewareFunc {
	if log == nil {
		log = slog.Default()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Extract trace context from headers or create new
			traceCtx := traceService.ExtractTraceContext(headerMap(req.Header))
			if traceCtx.TraceID == "" {
				// Generate new trace
				_, spanEnd := traceService.StartSpan(c.Request().Context(), req.URL.Path, map[string]interface{}{
					"method": req.Method,
					"path":   req.URL.Path,
				})
				defer func() { _ = spanEnd() }()
			}

			// Inject trace into context using context.WithValue
			type contextKey string
			newCtx := context.WithValue(c.Request().Context(), contextKey("trace_id"), traceCtx.TraceID)
			newCtx = context.WithValue(newCtx, contextKey("span_id"), traceCtx.SpanID)
			newCtx = context.WithValue(newCtx, contextKey("parent_span_id"), traceCtx.ParentID)
			c.SetRequest(c.Request().WithContext(newCtx))

			// Inject into response headers
			c.Response().Header().Set("X-Trace-ID", traceCtx.TraceID)
			c.Response().Header().Set("X-Span-ID", traceCtx.SpanID)

			log.Debug("trace middleware", "trace_id", traceCtx.TraceID, "span_id", traceCtx.SpanID)

			return next(c)
		}
	}
}

// responseCapture captures response metrics
type responseCapture struct {
	http.ResponseWriter
	written int
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	n, err := rc.ResponseWriter.Write(b)
	rc.written += n
	return n, err
}

// headerMap converts http.Header to map[string]string
func headerMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range h {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}



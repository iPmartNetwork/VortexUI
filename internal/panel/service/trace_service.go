package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// TraceManagerService implements distributed tracing
type TraceManagerService struct {
	log *slog.Logger
}

// NewTraceManagerService creates a new trace manager
func NewTraceManagerService(log *slog.Logger) *TraceManagerService {
	if log == nil {
		log = slog.Default()
	}
	return &TraceManagerService{
		log: log,
	}
}

// StartSpan creates a new trace span
func (t *TraceManagerService) StartSpan(ctx context.Context, spanName string, attrs map[string]interface{}) (context.Context, func() error) {
	// Generate span ID
	spanID := generateID()

	// Get or create trace ID
	traceID := ""
	if trace := ctx.Value("trace_id"); trace != nil {
		if str, ok := trace.(string); ok {
			traceID = str
		}
	}
	if traceID == "" {
		traceID = generateID()
	}

	type contextKey string
	// Get parent span ID if exists
	parentID := ""
	if parent := ctx.Value(contextKey("span_id")); parent != nil {
		if str, ok := parent.(string); ok {
			parentID = str
		}
	}

	// Create new context with span
	newCtx := context.WithValue(ctx, contextKey("trace_id"), traceID)
	newCtx = context.WithValue(newCtx, contextKey("span_id"), spanID)
	newCtx = context.WithValue(newCtx, contextKey("parent_span_id"), parentID)

	startTime := time.Now()

	// Return context and cleanup function
	return newCtx, func() error {
		duration := time.Since(startTime)
		t.log.Debug("span completed",
			"trace_id", traceID,
			"span_id", spanID,
			"span_name", spanName,
			"duration_ms", duration.Milliseconds(),
		)
		return nil
	}
}

// GetTraceContext retrieves the current trace context
func (t *TraceManagerService) GetTraceContext(ctx context.Context) *domain.TraceContext {
	traceID := ""
	if trace := ctx.Value("trace_id"); trace != nil {
		if str, ok := trace.(string); ok {
			traceID = str
		}
	}

	spanID := ""
	if span := ctx.Value("span_id"); span != nil {
		if str, ok := span.(string); ok {
			spanID = str
		}
	}

	parentID := ""
	if parent := ctx.Value("parent_span_id"); parent != nil {
		if str, ok := parent.(string); ok {
			parentID = str
		}
	}

	sampled := false
	if s := ctx.Value("sampled"); s != nil {
		if b, ok := s.(bool); ok {
			sampled = b
		}
	}

	return &domain.TraceContext{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		Sampled:   sampled,
		Timestamp: time.Now().Unix(),
	}
}

// InjectTraceContext injects trace context into headers
func (t *TraceManagerService) InjectTraceContext(ctx context.Context, headers map[string]string) error {
	traceCtx := t.GetTraceContext(ctx)

	if headers == nil {
		return fmt.Errorf("headers cannot be nil")
	}

	headers["X-Trace-ID"] = traceCtx.TraceID
	headers["X-Span-ID"] = traceCtx.SpanID
	if traceCtx.ParentID != "" {
		headers["X-Parent-Span-ID"] = traceCtx.ParentID
	}

	return nil
}

// ExtractTraceContext extracts trace context from headers
func (t *TraceManagerService) ExtractTraceContext(headers map[string]string) *domain.TraceContext {
	return &domain.TraceContext{
		TraceID:   headers["X-Trace-ID"],
		SpanID:    headers["X-Span-ID"],
		ParentID:  headers["X-Parent-Span-ID"],
		Sampled:   headers["X-Sampled"] == "true",
		Timestamp: time.Now().Unix(),
	}
}

// generateID generates a random trace/span ID
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// StructuredLoggerService implements structured logging
type StructuredLoggerService struct {
	logs   []*domain.LogEntry
	maxLen int
	mu     sync.RWMutex
	log    *slog.Logger
}

// NewStructuredLoggerService creates a new structured logger
func NewStructuredLoggerService(log *slog.Logger, maxLogs int) *StructuredLoggerService {
	if log == nil {
		log = slog.Default()
	}
	if maxLogs <= 0 {
		maxLogs = 1000
	}
	return &StructuredLoggerService{
		logs:   make([]*domain.LogEntry, 0, maxLogs),
		maxLen: maxLogs,
		log:    log,
	}
}

// LogWithContext logs a message with context
func (s *StructuredLoggerService) LogWithContext(ctx context.Context, level string, message string, attrs map[string]interface{}) error {
	traceID := ""
	if trace := extractTraceID(ctx); trace != "" {
		traceID = trace
	}

	entry := &domain.LogEntry{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		Level:      level,
		Message:    message,
		TraceID:    traceID,
		Attributes: attrs,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = append(s.logs, entry)

	// Keep buffer size bounded
	if len(s.logs) > s.maxLen {
		s.logs = s.logs[1:]
	}

	// Also log to default logger
	switch level {
	case "error":
		s.log.Error(message, "attrs", attrs, "trace_id", traceID)
	case "warn":
		s.log.Warn(message, "attrs", attrs, "trace_id", traceID)
	case "info":
		s.log.Info(message, "attrs", attrs, "trace_id", traceID)
	default:
		s.log.Debug(message, "attrs", attrs, "trace_id", traceID)
	}

	return nil
}

// LogError logs an error with structured context
func (s *StructuredLoggerService) LogError(ctx context.Context, err error, message string, attrs map[string]interface{}) error {
	traceID := ""
	if trace := extractTraceID(ctx); trace != "" {
		traceID = trace
	}

	errorInfo := &domain.ErrorInfo{
		Message: err.Error(),
		Context: make(map[string]string),
	}

	if attrs != nil {
		for k, v := range attrs {
			if sv, ok := v.(string); ok {
				errorInfo.Context[k] = sv
			}
		}
	}

	entry := &domain.LogEntry{
		ID:         uuid.New(),
		Timestamp:  time.Now(),
		Level:      "error",
		Message:    message,
		TraceID:    traceID,
		Error:      errorInfo,
		Attributes: attrs,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = append(s.logs, entry)

	// Keep buffer size bounded
	if len(s.logs) > s.maxLen {
		s.logs = s.logs[1:]
	}

	s.log.Error(message, "error", err, "trace_id", traceID, "attrs", attrs)

	return nil
}

// GetLogs retrieves recent logs with optional filtering
func (s *StructuredLoggerService) GetLogs(ctx context.Context, level string, limit int) ([]*domain.LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	var result []*domain.LogEntry
	start := len(s.logs) - limit
	if start < 0 {
		start = 0
	}

	for i := len(s.logs) - 1; i >= start && len(result) < limit; i-- {
		entry := s.logs[i]
		if level == "" || entry.Level == level {
			result = append(result, entry)
		}
	}

	return result, nil
}

// extractTraceID extracts trace ID from context
func extractTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if str, ok := traceID.(string); ok {
			return str
		}
	}
	return ""
}

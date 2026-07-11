package domain

import (
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned by repositories when a requested entity does not
// exist, decoupling callers from any specific storage driver's sentinel.
var ErrNotFound = errors.New("not found")

// ErrInvalid wraps domain validation failures so callers (e.g. the HTTP layer)
// can map them to a 400 with errors.Is, regardless of the specific message.
var ErrInvalid = errors.New("invalid")

// ErrUnauthorized is returned when authentication fails
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden is returned when authorization fails
var ErrForbidden = errors.New("forbidden")

// ErrConflict is returned when there's a resource conflict
var ErrConflict = errors.New("conflict")

// ErrRateLimited is returned when rate limit is exceeded
var ErrRateLimited = errors.New("rate limited")

// ErrSessionExpired is returned when a session has expired
var ErrSessionExpired = errors.New("session expired")

// errInvalid builds an ErrInvalid-wrapped, formatted validation error.
func errInvalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// ErrorCode represents a standard error code in the system
type ErrorCode struct {
	Code       string    `db:"code" json:"code"`
	Description string  `db:"description" json:"description"`
	HTTPStatus int      `db:"http_status" json:"http_status"`
	Category   string   `db:"category" json:"category"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// StandardErrorCodes are predefined error codes
const (
	ErrCodeUnauthorized      = "ERR_UNAUTHORIZED"
	ErrCodeForbidden         = "ERR_FORBIDDEN"
	ErrCodeNotFound          = "ERR_NOT_FOUND"
	ErrCodeInvalidInput      = "ERR_INVALID_INPUT"
	ErrCodeConflict          = "ERR_CONFLICT"
	ErrCodeRateLimited       = "ERR_RATE_LIMITED"
	ErrCodeInternalError     = "ERR_INTERNAL_ERROR"
	ErrCodeDatabaseError     = "ERR_DATABASE_ERROR"
	ErrCodeExternalService   = "ERR_EXTERNAL_SERVICE"
	ErrCodeValidationFailed  = "ERR_VALIDATION_FAILED"
	ErrCodeSessionExpired    = "ERR_SESSION_EXPIRED"
	ErrCodeIPNotWhitelisted  = "ERR_IP_NOT_WHITELISTED"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
	Time    time.Time   `json:"time"`
}

// DomainError wraps an error with a code for better error handling
type DomainError struct {
	Code    string
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the cause error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

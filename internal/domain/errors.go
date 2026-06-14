package domain

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned by repositories when a requested entity does not
// exist, decoupling callers from any specific storage driver's sentinel.
var ErrNotFound = errors.New("not found")

// ErrInvalid wraps domain validation failures so callers (e.g. the HTTP layer)
// can map them to a 400 with errors.Is, regardless of the specific message.
var ErrInvalid = errors.New("invalid")

// errInvalid builds an ErrInvalid-wrapped, formatted validation error.
func errInvalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

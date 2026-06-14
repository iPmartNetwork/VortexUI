package domain

import "errors"

// ErrNotFound is returned by repositories when a requested entity does not
// exist, decoupling callers from any specific storage driver's sentinel.
var ErrNotFound = errors.New("not found")

package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditEntry records one mutating admin action for accountability. It is written
// by the audit middleware for every authenticated POST/PUT/DELETE request.
type AuditEntry struct {
	ID       uuid.UUID  `json:"id"`
	Time     time.Time  `json:"time"`
	AdminID  *uuid.UUID `json:"admin_id,omitempty"`
	Username string     `json:"username"` // denormalized for display (admin may be deleted)
	Method   string     `json:"method"`
	Path     string     `json:"path"`
	Status   int        `json:"status"`
	IP       string     `json:"ip"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIToken is a long-lived automation credential. The raw secret is shown once
// at creation; only its hash is stored. It authenticates as its owning admin.
type APIToken struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	AdminID    uuid.UUID  `json:"admin_id"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

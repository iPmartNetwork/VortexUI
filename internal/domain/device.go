package domain

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a user's registered hardware device, identified by a
// unique hardware ID (HWID). Used for device-based access control and
// connection tracking.
type Device struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	HWID      string    `json:"hwid"`
	OS        string    `json:"os"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
}

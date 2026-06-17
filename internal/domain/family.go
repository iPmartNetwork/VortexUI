package domain

import (
	"time"

	"github.com/google/uuid"
)

// FamilyGroup represents a shared subscription group where one parent account
// manages multiple member devices/users under a shared data pool.
type FamilyGroup struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	OwnerID     uuid.UUID      `json:"owner_id"`     // the parent user
	OwnerName   string         `json:"owner_name,omitempty"`
	DataLimit   int64          `json:"data_limit"`   // shared pool bytes; 0 = unlimited
	UsedTraffic int64          `json:"used_traffic"` // sum of all members
	MaxMembers  int            `json:"max_members"`  // max devices/sub-users
	MemberQuota int64          `json:"member_quota"` // per-member cap; 0 = no per-member limit
	Members     []FamilyMember `json:"members,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// FamilyMember is a member of a family group.
type FamilyMember struct {
	ID          uuid.UUID `json:"id"`
	GroupID     uuid.UUID `json:"group_id"`
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username,omitempty"`
	UsedTraffic int64     `json:"used_traffic"`
	Label       string    `json:"label"` // e.g. "Dad's Phone", "Living Room TV"
	JoinedAt    time.Time `json:"joined_at"`
}

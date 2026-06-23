package domain

import "github.com/google/uuid"

// AdminUserStats aggregates usage for a reseller's owned users.
type AdminUserStats struct {
	UserCount        int64 `json:"user_count"`
	TrafficUsed      int64 `json:"traffic_used"`      // sum(used_traffic)
	TrafficAllocated int64 `json:"traffic_allocated"` // sum(data_limit)
}

// AdminQuotaUsage combines admin limits with live usage (for dashboards).
type AdminQuotaUsage struct {
	AdminID          uuid.UUID `json:"admin_id"`
	Username         string    `json:"username"`
	UserQuota        int       `json:"user_quota"`        // 0 = unlimited
	UserCount        int64     `json:"user_count"`
	UsersRemaining   *int64    `json:"users_remaining"`   // nil when unlimited
	TrafficQuota     int64     `json:"traffic_quota"`     // 0 = unlimited (assignment pool, bytes)
	TrafficQuotaMode string    `json:"traffic_quota_mode"`
	TrafficUsed      int64     `json:"traffic_used"`
	TrafficAllocated int64     `json:"traffic_allocated"`
	TrafficRemaining *int64    `json:"traffic_remaining"` // nil when unlimited
}

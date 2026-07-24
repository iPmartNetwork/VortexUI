package domain

import (
	"time"

	"github.com/google/uuid"
)

// BulkOperationType enumerates the supported bulk modification types.
type BulkOperationType string

const (
	BulkOpAddGroups       BulkOperationType = "add_groups"
	BulkOpRemoveGroups    BulkOperationType = "remove_groups"
	BulkOpAddExpireDays   BulkOperationType = "add_expire_days"
	BulkOpSubExpireDays   BulkOperationType = "sub_expire_days"
	BulkOpAddDataLimit    BulkOperationType = "add_data_limit"
	BulkOpSubDataLimit    BulkOperationType = "sub_data_limit"
	BulkOpUpdateProxy     BulkOperationType = "update_proxy_settings"
	BulkOpAllocateWGPeers BulkOperationType = "allocate_wg_peers"
	BulkOpRepairWGPeers   BulkOperationType = "repair_wg_peers"
	BulkOpChangeStatus    BulkOperationType = "change_status"
)

// BulkFilter determines which users a bulk operation targets.
type BulkFilter struct {
	Statuses []UserStatus `json:"statuses,omitempty"`
	AdminID  *uuid.UUID   `json:"admin_id,omitempty"`
	Groups   []string     `json:"groups,omitempty"`
}

// BulkOperation represents an executed (or executing) bulk operation with its
// recorded history metadata.
type BulkOperation struct {
	ID            uuid.UUID         `json:"id"`
	AdminID       uuid.UUID         `json:"admin_id"`
	OperationType BulkOperationType `json:"operation_type"`
	Parameters    map[string]any    `json:"parameters"`
	Filters       BulkFilter        `json:"filters"`
	AffectedCount int               `json:"affected_count"`
	Status        string            `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
}

// BulkPreviewResult is the dry-run output showing how many users would be
// affected and a human-readable summary of the proposed changes.
type BulkPreviewResult struct {
	AffectedCount int            `json:"affected_count"`
	Summary       map[string]any `json:"summary"`
}

package events

// All webhook-deliverable event types. The webhook subscriber (internal/notify/webhook.go)
// delivers any event published to the bus that has a non-empty Type field.
//
// Admin-facing events (delivered to panel webhook):
//   - node.up            — node reconnected to panel
//   - node.down          — node went offline / failover triggered
//   - node.alert         — resource threshold crossed (CPU/memory/disk)
//   - node.disconnect_alert — unreachable > 5 minutes
//   - node.auto_recover  — panel auto-recovered a node
//   - user.created       — new user provisioned
//   - user.deleted       — user removed
//   - user.quota_warning — user traffic >= 80% of data limit
//   - user.expired       — user subscription expired
//   - cert.expiring      — TLS certificate expires within 7 days
//   - admin.login        — admin logged into panel
//   - admin.quota_warning — reseller approaching quota
//   - security.probe     — active probing attempt detected
//   - backup.completed   — auto-backup finished
//
// Reseller-facing events (delivered to per-reseller webhook):
//   - user.created       — reseller's user provisioned
//   - user.deleted       — reseller's user removed

// AdminEventTypes lists events routed to panel-level webhook.
var AdminEventTypes = []Type{
	NodeUp,
	NodeDown,
	NodeAlert,
	NodeDisconnectAlert,
	NodeAutoRecover,
	UserCreated,
	UserDeleted,
	UserQuotaWarn,
	UserExpired,
	CertExpiring,
	AdminQuotaWarning,
	SecurityProbe,
	BackupCompleted,
}

// BackupCompleted fires when an auto-backup finishes successfully.
const BackupCompleted Type = "backup.completed"

// AdminLogin fires when an admin successfully authenticates to the panel.
const AdminLogin Type = "admin.login"

// ResellerEventTypes lists events routed to per-reseller webhooks.
var ResellerEventTypes = []Type{
	UserCreated,
	UserDeleted,
}

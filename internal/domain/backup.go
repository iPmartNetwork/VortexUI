package domain

import (
	"time"

	"github.com/google/uuid"
)

// BackupVersion is the schema version of the exported configuration document.
// Restore accepts the current version and legacy versions (1, 2).
const (
	BackupVersion       = 3
	BackupVersionV2     = 2
	BackupVersionLegacy = 1

	ResellerBackupVersion = 2
)

// BackupFormat identifies the export transport.
type BackupFormat string

const (
	BackupFormatJSON BackupFormat = "json"
	BackupFormatFull BackupFormat = "full"
)

// BackupRestoreMode controls how a restore is applied.
type BackupRestoreMode string

const (
	// BackupRestoreConfig replaces proxy config and supplemental business tables
	// from a JSON v3 document.
	BackupRestoreConfig BackupRestoreMode = "config"
	// BackupRestoreFull replaces the entire PostgreSQL database from a pg_dump
	// archive (server migration).
	BackupRestoreFull BackupRestoreMode = "full"
)

// BackupAdminScope captures per-admin resource access restored after nodes/plans
// exist.
type BackupAdminScope struct {
	AdminID    uuid.UUID   `json:"admin_id"`
	InboundIDs []uuid.UUID `json:"inbound_ids,omitempty"`
	PlanIDs    []uuid.UUID `json:"plan_ids,omitempty"`
	NodeIDs    []uuid.UUID `json:"node_ids,omitempty"`
}

// BackupAdminCredential stores login secrets omitted from the public Admin JSON.
type BackupAdminCredential struct {
	AdminID       uuid.UUID `json:"admin_id"`
	PasswordHash  string    `json:"password_hash"`
	TOTPSecret    string    `json:"totp_secret,omitempty"`
	WebhookSecret string    `json:"webhook_secret,omitempty"`
}

// BackupSupplemental holds rows from business tables not covered by the core
// config snapshot. Keys are table names; values are row objects (column → value).
type BackupSupplemental struct {
	Tables map[string][]map[string]any `json:"tables"`
}

// BackupManifest summarizes export/restore coverage for the UI and migration
// planning.
type BackupManifest struct {
	Format         BackupFormat      `json:"format"`
	ExportedAt     time.Time         `json:"exported_at"`
	Version        int               `json:"version"`
	Counts         BackupCounts      `json:"counts"`
	Usage          BackupUsageSummary `json:"usage"`
	IncludedTables []string          `json:"included_tables,omitempty"`
	ExcludedTables []string          `json:"excluded_tables,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
}

// BackupCounts holds entity totals included in the snapshot.
type BackupCounts struct {
	Roles          int `json:"roles"`
	Admins         int `json:"admins"`
	Plans          int `json:"plans"`
	Nodes          int `json:"nodes"`
	Inbounds       int `json:"inbounds"`
	Outbounds      int `json:"outbounds"`
	Routing        int `json:"routing"`
	Balancers      int `json:"balancers"`
	Users          int `json:"users"`
	Bindings       int `json:"bindings"`
	Orders         int `json:"orders"`
	WalletLedger   int `json:"wallet_ledger"`
	WalletDeposits int `json:"wallet_deposits"`
	WalletPackages int `json:"wallet_packages"`
}

// BackupUsageSummary aggregates traffic and wallet visibility for operators.
type BackupUsageSummary struct {
	TotalUsers            int                    `json:"total_users"`
	TotalUsedTraffic      int64                  `json:"total_used_traffic"`
	TotalDataLimit        int64                  `json:"total_data_limit"`
	TotalRemainingTraffic int64                  `json:"total_remaining_traffic"`
	UsersOverLimit        int                    `json:"users_over_limit"`
	UnlimitedUsers        int                    `json:"unlimited_users"`
	Resellers             []ResellerUsageSummary `json:"resellers,omitempty"`
}

// ResellerUsageSummary is per-reseller traffic and wallet visibility.
type ResellerUsageSummary struct {
	AdminID            uuid.UUID `json:"admin_id"`
	Username           string    `json:"username"`
	UserCount          int       `json:"user_count"`
	AllocatedTraffic   int64     `json:"allocated_traffic"`
	UsedTraffic        int64     `json:"used_traffic"`
	RemainingTraffic   int64     `json:"remaining_traffic"`
	WalletTrafficBytes int64     `json:"wallet_traffic_bytes"`
	WalletUserCredits  int       `json:"wallet_user_credits"`
}

// Backup is a portable snapshot of the panel's proxy configuration, operator
// accounts, billing/wallet state, and user traffic counters.
type Backup struct {
	Version          int                     `json:"version"`
	ExportedAt       time.Time               `json:"exported_at"`
	Manifest         *BackupManifest         `json:"manifest,omitempty"`
	Roles            []*Role                 `json:"roles,omitempty"`
	Admins           []*Admin                `json:"admins,omitempty"`
	AdminCredentials []BackupAdminCredential `json:"admin_credentials,omitempty"`
	AdminScopes      []BackupAdminScope      `json:"admin_scopes,omitempty"`
	Plans            []*Plan                 `json:"plans,omitempty"`
	PortalBranding   []*PortalBranding       `json:"portal_branding,omitempty"`
	Nodes            []*Node                 `json:"nodes"`
	Inbounds         []*Inbound              `json:"inbounds"`
	Outbounds        []*Outbound             `json:"outbounds"`
	Routing          []*RoutingRule          `json:"routing"`
	Balancers        []*Balancer             `json:"balancers"`
	Users            []*User                 `json:"users"`
	Bindings         []UserProxy             `json:"bindings"`
	Supplemental     *BackupSupplemental     `json:"supplemental,omitempty"`
}

// ResellerBackup is a scoped snapshot for a single reseller account.
type ResellerBackup struct {
	Version        int                     `json:"version"`
	ExportedAt     time.Time               `json:"exported_at"`
	AdminID        uuid.UUID               `json:"admin_id"`
	AdminUsername  string                  `json:"admin_username,omitempty"`
	Manifest       ResellerBackupManifest  `json:"manifest"`
	Users          []*User                 `json:"users"`
	Bindings       []UserProxy             `json:"bindings"`
	WalletLedger   []WalletLedgerEntry     `json:"wallet_ledger,omitempty"`
	PaymentConfig  *ResellerPaymentConfig  `json:"payment_config,omitempty"`
	Orders         []Order                 `json:"orders,omitempty"`
	PortalBranding *PortalBranding         `json:"portal_branding,omitempty"`
}

// ResellerBackupManifest summarizes a reseller export.
type ResellerBackupManifest struct {
	UserCount             int   `json:"user_count"`
	TotalUsedTraffic      int64 `json:"total_used_traffic"`
	TotalDataLimit        int64 `json:"total_data_limit"`
	TotalRemainingTraffic int64 `json:"total_remaining_traffic"`
	OrdersCount           int   `json:"orders_count"`
	LedgerEntries         int   `json:"ledger_entries"`
}

// BackupRestoreReport is returned after a successful restore.
type BackupRestoreReport struct {
	Mode     BackupRestoreMode `json:"mode"`
	Restored BackupCounts      `json:"restored"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ExcludedFromJSONBackup documents tables intentionally omitted from JSON export.
var ExcludedFromJSONBackup = []string{
	"audit_log",
	"api_tokens",
	"reality_scans",
	"probe_events",
	"doh_query_logs",
	"fingerprint_events",
	"ip_limit_events",
	"clean_ip_scans",
	"migration_events",
	"referral_events",
	"federation_sync_events",
	"quota_notify_events",
	"ticket_messages",
	"tickets",
}

func SupportedBackupVersions() []int {
	return []int{BackupVersion, BackupVersionV2, BackupVersionLegacy}
}

// IsSupportedBackupVersion reports whether v can be restored.
func IsSupportedBackupVersion(v int) bool {
	for _, ok := range SupportedBackupVersions() {
		if ok == v {
			return true
		}
	}
	return false
}

// BuildUsageSummary computes traffic totals from users and optional admin list.
func BuildUsageSummary(users []*User, admins []*Admin) BackupUsageSummary {
	out := BackupUsageSummary{TotalUsers: len(users)}
	byAdmin := make(map[uuid.UUID]*ResellerUsageSummary)
	for _, a := range admins {
		byAdmin[a.ID] = &ResellerUsageSummary{
			AdminID:            a.ID,
			Username:           a.Username,
			WalletTrafficBytes: a.WalletTrafficBytes,
			WalletUserCredits:  a.WalletUserCredits,
		}
	}
	for _, u := range users {
		out.TotalUsedTraffic += u.UsedTraffic
		if u.DataLimit > 0 {
			out.TotalDataLimit += u.DataLimit
			rem := u.DataLimit - u.UsedTraffic
			if rem < 0 {
				rem = 0
				out.UsersOverLimit++
			}
			out.TotalRemainingTraffic += rem
		} else {
			out.UnlimitedUsers++
		}
		if u.AdminID == nil {
			continue
		}
		rs, ok := byAdmin[*u.AdminID]
		if !ok {
			rs = &ResellerUsageSummary{AdminID: *u.AdminID}
			byAdmin[*u.AdminID] = rs
		}
		rs.UserCount++
		rs.UsedTraffic += u.UsedTraffic
		if u.DataLimit > 0 {
			rs.AllocatedTraffic += u.DataLimit
			rem := u.DataLimit - u.UsedTraffic
			if rem < 0 {
				rem = 0
			}
			rs.RemainingTraffic += rem
		}
	}
	if len(byAdmin) > 0 {
		out.Resellers = make([]ResellerUsageSummary, 0, len(byAdmin))
		for _, rs := range byAdmin {
			if rs.Username == "" {
				continue
			}
			out.Resellers = append(out.Resellers, *rs)
		}
	}
	return out
}

// BuildBackupManifest assembles a manifest from a backup document.
func BuildBackupManifest(b *Backup, format BackupFormat) BackupManifest {
	if b == nil {
		return BackupManifest{Format: format, Version: BackupVersion}
	}
	m := BackupManifest{
		Format:     format,
		ExportedAt: b.ExportedAt,
		Version:    b.Version,
		Counts: BackupCounts{
			Roles: len(b.Roles), Admins: len(b.Admins), Plans: len(b.Plans),
			Nodes: len(b.Nodes), Inbounds: len(b.Inbounds), Outbounds: len(b.Outbounds),
			Routing: len(b.Routing), Balancers: len(b.Balancers),
			Users: len(b.Users), Bindings: len(b.Bindings),
		},
		Usage: BuildUsageSummary(b.Users, b.Admins),
	}
	if b.Supplemental != nil && b.Supplemental.Tables != nil {
		m.Counts.Orders = len(b.Supplemental.Tables["orders"])
		m.Counts.WalletLedger = len(b.Supplemental.Tables["admin_wallet_ledger"])
		m.Counts.WalletDeposits = len(b.Supplemental.Tables["wallet_deposits"])
		m.Counts.WalletPackages = len(b.Supplemental.Tables["wallet_packages"])
		for name := range b.Supplemental.Tables {
			m.IncludedTables = append(m.IncludedTables, name)
		}
	}
	if b.Version < BackupVersion {
		m.Warnings = append(m.Warnings, "legacy backup version; supplemental billing data may be missing")
	}
	if len(b.AdminCredentials) == 0 && len(b.Admins) > 0 {
		m.Warnings = append(m.Warnings, "admin login secrets not included; restored admins may need password reset")
	}
	return m
}

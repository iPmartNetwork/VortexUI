package domain

import (
	"time"

	"github.com/google/uuid"
)

// BackupVersion is the schema version of the exported configuration document.
// Restore accepts the current version and BackupVersionLegacy (1).
const (
	BackupVersion        = 2
	BackupVersionLegacy  = 1
)

// BackupAdminScope captures per-admin resource access restored after nodes/plans
// exist.
type BackupAdminScope struct {
	AdminID    uuid.UUID   `json:"admin_id"`
	InboundIDs []uuid.UUID `json:"inbound_ids,omitempty"`
	PlanIDs    []uuid.UUID `json:"plan_ids,omitempty"`
	NodeIDs    []uuid.UUID `json:"node_ids,omitempty"`
}

// Backup is a portable snapshot of the panel's proxy configuration and operator
// accounts: the node fleet, their inbounds/outbounds/routing/balancers, reseller
// admins (with scopes and shop plans), service users (with credentials and traffic
// counters), and user→inbound bindings. It deliberately excludes time-series
// traffic metrics and audit/auth session records.
type Backup struct {
	Version    int            `json:"version"`
	ExportedAt time.Time      `json:"exported_at"`
	Roles      []*Role        `json:"roles,omitempty"`
	Admins     []*Admin       `json:"admins,omitempty"`
	AdminScopes []BackupAdminScope `json:"admin_scopes,omitempty"`
	Plans      []*Plan        `json:"plans,omitempty"`
	PortalBranding []*PortalBranding `json:"portal_branding,omitempty"`
	Nodes      []*Node        `json:"nodes"`
	Inbounds   []*Inbound     `json:"inbounds"`
	Outbounds  []*Outbound    `json:"outbounds"`
	Routing    []*RoutingRule `json:"routing"`
	Balancers  []*Balancer    `json:"balancers"`
	Users      []*User        `json:"users"`
	Bindings   []UserProxy    `json:"bindings"`
}

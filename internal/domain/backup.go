package domain

import "time"

// BackupVersion is the schema version of the exported configuration document.
// Restore refuses a document whose version it does not understand.
const BackupVersion = 1

// Backup is a portable, engine-neutral snapshot of the panel's proxy
// configuration: the node fleet, their inbounds/outbounds/routing/balancers, the
// service users (with their credentials), and the user→inbound bindings. It
// deliberately excludes time-series traffic metrics and admin/auth records.
type Backup struct {
	Version    int            `json:"version"`
	ExportedAt time.Time      `json:"exported_at"`
	Nodes      []*Node        `json:"nodes"`
	Inbounds   []*Inbound     `json:"inbounds"`
	Outbounds  []*Outbound    `json:"outbounds"`
	Routing    []*RoutingRule `json:"routing"`
	Balancers  []*Balancer    `json:"balancers"`
	Users      []*User        `json:"users"`
	Bindings   []UserProxy    `json:"bindings"`
}

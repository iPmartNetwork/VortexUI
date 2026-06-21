package domain

// RoutingPack is a named, reusable set of routing rules (plus any outbounds the
// rules reference) that can be listed, persisted, applied to a node, or embedded
// into subscription output. It is the persistent, selectable form of the
// built-in RoutingTemplate concept.
//
// ID is a string so a single type covers both kinds of pack: built-in packs use
// their Name as a stable ID and carry no database row (Builtin=true), while
// admin-defined custom packs use a UUID (in string form) and are persisted.
// Rules/Outbounds reuse the engine-neutral domain types; their node-specific
// fields (ID/NodeID) are unset on a pack and are assigned fresh when the pack is
// applied to a node.
type RoutingPack struct {
	ID          string        `json:"id"` // builtin: name; custom: uuid
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Builtin     bool          `json:"builtin"`
	Rules       []RoutingRule `json:"rules"`
	Outbounds   []Outbound    `json:"outbounds,omitempty"`
}

// RoutingPackSelection records which pack is chosen at a given scope. A global
// selection applies as the default; a subscription selection (UserID set)
// overrides it for one user. PackID is a builtin name or a custom uuid.
type RoutingPackSelection struct {
	Scope  string  `json:"scope"`             // "global" | "subscription"
	PackID string  `json:"pack_id"`           // builtin name or custom uuid
	UserID *string `json:"user_id,omitempty"` // set when scope=subscription
}

// BuiltinRoutingPacks exposes the built-in routing templates as RoutingPacks.
// Each built-in pack uses its Name as its ID and is flagged Builtin so callers
// can merge it with persisted custom packs without colliding on UUIDs.
func BuiltinRoutingPacks() []RoutingPack {
	templates := BuiltinRoutingTemplates()
	packs := make([]RoutingPack, len(templates))
	for i, t := range templates {
		packs[i] = RoutingPack{
			ID:          t.Name,
			Name:        t.Name,
			Description: t.Description,
			Category:    t.Category,
			Builtin:     true,
			Rules:       t.Rules,
			Outbounds:   t.Outbounds,
		}
	}
	return packs
}

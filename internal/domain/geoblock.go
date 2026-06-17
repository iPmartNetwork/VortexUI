package domain

// GeoPolicy defines which countries can (or cannot) connect to an inbound.
// When AllowedCountries is non-empty, only IPs from those countries are accepted.
// When BlockedCountries is non-empty, IPs from those countries are rejected.
// Both use ISO 3166-1 alpha-2 codes (e.g. "IR", "CN", "RU").
//
// This is enforced at the proxy core level via routing rules that reject
// connections not matching the geo policy before they reach the user's proxy.
type GeoPolicy struct {
	AllowedCountries []string `json:"allowed_countries,omitempty"` // whitelist (empty = all allowed)
	BlockedCountries []string `json:"blocked_countries,omitempty"` // blacklist
}

package domain

// SubSettings holds panel-configurable subscription delivery settings. It is a
// single-row config: how often subscription clients should re-fetch (and thus
// pick up live node/domain/endpoint changes) is controlled here.
type SubSettings struct {
	UpdateInterval int `json:"update_interval"` // hours; client re-fetch cadence
}

// DefaultSubSettings returns the built-in defaults.
func DefaultSubSettings() SubSettings {
	return SubSettings{UpdateInterval: 12}
}

package domain

// SubSettings holds panel-configurable subscription delivery settings. It is a
// single-row config: how often subscription clients should re-fetch (and thus
// pick up live node/domain/endpoint changes) is controlled here. It also stores
// format templates that control how subscription output (profile title, proxy
// remarks, and addresses) are rendered using format variables.
type SubSettings struct {
	UpdateInterval       int    `json:"update_interval"`        // hours; client re-fetch cadence
	ProfileTitleTemplate string `json:"profile_title_template"` // e.g. "{USERNAME} - VPN"
	RemarkTemplate       string `json:"remark_template"`        // e.g. "{PROTOCOL} - {NODE_FLAG} {NODE_NAME}"
	AddressTemplate      string `json:"address_template"`       // e.g. "{SERVER_IP}"
}

// DefaultSubSettings returns the built-in defaults.
func DefaultSubSettings() SubSettings {
	return SubSettings{
		UpdateInterval:       12,
		ProfileTitleTemplate: "{USERNAME}",
		RemarkTemplate:       "{PROTOCOL} - {NODE_NAME}",
		AddressTemplate:      "",
	}
}

package domain

// Reseller setting section keys (Settings page + sudo toggles in Edit Admin).
const (
	ResellerSettingAppearance     = "appearance"
	ResellerSettingPassword       = "password"
	ResellerSettingTOTP           = "totp"
	ResellerSettingAPITokens      = "api_tokens"
	ResellerSettingBackup         = "backup"
	ResellerSettingConfigTemplate = "config_template"
	ResellerSettingSubUpdate      = "sub_update"
	ResellerSettingIPGuard        = "ip_guard"
	ResellerSettingBranding       = "branding"
	ResellerSettingAutoBackup     = "auto_backup"
	ResellerSettingUpdate         = "update"
)

// AllResellerSettingKeys lists every Settings section sudo can grant.
func AllResellerSettingKeys() []string {
	return []string{
		ResellerSettingAppearance,
		ResellerSettingPassword,
		ResellerSettingTOTP,
		ResellerSettingAPITokens,
		ResellerSettingBackup,
		ResellerSettingConfigTemplate,
		ResellerSettingSubUpdate,
		ResellerSettingIPGuard,
		ResellerSettingBranding,
		ResellerSettingAutoBackup,
		ResellerSettingUpdate,
	}
}

// DefaultResellerSettings returns safe defaults for a new reseller.
func DefaultResellerSettings() map[string]bool {
	return map[string]bool{
		ResellerSettingAppearance:     true,
		ResellerSettingPassword:       true,
		ResellerSettingTOTP:           true,
		ResellerSettingAPITokens:      false,
		ResellerSettingBackup:         false,
		ResellerSettingConfigTemplate: false,
		ResellerSettingSubUpdate:      false,
		ResellerSettingIPGuard:        false,
		ResellerSettingBranding:       false,
		ResellerSettingAutoBackup:     false,
		ResellerSettingUpdate:         false,
	}
}

// MergeResellerSettings overlays stored values onto defaults.
func MergeResellerSettings(stored map[string]bool) map[string]bool {
	out := DefaultResellerSettings()
	for k, v := range stored {
		out[k] = v
	}
	return out
}

// ResellerSettingEnabled reports whether a section is on for this admin.
func ResellerSettingEnabled(stored map[string]bool, key string) bool {
	if stored != nil {
		if v, ok := stored[key]; ok {
			return v
		}
	}
	def := DefaultResellerSettings()
	return def[key]
}

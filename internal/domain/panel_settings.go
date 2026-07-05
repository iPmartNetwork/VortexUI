package domain

// PanelSettings holds panel-wide configuration persisted in the database.
// Replaces browser localStorage for Settings tabs (general, security, appearance, notifications, backup).
type PanelSettings struct {
	PanelName      string `json:"panel_name"`
	PanelDomain    string `json:"panel_domain"`
	SubURLTemplate string `json:"sub_url_template"`
	AutoSyncNodes  bool   `json:"auto_sync_nodes"`
	DebugMode      bool   `json:"debug_mode"`
	ClashRulesExtra string `json:"clash_rules_extra"`
	SingboxDNSExtra string `json:"singbox_dns_extra"`

	AccentColor string `json:"accent_color"`
	LogoURL     string `json:"logo_url"`
	FooterText  string `json:"footer_text"`

	IPWhitelist string `json:"ip_whitelist"`
	IPBlacklist string `json:"ip_blacklist"`

	PushNotifications    bool   `json:"push_notifications"`
	EmailAlerts          bool   `json:"email_alerts"`
	NotifyTelegramToken  string `json:"notify_telegram_token"`

	Require2FA       bool `json:"require_2fa"`
	APIAccessEnabled bool `json:"api_access_enabled"`

	AutoBackupEnabled          bool   `json:"auto_backup_enabled"`
	AutoBackupIntervalHours    int    `json:"auto_backup_interval_hours"`
	AutoBackupTelegramChatID   string `json:"auto_backup_telegram_chat_id"`
	AutoBackupS3Endpoint       string `json:"auto_backup_s3_endpoint"`
	AutoBackupS3Bucket         string `json:"auto_backup_s3_bucket"`
	AutoBackupS3AccessKey      string `json:"auto_backup_s3_access_key"`
	AutoBackupS3SecretKey      string `json:"auto_backup_s3_secret_key"`
}

// DefaultPanelSettings returns factory defaults for a fresh install.
func DefaultPanelSettings() PanelSettings {
	return PanelSettings{
		PanelName:               "VortexUI",
		SubURLTemplate:          "https://{domain}/sub/{token}",
		AutoSyncNodes:           true,
		AccentColor:             "#6366f1",
		FooterText:              "© 2026 iPmart Network. All rights reserved.",
		PushNotifications:       true,
		APIAccessEnabled:        true,
		AutoBackupEnabled:       false,
		AutoBackupIntervalHours: 24,
	}
}

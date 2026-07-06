package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
)

// PanelSettingsRepository persists panel-wide settings as JSON.
type PanelSettingsRepository interface {
	Get(ctx context.Context) (*domain.PanelSettings, error)
	Save(ctx context.Context, s *domain.PanelSettings) error
}

// PanelSettingsHooks apply runtime side-effects when settings change.
type PanelSettingsHooks struct {
	OnIPGuard func(whitelist, blacklist string)
}

type panelSettingsPayload struct {
	PanelName                *string `json:"panel_name"`
	PanelDomain              *string `json:"panel_domain"`
	SubURLTemplate           *string `json:"sub_url_template"`
	AutoSyncNodes            *bool   `json:"auto_sync_nodes"`
	DebugMode                *bool   `json:"debug_mode"`
	ClashRulesExtra          *string `json:"clash_rules_extra"`
	SingboxDNSExtra          *string `json:"singbox_dns_extra"`
	AccentColor              *string `json:"accent_color"`
	LogoURL                  *string `json:"logo_url"`
	FooterText               *string `json:"footer_text"`
	IPWhitelist              *string `json:"ip_whitelist"`
	IPBlacklist              *string `json:"ip_blacklist"`
	PushNotifications        *bool   `json:"push_notifications"`
	EmailAlerts              *bool   `json:"email_alerts"`
	NotifyTelegramToken      *string `json:"notify_telegram_token"`
	Require2FA               *bool   `json:"require_2fa"`
	APIAccessEnabled         *bool   `json:"api_access_enabled"`
	AutoBackupEnabled        *bool   `json:"auto_backup_enabled"`
	AutoBackupIntervalHours  *int    `json:"auto_backup_interval_hours"`
	AutoBackupTelegramChatID *string `json:"auto_backup_telegram_chat_id"`
	AutoBackupS3Endpoint     *string `json:"auto_backup_s3_endpoint"`
	AutoBackupS3Bucket       *string `json:"auto_backup_s3_bucket"`
	AutoBackupS3AccessKey    *string `json:"auto_backup_s3_access_key"`
	AutoBackupS3SecretKey    *string `json:"auto_backup_s3_secret_key"`
}

// PanelSettingsService manages persisted panel configuration.
type PanelSettingsService struct {
	repo  PanelSettingsRepository
	hooks PanelSettingsHooks
}

// NewPanelSettingsService wires the service.
func NewPanelSettingsService(repo PanelSettingsRepository, hooks PanelSettingsHooks) *PanelSettingsService {
	return &PanelSettingsService{repo: repo, hooks: hooks}
}

// Get returns stored settings merged with defaults.
func (s *PanelSettingsService) Get(ctx context.Context) (*domain.PanelSettings, error) {
	def := domain.DefaultPanelSettings()
	stored, err := s.repo.Get(ctx)
	if err != nil || stored == nil {
		return &def, nil
	}

	// Treat a zero-value repo payload as empty so defaults are preserved.
	if isZeroPanelSettings(stored) {
		return &def, nil
	}

	payload := panelSettingsPayload{}
	b, err := json.Marshal(stored)
	if err != nil {
		return &def, nil
	}
	if string(b) == "{}" {
		return &def, nil
	}
	if err := json.Unmarshal(b, &payload); err != nil {
		return &def, nil
	}

	out := def
	if payload.PanelName != nil {
		out.PanelName = *payload.PanelName
	}
	if payload.PanelDomain != nil {
		out.PanelDomain = *payload.PanelDomain
	}
	if payload.SubURLTemplate != nil {
		out.SubURLTemplate = *payload.SubURLTemplate
	}
	if payload.AutoSyncNodes != nil {
		out.AutoSyncNodes = *payload.AutoSyncNodes
	}
	if payload.DebugMode != nil {
		out.DebugMode = *payload.DebugMode
	}
	if payload.ClashRulesExtra != nil {
		out.ClashRulesExtra = *payload.ClashRulesExtra
	}
	if payload.SingboxDNSExtra != nil {
		out.SingboxDNSExtra = *payload.SingboxDNSExtra
	}
	if payload.AccentColor != nil {
		out.AccentColor = *payload.AccentColor
	}
	if payload.LogoURL != nil {
		out.LogoURL = *payload.LogoURL
	}
	if payload.FooterText != nil {
		out.FooterText = *payload.FooterText
	}
	if payload.IPWhitelist != nil {
		out.IPWhitelist = *payload.IPWhitelist
	}
	if payload.IPBlacklist != nil {
		out.IPBlacklist = *payload.IPBlacklist
	}
	if payload.PushNotifications != nil {
		out.PushNotifications = *payload.PushNotifications
	}
	if payload.EmailAlerts != nil {
		out.EmailAlerts = *payload.EmailAlerts
	}
	if payload.NotifyTelegramToken != nil {
		out.NotifyTelegramToken = *payload.NotifyTelegramToken
	}
	if payload.Require2FA != nil {
		out.Require2FA = *payload.Require2FA
	}
	if payload.APIAccessEnabled != nil {
		out.APIAccessEnabled = *payload.APIAccessEnabled
	}
	if payload.AutoBackupEnabled != nil {
		out.AutoBackupEnabled = *payload.AutoBackupEnabled
	}
	if payload.AutoBackupIntervalHours != nil {
		out.AutoBackupIntervalHours = *payload.AutoBackupIntervalHours
	}
	if payload.AutoBackupTelegramChatID != nil {
		out.AutoBackupTelegramChatID = *payload.AutoBackupTelegramChatID
	}
	if payload.AutoBackupS3Endpoint != nil {
		out.AutoBackupS3Endpoint = *payload.AutoBackupS3Endpoint
	}
	if payload.AutoBackupS3Bucket != nil {
		out.AutoBackupS3Bucket = *payload.AutoBackupS3Bucket
	}
	if payload.AutoBackupS3AccessKey != nil {
		out.AutoBackupS3AccessKey = *payload.AutoBackupS3AccessKey
	}
	if payload.AutoBackupS3SecretKey != nil {
		out.AutoBackupS3SecretKey = *payload.AutoBackupS3SecretKey
	}
	return &out, nil
}

// Update persists the full settings object and runs hooks.
func (s *PanelSettingsService) Update(ctx context.Context, next domain.PanelSettings) (*domain.PanelSettings, error) {
	def := domain.DefaultPanelSettings()
	if next.PanelName == "" {
		next.PanelName = def.PanelName
	}
	if next.SubURLTemplate == "" {
		next.SubURLTemplate = def.SubURLTemplate
	}
	if next.AccentColor == "" {
		next.AccentColor = def.AccentColor
	}
	if next.FooterText == "" {
		next.FooterText = def.FooterText
	}
	if next.AutoBackupIntervalHours <= 0 {
		next.AutoBackupIntervalHours = def.AutoBackupIntervalHours
	}
	if next.AutoBackupIntervalHours > 168 {
		next.AutoBackupIntervalHours = 168
	}
	if next.PanelName != "" {
		next.PanelName = trimSpaceOrDefault(next.PanelName, def.PanelName)
	}
	if next.SubURLTemplate != "" {
		next.SubURLTemplate = trimSpaceOrDefault(next.SubURLTemplate, def.SubURLTemplate)
	}
	if next.AccentColor != "" {
		next.AccentColor = trimSpaceOrDefault(next.AccentColor, def.AccentColor)
	}
	if next.FooterText != "" {
		next.FooterText = trimSpaceOrDefault(next.FooterText, def.FooterText)
	}
	if next.IPWhitelist != "" {
		next.IPWhitelist = trimSpaceOrDefault(next.IPWhitelist, "")
	}
	if next.IPBlacklist != "" {
		next.IPBlacklist = trimSpaceOrDefault(next.IPBlacklist, "")
	}
	if err := s.repo.Save(ctx, &next); err != nil {
		return nil, err
	}
	if s.hooks.OnIPGuard != nil {
		s.hooks.OnIPGuard(next.IPWhitelist, next.IPBlacklist)
	}
	return &next, nil
}

func trimSpaceOrDefault(value, fallback string) string {
	trimmed := value
	if trimmed != "" {
		trimmed = strings.TrimSpace(trimmed)
	}
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func isZeroPanelSettings(s *domain.PanelSettings) bool {
	if s == nil {
		return true
	}
	return *s == (domain.PanelSettings{})
}

// AutoBackupConfig is a snapshot for the auto-backup loop.
type AutoBackupConfig struct {
	Enabled        bool
	IntervalHours  int
	TelegramChatID string
	S3Endpoint     string
	S3Bucket       string
	S3AccessKey    string
	S3SecretKey    string
}

// Snapshot implements AutoBackupSettingsProvider.
func (s *PanelSettingsService) Snapshot(ctx context.Context) AutoBackupConfig {
	return s.AutoBackupSnapshot(ctx)
}

// AutoBackupSnapshot reads current auto-backup settings.
func (s *PanelSettingsService) AutoBackupSnapshot(ctx context.Context) AutoBackupConfig {
	cfg, err := s.Get(ctx)
	if err != nil || cfg == nil {
		return AutoBackupConfig{IntervalHours: 24}
	}
	return AutoBackupConfig{
		Enabled:        cfg.AutoBackupEnabled,
		IntervalHours:  cfg.AutoBackupIntervalHours,
		TelegramChatID: cfg.AutoBackupTelegramChatID,
		S3Endpoint:     cfg.AutoBackupS3Endpoint,
		S3Bucket:       cfg.AutoBackupS3Bucket,
		S3AccessKey:    cfg.AutoBackupS3AccessKey,
		S3SecretKey:    cfg.AutoBackupS3SecretKey,
	}
}

// SettingsJSON is used by the repository layer.
func SettingsJSON(s domain.PanelSettings) ([]byte, error) {
	return json.Marshal(s)
}

package service

import (
	"context"
	"encoding/json"

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
	stored, err := s.repo.Get(ctx)
	if err != nil {
		def := domain.DefaultPanelSettings()
		return &def, nil
	}
	out := domain.DefaultPanelSettings()
	if stored != nil {
		// Overlay stored values on defaults.
		b, _ := json.Marshal(stored)
		_ = json.Unmarshal(b, &out)
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
	if err := s.repo.Save(ctx, &next); err != nil {
		return nil, err
	}
	if s.hooks.OnIPGuard != nil {
		s.hooks.OnIPGuard(next.IPWhitelist, next.IPBlacklist)
	}
	return &next, nil
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

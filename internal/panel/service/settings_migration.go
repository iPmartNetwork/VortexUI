package service

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PanelSettings represents all exportable panel settings.
type PanelSettings struct {
	Version  string         `yaml:"version" json:"version"`
	General  map[string]any `yaml:"general" json:"general"`
	Security map[string]any `yaml:"security" json:"security"`
	Nodes    map[string]any `yaml:"nodes" json:"nodes"`
	Sub      map[string]any `yaml:"subscription" json:"subscription"`
	Notify   map[string]any `yaml:"notifications" json:"notifications"`
	Backup   map[string]any `yaml:"backup" json:"backup"`
}

// SettingsMigrationService handles exporting and importing panel settings as YAML.
type SettingsMigrationService struct{}

// NewSettingsMigrationService creates the service.
func NewSettingsMigrationService() *SettingsMigrationService {
	return &SettingsMigrationService{}
}

// Export serializes current panel settings to YAML bytes.
func (s *SettingsMigrationService) Export(ctx context.Context, settings *PanelSettings) ([]byte, error) {
	if settings.Version == "" {
		settings.Version = "1.0"
	}
	data, err := yaml.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("marshal settings: %w", err)
	}
	return data, nil
}

// ExportToFile writes settings to a YAML file.
func (s *SettingsMigrationService) ExportToFile(ctx context.Context, settings *PanelSettings, path string) error {
	data, err := s.Export(ctx, settings)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Import reads and validates YAML settings.
func (s *SettingsMigrationService) Import(ctx context.Context, data []byte) (*PanelSettings, error) {
	var settings PanelSettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal settings: %w", err)
	}

	if err := s.validate(&settings); err != nil {
		return nil, fmt.Errorf("validate settings: %w", err)
	}

	return &settings, nil
}

// ImportFromFile reads settings from a YAML file.
func (s *SettingsMigrationService) ImportFromFile(ctx context.Context, path string) (*PanelSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return s.Import(ctx, data)
}

func (s *SettingsMigrationService) validate(settings *PanelSettings) error {
	if settings.Version == "" {
		return fmt.Errorf("version field is required")
	}
	// Add more validation rules as needed
	return nil
}

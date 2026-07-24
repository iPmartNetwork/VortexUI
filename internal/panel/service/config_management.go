package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ConfigManagementService handles inbound config validation, versioning, diff,
// rollback, import/export, and core auto-update operations.
type ConfigManagementService struct {
	versionRepo port.ConfigVersionRepository
	validator   *ConfigValidator
	httpClient  *http.Client
}

// NewConfigManagementService wires dependencies.
func NewConfigManagementService(
	versionRepo port.ConfigVersionRepository,
	validator *ConfigValidator,
) *ConfigManagementService {
	return &ConfigManagementService{
		versionRepo: versionRepo,
		validator:   validator,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
	}
}

// ValidateConfig validates config and returns any errors without persisting.
func (s *ConfigManagementService) ValidateConfig(protocol domain.Protocol, network string, security domain.Security, config map[string]any) []domain.ConfigValidationError {
	return s.validator.Validate(protocol, network, security, config)
}

// GetDefaults returns valid defaults for the given combination.
func (s *ConfigManagementService) GetDefaults(protocol domain.Protocol, network string, security domain.Security) domain.ConfigDefaults {
	return s.validator.DefaultsFor(protocol, network, security)
}

// ValidateAndSave validates the config, creates a new version, and returns it.
// Returns validation errors if the config is invalid.
func (s *ConfigManagementService) ValidateAndSave(ctx context.Context, inboundID uuid.UUID, protocol domain.Protocol, network string, security domain.Security, config map[string]any, comment string, adminID *uuid.UUID) (*domain.ConfigVersion, []domain.ConfigValidationError, error) {
	errs := s.validator.Validate(protocol, network, security, config)
	if len(errs) > 0 {
		return nil, errs, nil
	}

	nextVer, err := s.versionRepo.NextVersion(ctx, inboundID)
	if err != nil {
		return nil, nil, fmt.Errorf("get next version: %w", err)
	}

	v := &domain.ConfigVersion{
		ID:         uuid.New(),
		InboundID:  inboundID,
		Version:    nextVer,
		ConfigData: config,
		Comment:    comment,
		AdminID:    adminID,
		CreatedAt:  time.Now(),
	}

	if err := s.versionRepo.Create(ctx, v); err != nil {
		return nil, nil, fmt.Errorf("create config version: %w", err)
	}

	return v, nil, nil
}

// Diff compares two versions of an inbound config and returns the changes.
// If oldVersion is 0, it compares against the version immediately preceding newVersion.
func (s *ConfigManagementService) Diff(ctx context.Context, inboundID uuid.UUID, oldVersion, newVersion int) (*domain.ConfigDiff, error) {
	newCfg, err := s.versionRepo.GetByVersion(ctx, inboundID, newVersion)
	if err != nil {
		return nil, fmt.Errorf("get new version %d: %w", newVersion, err)
	}

	if oldVersion == 0 {
		oldVersion = newVersion - 1
	}
	if oldVersion < 1 {
		// First version, diff against empty
		return &domain.ConfigDiff{
			InboundID:  inboundID,
			OldVersion: 0,
			NewVersion: newVersion,
			OldConfig:  map[string]any{},
			NewConfig:  newCfg.ConfigData,
			Changes:    computeChanges("", map[string]any{}, newCfg.ConfigData),
		}, nil
	}

	oldCfg, err := s.versionRepo.GetByVersion(ctx, inboundID, oldVersion)
	if err != nil {
		return nil, fmt.Errorf("get old version %d: %w", oldVersion, err)
	}

	return &domain.ConfigDiff{
		InboundID:  inboundID,
		OldVersion: oldVersion,
		NewVersion: newVersion,
		OldConfig:  oldCfg.ConfigData,
		NewConfig:  newCfg.ConfigData,
		Changes:    computeChanges("", oldCfg.ConfigData, newCfg.ConfigData),
	}, nil
}

// Rollback restores a previous config version by creating a new version with the old config data.
func (s *ConfigManagementService) Rollback(ctx context.Context, inboundID uuid.UUID, targetVersion int, adminID *uuid.UUID) (*domain.ConfigVersion, error) {
	target, err := s.versionRepo.GetByVersion(ctx, inboundID, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("get target version %d: %w", targetVersion, err)
	}

	nextVer, err := s.versionRepo.NextVersion(ctx, inboundID)
	if err != nil {
		return nil, fmt.Errorf("get next version: %w", err)
	}

	v := &domain.ConfigVersion{
		ID:         uuid.New(),
		InboundID:  inboundID,
		Version:    nextVer,
		ConfigData: target.ConfigData,
		Comment:    fmt.Sprintf("Rollback to version %d", targetVersion),
		AdminID:    adminID,
		CreatedAt:  time.Now(),
	}

	if err := s.versionRepo.Create(ctx, v); err != nil {
		return nil, fmt.Errorf("create rollback version: %w", err)
	}

	return v, nil
}

// ListVersions returns all config versions for an inbound.
func (s *ConfigManagementService) ListVersions(ctx context.Context, inboundID uuid.UUID) ([]*domain.ConfigVersion, error) {
	return s.versionRepo.ListForInbound(ctx, inboundID)
}

// GetLatestVersion returns the most recent config version.
func (s *ConfigManagementService) GetLatestVersion(ctx context.Context, inboundID uuid.UUID) (*domain.ConfigVersion, error) {
	return s.versionRepo.GetLatest(ctx, inboundID)
}

// ExportConfig serializes an inbound's latest config version to JSON bytes.
func (s *ConfigManagementService) ExportConfig(ctx context.Context, inboundID uuid.UUID) ([]byte, error) {
	latest, err := s.versionRepo.GetLatest(ctx, inboundID)
	if err != nil {
		return nil, fmt.Errorf("get latest version: %w", err)
	}

	export := map[string]any{
		"inbound_id": inboundID.String(),
		"version":    latest.Version,
		"config":     latest.ConfigData,
		"exported_at": time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal export: %w", err)
	}
	return data, nil
}

// ImportConfig validates and saves imported configuration JSON for an inbound.
func (s *ConfigManagementService) ImportConfig(ctx context.Context, inboundID uuid.UUID, protocol domain.Protocol, network string, security domain.Security, data []byte, adminID *uuid.UUID) (*domain.ConfigVersion, []domain.ConfigValidationError, error) {
	var imported struct {
		Config map[string]any `json:"config"`
	}
	if err := json.Unmarshal(data, &imported); err != nil {
		return nil, []domain.ConfigValidationError{{Field: "_json", Message: "invalid JSON: " + err.Error()}}, nil
	}

	if imported.Config == nil {
		return nil, []domain.ConfigValidationError{{Field: "config", Message: "config field is required"}}, nil
	}

	return s.ValidateAndSave(ctx, inboundID, protocol, network, security, imported.Config, "Imported configuration", adminID)
}

// AutoUpdateRequest contains details for a core binary auto-update.
type AutoUpdateRequest struct {
	NodeID      uuid.UUID `json:"node_id"`
	CoreType    string    `json:"core_type"`    // "xray", "sing-box"
	DownloadURL string    `json:"download_url"` // optional override; if empty, uses latest release
}

// AutoUpdateResult describes the outcome of an auto-update attempt.
type AutoUpdateResult struct {
	NodeID     uuid.UUID `json:"node_id"`
	Success    bool      `json:"success"`
	OldVersion string    `json:"old_version"`
	NewVersion string    `json:"new_version"`
	Message    string    `json:"message"`
}

// AutoUpdate downloads the latest core binary and triggers an update on the node.
// The actual node push is delegated to the node service via gRPC (out of scope here);
// this method resolves the download URL and fetches the binary metadata.
func (s *ConfigManagementService) AutoUpdate(ctx context.Context, req AutoUpdateRequest) (*AutoUpdateResult, error) {
	downloadURL := req.DownloadURL
	if downloadURL == "" {
		switch req.CoreType {
		case "xray":
			downloadURL = "https://api.github.com/repos/XTLS/Xray-core/releases/latest"
		case "sing-box":
			downloadURL = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"
		default:
			return nil, fmt.Errorf("unsupported core type: %s", req.CoreType)
		}
	}

	// Fetch release info to determine version
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return &AutoUpdateResult{
			NodeID:  req.NodeID,
			Success: false,
			Message: fmt.Sprintf("failed to fetch release info: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &AutoUpdateResult{
			NodeID:  req.NodeID,
			Success: false,
			Message: fmt.Sprintf("release API returned %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var release struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return &AutoUpdateResult{
			NodeID:  req.NodeID,
			Success: false,
			Message: fmt.Sprintf("failed to decode release info: %v", err),
		}, nil
	}

	// In production, this would trigger a gRPC call to the node agent to download
	// and replace the binary. For now, we return the resolved version.
	return &AutoUpdateResult{
		NodeID:     req.NodeID,
		Success:    true,
		NewVersion: release.TagName,
		Message:    fmt.Sprintf("Resolved latest %s release: %s", req.CoreType, release.TagName),
	}, nil
}

// --- diff computation helpers ---

func computeChanges(prefix string, old, new map[string]any) []domain.ConfigChange {
	var changes []domain.ConfigChange

	// Check for removed and modified keys
	for key, oldVal := range old {
		path := joinPath(prefix, key)
		newVal, exists := new[key]
		if !exists {
			changes = append(changes, domain.ConfigChange{
				Path:     path,
				OldValue: oldVal,
				NewValue: nil,
				Type:     "removed",
			})
			continue
		}
		// Recurse into nested maps
		oldMap, oldIsMap := oldVal.(map[string]any)
		newMap, newIsMap := newVal.(map[string]any)
		if oldIsMap && newIsMap {
			changes = append(changes, computeChanges(path, oldMap, newMap)...)
		} else if !reflect.DeepEqual(oldVal, newVal) {
			changes = append(changes, domain.ConfigChange{
				Path:     path,
				OldValue: oldVal,
				NewValue: newVal,
				Type:     "modified",
			})
		}
	}

	// Check for added keys
	for key, newVal := range new {
		if _, exists := old[key]; !exists {
			path := joinPath(prefix, key)
			changes = append(changes, domain.ConfigChange{
				Path:     path,
				OldValue: nil,
				NewValue: newVal,
				Type:     "added",
			})
		}
	}

	return changes
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + strings.TrimPrefix(key, ".")
}

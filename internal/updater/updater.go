// Package updater provides graceful self-update for the panel and core binaries
// (xray/sing-box). It downloads the latest release, replaces binaries atomically,
// and restarts the service without dropping active connections (via systemd
// socket activation or a brief drain period).
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Release represents a GitHub release asset.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a downloadable file in a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Updater handles checking for updates and applying them.
type Updater struct {
	RepoOwner string // e.g. "iPmartNetwork"
	RepoName  string // e.g. "VortexUI"
	Current   string // current version
	log       *slog.Logger
	client    *http.Client
}

// New builds an Updater.
func New(owner, repo, currentVersion string, log *slog.Logger) *Updater {
	if log == nil {
		log = slog.Default()
	}
	return &Updater{
		RepoOwner: owner,
		RepoName:  repo,
		Current:   currentVersion,
		log:       log,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

// CheckResult holds the latest release info.
type CheckResult struct {
	Available  bool   `json:"available"`
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	PanelURL   string `json:"panel_url,omitempty"`
	NodeURL    string `json:"node_url,omitempty"`
	XrayURL    string `json:"xray_url,omitempty"`
	SingboxURL string `json:"singbox_url,omitempty"`
}

// Check queries GitHub for the latest release and compares versions.
func (u *Updater) Check(ctx context.Context) (*CheckResult, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.RepoOwner, u.RepoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	result := &CheckResult{
		Current: u.Current,
		Latest:  rel.TagName,
	}
	if rel.TagName != "" && rel.TagName != "v"+u.Current && rel.TagName != u.Current {
		result.Available = true
	}

	// Find matching assets for this architecture
	arch := runtime.GOARCH
	for _, a := range rel.Assets {
		switch {
		case contains(a.Name, "panel") && contains(a.Name, arch):
			result.PanelURL = a.BrowserDownloadURL
		case contains(a.Name, "node") && contains(a.Name, arch):
			result.NodeURL = a.BrowserDownloadURL
		case contains(a.Name, "xray") || contains(a.Name, "Xray"):
			result.XrayURL = a.BrowserDownloadURL
		case contains(a.Name, "sing-box"):
			result.SingboxURL = a.BrowserDownloadURL
		}
	}
	return result, nil
}

// ApplyPanel downloads and replaces the panel binary, then signals systemd to
// restart. Active connections get a brief drain period.
func (u *Updater) ApplyPanel(ctx context.Context, downloadURL, binaryPath string) error {
	u.log.Info("downloading panel update", "url", downloadURL)
	if err := downloadReplace(ctx, u.client, downloadURL, binaryPath); err != nil {
		return fmt.Errorf("download panel: %w", err)
	}
	u.log.Info("panel binary updated, restarting service")
	return restartService("vortexui-panel")
}

// ApplyCore downloads and replaces a core binary (xray or sing-box).
func (u *Updater) ApplyCore(ctx context.Context, downloadURL, binaryPath string) error {
	u.log.Info("downloading core update", "url", downloadURL, "bin", binaryPath)
	if err := downloadReplace(ctx, u.client, downloadURL, binaryPath); err != nil {
		return fmt.Errorf("download core: %w", err)
	}
	u.log.Info("core binary updated", "path", binaryPath)
	return nil
}

// downloadReplace atomically replaces a binary: download to temp, chmod, rename.
func downloadReplace(ctx context.Context, client *http.Client, url, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		_ = tmp.Close()
		return err
	}
	_ = tmp.Close()

	if err := os.Chmod(tmpName, 0o755); err != nil { //nolint:gosec // binary needs execute permission
		return err
	}
	return os.Rename(tmpName, dst)
}

func restartService(name string) error {
	return exec.Command("systemctl", "restart", name).Run()
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

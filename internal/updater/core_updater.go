package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// CoreUpdater checks and updates xray-core and sing-box binaries automatically.
type CoreUpdater struct {
	xrayBin    string // path to xray binary
	singboxBin string // path to sing-box binary
	log        *slog.Logger
	client     *http.Client
	interval   time.Duration
}

// NewCoreUpdater builds the auto-updater.
func NewCoreUpdater(xrayBin, singboxBin string, log *slog.Logger) *CoreUpdater {
	if log == nil {
		log = slog.Default()
	}
	return &CoreUpdater{
		xrayBin:    xrayBin,
		singboxBin: singboxBin,
		log:        log,
		client:     &http.Client{Timeout: 60 * time.Second},
		interval:   24 * time.Hour,
	}
}

// Run periodically checks for core updates. Blocks until ctx cancelled.
func (u *CoreUpdater) Run(ctx context.Context) {
	// Check on startup after a delay
	time.Sleep(30 * time.Second)
	u.check(ctx)

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.check(ctx)
		}
	}
}

func (u *CoreUpdater) check(ctx context.Context) {
	// Check Xray
	if u.xrayBin != "" {
		latest, url, err := u.latestXray(ctx)
		if err == nil && url != "" {
			u.log.Info("xray latest release", "version", latest, "url", url)
			// Auto-download only if configured (for now just log)
		}
	}
	// Check Sing-box
	if u.singboxBin != "" {
		latest, url, err := u.latestSingbox(ctx)
		if err == nil && url != "" {
			u.log.Info("sing-box latest release", "version", latest, "url", url)
		}
	}
}

// LatestXray returns the latest xray version and download URL.
func (u *CoreUpdater) latestXray(ctx context.Context) (version, url string, err error) {
	rel, err := u.githubLatest(ctx, "XTLS", "Xray-core")
	if err != nil {
		return "", "", err
	}
	arch := xrayArch()
	for _, a := range rel.Assets {
		if strings.Contains(a.Name, "linux") && strings.Contains(a.Name, arch) && strings.HasSuffix(a.Name, ".zip") {
			return rel.TagName, a.BrowserDownloadURL, nil
		}
	}
	return rel.TagName, "", nil
}

// LatestSingbox returns the latest sing-box version and download URL.
func (u *CoreUpdater) latestSingbox(ctx context.Context) (version, url string, err error) {
	rel, err := u.githubLatest(ctx, "SagerNet", "sing-box")
	if err != nil {
		return "", "", err
	}
	arch := runtime.GOARCH
	for _, a := range rel.Assets {
		if strings.Contains(a.Name, "linux") && strings.Contains(a.Name, arch) && strings.HasSuffix(a.Name, ".tar.gz") {
			return rel.TagName, a.BrowserDownloadURL, nil
		}
	}
	return rel.TagName, "", nil
}

func (u *CoreUpdater) githubLatest(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
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
	return &rel, nil
}

func xrayArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64-v8a"
	default:
		return "64"
	}
}

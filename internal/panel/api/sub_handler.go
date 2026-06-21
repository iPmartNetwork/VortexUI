package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/subscription"
)

// deviceWindow is how long a device stays "active" (and occupies a slot) after
// its last subscription fetch.
const deviceWindow = 24 * time.Hour

// SubUpdateIntervalSource provides the client subscription auto-update interval
// in hours. *service.SubSettingsService satisfies it.
type SubUpdateIntervalSource interface {
	UpdateInterval(ctx context.Context) int
}

// Subscribe serves a user's subscription. It is public and authenticated solely
// by the opaque token in the path, so it must never leak which tokens exist:
// any failure returns 404. The response format is picked from the client's
// User-Agent, overridable with ?format=clash|singbox|base64.
func (h *Handlers) Subscribe(c echo.Context) error {
	token := c.Param("token")

	// If a regular browser opens the link (not a proxy client) and no explicit
	// format is requested, show the pretty info page instead of raw configs.
	if c.QueryParam("format") == "" && isBrowser(c.Request().UserAgent()) {
		return h.SubscriptionInfoPage(c)
	}

	res, err := h.Sub.Build(c.Request().Context(), token)
	if errors.Is(err, domain.ErrNotFound) || res == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "subscription failed")
	}

	// Device limiting. The token already proves ownership, so a clear 403 here is
	// safe (no token-existence oracle, unlike the 404 path above).
	if err := h.enforceDevice(c, res.User); err != nil {
		return err
	}

	format := subscription.Detect(c.Request().UserAgent())
	if q := c.QueryParam("format"); q != "" {
		format = subscription.Format(q)
	}

	body, err := subscription.RenderWith(format, res.Proxies, subscription.RenderOpts{Title: "VortexUI", Rules: res.Rules})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "render failed")
	}

	// Best-effort: record the country this user fetched from, enabling the
	// "Traffic by Country" analytics. A single fast upsert; never blocks.
	if h.Geo != nil {
		h.Geo.RecordUserIP(c.Request().Context(), res.User.ID, c.RealIP())
	}

	// Standard headers consumed by subscription clients to show quota/expiry and
	// to label and auto-refresh the profile.
	c.Response().Header().Set("Subscription-Userinfo", userInfoHeader(res.User))
	c.Response().Header().Set("Profile-Title", res.User.Username)
	interval := 12
	if h.SubSettings != nil {
		if n := h.SubSettings.UpdateInterval(c.Request().Context()); n > 0 {
			interval = n
		}
	}
	c.Response().Header().Set("Profile-Update-Interval", fmt.Sprintf("%d", interval))
	return c.Blob(http.StatusOK, format.ContentType(), body)
}

// SubscribeWireGuard serves a user's WireGuard client .conf as a downloadable
// text file. Like Subscribe it is public and authenticated solely by the opaque
// token, so any failure returns 404 (no token-existence oracle). It picks the
// first enabled WireGuard inbound the user is bound to and resolves the hosting
// node's public host the same way the subscription service does.
func (h *Handlers) SubscribeWireGuard(c echo.Context) error {
	if h.WireGuard == nil || h.NodeRepo == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	ctx := c.Request().Context()
	token := c.Param("token")

	user, err := h.Repo.GetBySubToken(ctx, token)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}

	conf, ok, err := h.wireGuardClientConfig(ctx, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "wireguard config failed")
	}
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", user.Username+".conf"))
	return c.Blob(http.StatusOK, "text/plain; charset=utf-8", []byte(conf))
}

// wireGuardClientConfig resolves the user's first enabled WireGuard inbound and
// renders their client .conf. ok is false (with err nil) when the user has no
// usable WireGuard inbound (WG not wired, no enabled WG inbound, or the hosting
// node can't be resolved), so callers can render gracefully or return 404. A
// non-nil err is reserved for a genuine render failure (maps to 500).
func (h *Handlers) wireGuardClientConfig(ctx context.Context, user *domain.User) (string, bool, error) {
	if h.WireGuard == nil || h.NodeRepo == nil {
		return "", false, nil
	}
	inbounds, err := h.Repo.InboundsFor(ctx, user.ID)
	if err != nil {
		return "", false, nil
	}
	var wgInbound *domain.Inbound
	for i := range inbounds {
		if inbounds[i].Enabled && inbounds[i].Protocol == domain.ProtoWireGuard {
			wgInbound = &inbounds[i]
			break
		}
	}
	if wgInbound == nil {
		return "", false, nil
	}

	node, err := h.NodeRepo.GetByID(ctx, wgInbound.NodeID)
	if err != nil || node == nil {
		return "", false, nil
	}
	host := node.Endpoint
	if host == "" {
		host = hostOf(node.Address)
	}
	if host == "" {
		return "", false, nil
	}

	conf, err := h.WireGuard.ClientConfig(ctx, *wgInbound, user, host)
	if err != nil {
		return "", false, err
	}
	return conf, true, nil
}

// isBrowser returns true if the UA looks like a standard web browser (not a
// proxy client like clash, sing-box, v2ray, etc.).
func isBrowser(ua string) bool {
	// Proxy clients typically identify themselves clearly.
	proxyClients := []string{
		"clash", "mihomo", "sing-box", "singbox", "v2ray", "xray",
		"quantumult", "surge", "shadowrocket", "hiddify", "nekobox",
		"nekoray", "v2rayng", "v2rayn", "streisand", "karing",
	}
	lower := strings.ToLower(ua)
	for _, client := range proxyClients {
		if strings.Contains(lower, client) {
			return false
		}
	}
	// Common browser signatures
	browsers := []string{"mozilla", "chrome", "safari", "firefox", "edge", "opera"}
	for _, b := range browsers {
		if strings.Contains(lower, b) {
			return true
		}
	}
	return false
}

// enforceDevice applies both device controls: an explicit HWID allowlist (if the
// user has one) and a rolling device-count cap (if set and a tracker is wired).
func (h *Handlers) enforceDevice(c echo.Context, u *domain.User) error {
	device := deviceID(c)

	if len(u.AllowedHWIDs) > 0 && !contains(u.AllowedHWIDs, device) {
		return echo.NewHTTPError(http.StatusForbidden, "device not authorized")
	}

	if u.DeviceLimit > 0 && h.Devices != nil {
		ok, err := h.Devices.Allow(c.Request().Context(), u.ID.String(), device, u.DeviceLimit, deviceWindow)
		if err != nil {
			return nil // fail open: a tracker error must not lock out a valid user
		}
		if !ok {
			return echo.NewHTTPError(http.StatusForbidden, "device limit reached")
		}
	}
	return nil
}

// deviceID derives a stable device identifier from the client-supplied header,
// falling back to the source IP when no device header is present.
func deviceID(c echo.Context) string {
	for _, hdr := range []string{"X-Device-Id", "X-Hwid", "X-Device-ID"} {
		if v := c.Request().Header.Get(hdr); v != "" {
			return v
		}
	}
	return c.RealIP()
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func userInfoHeader(u *domain.User) string {
	var expire int64
	if u.ExpireAt != nil {
		expire = u.ExpireAt.Unix()
	}
	// upload is always reported as 0; we account a single combined counter.
	return fmt.Sprintf("upload=0; download=%d; total=%d; expire=%d",
		u.UsedTraffic, u.DataLimit, expire)
}

// hostOf extracts the host from a "host:port" node address, tolerating a bare
// host (mirrors the subscription service's resolver).
func hostOf(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

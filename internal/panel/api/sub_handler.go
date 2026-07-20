package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
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

	// ISP hint: when the client (or frontend) passes ?isp=mci, auto-apply the
	// ISP-specific TLS tricks profile to proxies that lack a per-inbound profile.
	// When no explicit hint is given, auto-detect ISP from the client IP.
	ispHint := c.QueryParam("isp")
	if ispHint == "" && h.Geo != nil {
		ispHint = h.Geo.DetectISP(c.RealIP())
	}
	if ispHint != "" {
		service.ApplyISPHint(res.Proxies, ispHint)
		// Reorder protocol groups based on ISP-preferred protocol ordering so
		// clients try the ISP-optimized protocol first within each group.
		h.Sub.ReorderGroupsByISP(c.Request().Context(), res.Groups, res.Proxies, ispHint)
		// Smart Config: apply per-ISP optimal anti-DPI settings (fragment, mux,
		// padding, ECH, fingerprint) to proxies that lack a per-inbound profile.
		service.ApplySmartProfiles(res.Proxies, domain.ISPPreset(service.NormalizeISP(ispHint)), domain.CDNNone)
		// Smart Mux: apply ISP-optimized multiplexing settings (protocol, concurrency,
		// padding, XUDP) to all proxies that have mux enabled.
		service.ApplySmartMux(res.Proxies, domain.ISPPreset(service.NormalizeISP(ispHint)))
		// Quality Score: compute per-proxy quality and reorder by score descending.
		service.ComputeQualityScores(res.Proxies, domain.ISPPreset(service.NormalizeISP(ispHint)))
		// Dynamic SNI: assign rotated SNIs from ISP-specific pool to REALITY proxies.
		service.ApplyDynamicSNI(res.Proxies, domain.ISPPreset(service.NormalizeISP(ispHint)))
		// Transport Optimization: obfuscate gRPC service names, fix paths, set authority headers.
		service.ApplyTransportOptimization(res.Proxies, domain.ISPPreset(service.NormalizeISP(ispHint)))
	}

	format := subscription.Detect(c.Request().UserAgent())
	if q := c.QueryParam("format"); q != "" {
		format = subscription.Format(q)
	}

	body, err := subscription.RenderWith(format, res.Proxies, subscription.RenderOpts{Title: "VortexUI", Rules: res.Rules, Groups: res.Groups})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "render failed")
	}

	// Best-effort: record the country this user fetched from, enabling the
	// "Traffic by Country" analytics. A single fast upsert; never blocks.
	if h.Geo != nil {
		h.Geo.RecordUserIP(c.Request().Context(), res.User.ID, c.RealIP())
	}

	// Cache subscription output for 60s to reduce DB pressure on rapid client polls.
	c.Response().Header().Set("Cache-Control", "private, max-age=60")
	c.Response().Header().Set("X-Subscription-Cached", "false")

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

// --- Public switch event reporting ---

type reportSwitchRequest struct {
	SourceProtocol string `json:"source_protocol"`
	TargetProtocol string `json:"target_protocol"`
	NodeID         string `json:"node_id"`
	ISP            string `json:"isp"`
}

// ReportSwitch is a public endpoint that proxy clients can hit after an
// auto-protocol switch. Authenticated solely by the subscription token (same as
// /sub/:token). Returns 404 for invalid tokens (no oracle). Clients report the
// protocols they switched between; the ISP is auto-detected from the client IP
// if not provided.
func (h *Handlers) ReportSwitch(c echo.Context) error {
	if h.SwitchEvents == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	token := c.Param("token")
	ctx := c.Request().Context()

	user, err := h.Repo.GetBySubToken(ctx, token)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}

	var req reportSwitchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SourceProtocol == "" || req.TargetProtocol == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "source_protocol and target_protocol required")
	}

	// Auto-detect ISP from client IP if not explicitly provided.
	isp := req.ISP
	if isp == "" && h.Geo != nil {
		isp = h.Geo.DetectISP(c.RealIP())
	}

	e := &domain.SwitchEvent{
		UserID:         user.ID,
		SourceProtocol: req.SourceProtocol,
		TargetProtocol: req.TargetProtocol,
		ISP:            isp,
	}
	// NodeID is optional; parse if provided.
	if req.NodeID != "" {
		if id, perr := uuid.Parse(req.NodeID); perr == nil {
			e.NodeID = id
		}
	}

	if err := h.SwitchEvents.Record(ctx, e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}

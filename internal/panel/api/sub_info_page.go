package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/subscription"
)

// qrDataURI encodes content into a PNG QR code entirely on the server and
// returns it as a base64 data URI. Generating QR codes locally avoids leaking
// sensitive payloads (e.g. the WireGuard client private key inside a .conf) to
// any third-party QR rendering service.
func qrDataURI(content string, size int) string {
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
}

// SubscriptionInfoPage renders a beautiful user-facing HTML page showing their
// subscription status, traffic usage, devices, configs and QR codes.
func (h *Handlers) SubscriptionInfoPage(c echo.Context) error {
	token := c.Param("token")
	res, err := h.Sub.Build(c.Request().Context(), token)
	if errors.Is(err, domain.ErrNotFound) || res == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "subscription failed")
	}

	u := res.User
	base := c.Scheme() + "://" + c.Request().Host
	subURL := base + "/sub/" + token

	// Build share links
	var links []string
	for _, p := range res.Proxies {
		if l := subscription.ShareLink(p); l != "" {
			links = append(links, l)
		}
	}

	// Calculate usage
	usedGB := float64(u.UsedTraffic) / (1024 * 1024 * 1024)
	limitGB := float64(u.DataLimit) / (1024 * 1024 * 1024)
	usedPercent := 0.0
	if u.DataLimit > 0 {
		usedPercent = math.Min(100, float64(u.UsedTraffic)/float64(u.DataLimit)*100)
	}

	// Days remaining
	daysLeft := -1
	if u.ExpireAt != nil {
		daysLeft = int(time.Until(*u.ExpireAt).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
	}

	// Device count
	deviceCount := 0
	if h.Online != nil {
		n, err := h.Online.Online(c.Request().Context(), u.ID.String(), deviceWindow)
		if err == nil {
			deviceCount = n
		}
	}

	// WireGuard: surface a downloadable client .conf + a QR of the conf text
	// only when the user is bound to an enabled WireGuard inbound. Reuses the
	// same helper that backs GET /sub/:token/wireguard.
	var hasWG bool
	var wgConf, wgURL string
	var wgQR template.URL
	if conf, ok, wgErr := h.wireGuardClientConfig(c.Request().Context(), u); wgErr == nil && ok {
		hasWG = true
		wgConf = conf
		wgURL = subURL + "/wireguard"
		// The QR encodes the .conf TEXT itself (so WireGuard apps can import it),
		// generated server-side so the client private key inside the conf never
		// leaves this server.
		// qrDataURI returns a fixed "data:image/png;base64,<base64 bytes>"
		// string with no HTML, so the template.URL conversion is safe.
		wgQR = template.URL(qrDataURI(conf, 240)) //nolint:gosec // G203: trusted server-generated data URI, not HTML
	}

	data := subInfoData{
		Username:      u.Username,
		Status:        string(u.Status),
		UsedGB:        fmt.Sprintf("%.2f", usedGB),
		LimitGB:       fmt.Sprintf("%.2f", limitGB),
		UsedPercent:   fmt.Sprintf("%.0f", usedPercent),
		DaysLeft:      daysLeft,
		DeviceCount:   deviceCount,
		DeviceLimit:   u.DeviceLimit,
		SubURL:        subURL,
		SubQR:         template.URL(qrDataURI(subURL, 240)), //nolint:gosec // G203: trusted server-generated data URI, not HTML
		ClashURL:      subURL + "?format=clash",
		SingboxURL:    subURL + "?format=singbox",
		Base64URL:     subURL + "?format=base64",
		XrayURL:       subURL + "?format=xray",
		OutlineURL:    subURL + "?format=outline",
		LinksURL:      subURL + "?format=links",
		Links:         links,
		ConfigCount:   len(links),
		HasWireGuard:  hasWG,
		WireGuardConf: wgConf,
		WireGuardURL:  wgURL,
		WireGuardQR:   wgQR,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return subInfoTmpl.Execute(c.Response().Writer, data)
}

type subInfoData struct {
	Username    string
	Status      string
	UsedGB      string
	LimitGB     string
	UsedPercent string
	DaysLeft    int
	DeviceCount int
	DeviceLimit int
	SubURL      string
	SubQR       template.URL // server-generated QR (data URI) of the subscription URL
	ClashURL    string
	SingboxURL  string
	Base64URL   string
	XrayURL     string
	OutlineURL  string
	LinksURL    string
	Links       []string
	ConfigCount int

	// WireGuard (optional): present only when the user has an enabled WG inbound.
	HasWireGuard  bool
	WireGuardConf string // raw client .conf text
	WireGuardURL  string       // download endpoint (<SubURL>/wireguard)
	WireGuardQR   template.URL // server-generated QR (data URI) encoding the .conf text
}

var subInfoTmpl = template.Must(template.New("subinfo").Parse(subInfoHTML))

// SubscriptionUsage returns the user's traffic time-series (last 7 days) as
// JSON, authenticated solely by the subscription token. This powers the
// traffic chart on the public subscription info page.
func (h *Handlers) SubscriptionUsage(c echo.Context) error {
	token := c.Param("token")
	res, err := h.Sub.Build(c.Request().Context(), token)
	if err != nil || res == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if h.Traffic == nil {
		return c.JSON(http.StatusOK, echo.Map{"points": []any{}})
	}
	now := time.Now()
	q := port.SeriesQuery{
		FromUnix: now.Add(-7 * 24 * time.Hour).Unix(),
		ToUnix:   now.Unix(),
		Bucket:   "1d",
	}
	points, err := h.Traffic.UsageSeries(c.Request().Context(), res.User.ID, q)
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"points": []any{}})
	}
	return c.JSON(http.StatusOK, echo.Map{"points": points})
}

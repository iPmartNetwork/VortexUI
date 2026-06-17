package api

import (
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/subscription"
)

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

	data := subInfoData{
		Username:     u.Username,
		Status:       string(u.Status),
		UsedGB:       fmt.Sprintf("%.2f", usedGB),
		LimitGB:      fmt.Sprintf("%.2f", limitGB),
		UsedPercent:  fmt.Sprintf("%.0f", usedPercent),
		DaysLeft:     daysLeft,
		DeviceCount:  deviceCount,
		DeviceLimit:  u.DeviceLimit,
		SubURL:       subURL,
		ClashURL:     subURL + "?format=clash",
		SingboxURL:   subURL + "?format=singbox",
		Base64URL:    subURL + "?format=base64",
		Links:        links,
		ConfigCount:  len(links),
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
	ClashURL    string
	SingboxURL  string
	Base64URL   string
	Links       []string
	ConfigCount int
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

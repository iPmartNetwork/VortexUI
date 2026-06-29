package api

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
)

// SubscriptionShop renders a server-side HTML page showing available plans so
// that existing users (identified by their sub token in the URL) can purchase a
// renewal or upgrade directly. No authentication is required beyond the opaque
// token — same model as /sub/:token/info.
func (h *Handlers) SubscriptionShop(c echo.Context) error {
	token := c.Param("token")

	// Resolve user by subscription token.
	user, err := h.Repo.GetBySubToken(c.Request().Context(), token)
	if errors.Is(err, domain.ErrNotFound) || user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "lookup failed")
	}

	// Fetch enabled plans.
	plans, err := h.Plans.ListPlans(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "plans unavailable")
	}
	var enabled []*domain.Plan
	for _, p := range plans {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}

	data := shopPageData{
		Username: user.Username,
		Token:    token,
		Plans:    enabled,
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return shopTmpl.Execute(c.Response().Writer, data)
}

type shopPageData struct {
	Username string
	Token    string
	Plans    []*domain.Plan
}

var shopTmpl = template.Must(template.New("shop").Funcs(template.FuncMap{
	"dataGB": func(b int64) string {
		if b == 0 {
			return "Unlimited"
		}
		gb := float64(b) / (1024 * 1024 * 1024)
		if gb == float64(int(gb)) {
			return fmt.Sprintf("%d GB", int(gb))
		}
		return fmt.Sprintf("%.1f GB", gb)
	},
}).Parse(shopHTML))

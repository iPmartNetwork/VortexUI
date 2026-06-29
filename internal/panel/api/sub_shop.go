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

	// Load per-reseller payment config.
	data := shopPageData{
		Username: user.Username,
		Token:    token,
		Plans:    enabled,
	}

	if user.AdminID != nil && h.ResellerPayment != nil {
		cfg, err := h.ResellerPayment.GetPaymentConfig(c.Request().Context(), *user.AdminID)
		if err == nil && cfg != nil {
			data.EnabledMethods = cfg.EnabledMethods
			data.CardNumber = cfg.CardNumber
			data.CardHolder = cfg.CardHolder
			data.CardBank = cfg.CardBank
			data.CryptoAddresses = cfg.CryptoAddresses
			data.ManualInstructions = cfg.ManualInstructions
			// Compute convenience flags.
			for _, m := range cfg.EnabledMethods {
				switch m {
				case "zarinpal":
					data.HasZarinpal = cfg.ZarinpalMerchantID != "" || h.ZarinPal != nil
				case "card_to_card":
					data.HasCardToCard = cfg.CardNumber != ""
				case "crypto":
					data.HasCrypto = len(cfg.CryptoAddresses) > 0
				}
			}
		}
	} else {
		// No reseller config: fall back to global gateways.
		if h.ZarinPal != nil {
			data.HasZarinpal = true
			data.EnabledMethods = append(data.EnabledMethods, "zarinpal")
		}
		if h.NowPayments != nil {
			data.HasCrypto = true
			data.EnabledMethods = append(data.EnabledMethods, "nowpayments")
		}
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return shopTmpl.Execute(c.Response().Writer, data)
}

type shopPageData struct {
	Username           string
	Token              string
	Plans              []*domain.Plan
	EnabledMethods     []string
	CardNumber         string
	CardHolder         string
	CardBank           string
	CryptoAddresses    map[string]string
	ManualInstructions string
	HasZarinpal        bool
	HasCardToCard      bool
	HasCrypto          bool
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

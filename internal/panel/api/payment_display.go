package api

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// billingSettingsForReseller returns card/crypto display info for deposits and
// portal checkout. Per-reseller payment config takes precedence; otherwise the
// global billing_settings row is used.
func billingSettingsForReseller(
	ctx context.Context,
	rp *service.ResellerPaymentService,
	wb *service.WalletBillingService,
	adminID *uuid.UUID,
) (*domain.BillingSettings, error) {
	if adminID != nil && rp != nil {
		cfg, err := rp.GetPaymentConfig(ctx, *adminID)
		if err != nil {
			return nil, err
		}
		if cfg != nil && cfg.CardNumber != "" {
			crypto := cfg.CryptoAddresses
			if crypto == nil {
				crypto = map[string]string{}
			}
			return &domain.BillingSettings{
				CardNumber:         cfg.CardNumber,
				CardHolder:         cfg.CardHolder,
				CardBank:           cfg.CardBank,
				CryptoAddresses:    crypto,
				ManualInstructions: cfg.ManualInstructions,
			}, nil
		}
	}
	if wb != nil {
		return wb.GetBillingSettings(ctx)
	}
	return &domain.BillingSettings{CryptoAddresses: map[string]string{}}, nil
}

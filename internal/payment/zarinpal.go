// Package payment integrates external payment gateways for plan purchases.
package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Gateway is the interface every payment provider implements.
type Gateway interface {
	// CreatePayment initiates a payment and returns the redirect URL for the user.
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
	// VerifyPayment confirms a payment after the user returns from the gateway.
	VerifyPayment(ctx context.Context, authority string, amount int64) (*VerifyResponse, error)
}

// PaymentRequest is the input to create a new payment.
type PaymentRequest struct {
	Amount      int64  // smallest unit (toman for ZarinPal, cents for crypto)
	Description string
	CallbackURL string
	Email       string // optional
	Mobile      string // optional
}

// PaymentResponse is returned after creating a payment.
type PaymentResponse struct {
	Authority   string // gateway-specific payment ID
	RedirectURL string // send the user here to pay
}

// VerifyResponse is returned after verifying a payment.
type VerifyResponse struct {
	RefID  string // gateway reference ID
	Status string // "OK" or error
}

// --- ZarinPal ---

const (
	zarinpalRequestURL = "https://api.zarinpal.com/pg/v4/payment/request.json"
	zarinpalVerifyURL  = "https://api.zarinpal.com/pg/v4/payment/verify.json"
	zarinpalGatewayURL = "https://www.zarinpal.com/pg/StartPay/"
)

// ZarinPal implements Gateway for the ZarinPal payment provider (Iran).
type ZarinPal struct {
	merchantID string
	client     *http.Client
}

// NewZarinPal creates a ZarinPal gateway.
func NewZarinPal(merchantID string) *ZarinPal {
	return &ZarinPal{
		merchantID: merchantID,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

func (z *ZarinPal) CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	body := map[string]any{
		"merchant_id":  z.merchantID,
		"amount":       req.Amount,
		"description":  req.Description,
		"callback_url": req.CallbackURL,
	}
	if req.Email != "" {
		body["metadata"] = map[string]string{"email": req.Email, "mobile": req.Mobile}
	}
	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, zarinpalRequestURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := z.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Authority string `json:"authority"`
			Code      int    `json:"code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("zarinpal: parse response: %w", err)
	}
	if result.Data.Code != 100 {
		return nil, fmt.Errorf("zarinpal: code %d", result.Data.Code)
	}
	return &PaymentResponse{
		Authority:   result.Data.Authority,
		RedirectURL: zarinpalGatewayURL + result.Data.Authority,
	}, nil
}

func (z *ZarinPal) VerifyPayment(ctx context.Context, authority string, amount int64) (*VerifyResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"merchant_id": z.merchantID,
		"authority":   authority,
		"amount":      amount,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, zarinpalVerifyURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := z.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Code  int    `json:"code"`
			RefID string `json:"ref_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if result.Data.Code != 100 && result.Data.Code != 101 {
		return nil, fmt.Errorf("zarinpal verify: code %d", result.Data.Code)
	}
	return &VerifyResponse{RefID: result.Data.RefID, Status: "OK"}, nil
}

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

const nowPaymentsAPI = "https://api.nowpayments.io/v1"

// NowPayments implements Gateway for NOWPayments (crypto: USDT, BTC, TON, etc.).
type NowPayments struct {
	apiKey string
	client *http.Client
}

// NewNowPayments creates a NowPayments gateway.
func NewNowPayments(apiKey string) *NowPayments {
	return &NowPayments{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (n *NowPayments) CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	body := map[string]any{
		"price_amount":      req.Amount,
		"price_currency":    "usd",
		"pay_currency":      "usdttrc20", // default to USDT TRC20
		"order_description": req.Description,
		"ipn_callback_url":  req.CallbackURL,
	}
	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, nowPaymentsAPI+"/payment", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", n.apiKey)
	resp, err := n.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		PaymentID   string `json:"payment_id"`
		InvoiceURL  string `json:"invoice_url"`
		PayAddress  string `json:"pay_address"`
		PayAmount   float64 `json:"pay_amount"`
		PayCurrency string `json:"pay_currency"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("nowpayments: parse: %w", err)
	}
	if result.PaymentID == "" {
		return nil, fmt.Errorf("nowpayments: no payment_id in response: %s", string(respBody))
	}
	redirectURL := result.InvoiceURL
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("https://nowpayments.io/payment/?iid=%s", result.PaymentID)
	}
	return &PaymentResponse{
		Authority:   result.PaymentID,
		RedirectURL: redirectURL,
	}, nil
}

func (n *NowPayments) VerifyPayment(ctx context.Context, paymentID string, _ int64) (*VerifyResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/payment/%s", nowPaymentsAPI, paymentID), nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", n.apiKey)
	resp, err := n.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		PaymentStatus string `json:"payment_status"`
		PaymentID     string `json:"payment_id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if result.PaymentStatus == "finished" || result.PaymentStatus == "confirmed" {
		return &VerifyResponse{RefID: result.PaymentID, Status: "OK"}, nil
	}
	return nil, fmt.Errorf("nowpayments: status=%s", result.PaymentStatus)
}

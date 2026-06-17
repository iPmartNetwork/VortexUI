// Package payment — IPN (Instant Payment Notification) handler.
// Gateways call this endpoint asynchronously when a payment status changes,
// so we never miss a successful payment even if the user doesn't redirect back.
package payment

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// IPNHandler processes incoming payment notifications from gateways.
type IPNHandler struct {
	NowPaymentsIPNSecret string // HMAC secret for NowPayments signature verification
	OnPaymentConfirmed   func(gatewayID string, gateway string) // callback when payment is confirmed
}

// NowPaymentsIPN handles NowPayments IPN webhook POST.
// Verifies HMAC-SHA512 signature, checks status, and calls OnPaymentConfirmed.
func (h *IPNHandler) NowPaymentsIPN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// Verify signature if secret is configured
	if h.NowPaymentsIPNSecret != "" {
		sig := r.Header.Get("x-nowpayments-sig")
		if !h.verifyNowPaymentsSig(body, sig) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var payload struct {
		PaymentID     json.Number `json:"payment_id"`
		PaymentStatus string      `json:"payment_status"`
		OrderID       string      `json:"order_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Only process confirmed/finished payments
	status := strings.ToLower(payload.PaymentStatus)
	if status == "finished" || status == "confirmed" || status == "sending" {
		if h.OnPaymentConfirmed != nil {
			h.OnPaymentConfirmed(payload.PaymentID.String(), "nowpayments")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ZarinPalIPN handles ZarinPal's callback verification.
// ZarinPal doesn't have a traditional IPN — it redirects with Authority param.
// This endpoint can be used as an alternative callback receiver.
func (h *IPNHandler) ZarinPalIPN(w http.ResponseWriter, r *http.Request) {
	authority := r.URL.Query().Get("Authority")
	status := r.URL.Query().Get("Status")

	if status == "OK" && authority != "" {
		if h.OnPaymentConfirmed != nil {
			h.OnPaymentConfirmed(authority, "zarinpal")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *IPNHandler) verifyNowPaymentsSig(body []byte, signature string) bool {
	if signature == "" {
		return false
	}
	// NowPayments uses HMAC-SHA512 with sorted JSON keys
	mac := hmac.New(sha512.New, []byte(h.NowPaymentsIPNSecret))
	// Sort JSON keys for consistent hashing
	var obj map[string]any
	if json.Unmarshal(body, &obj) != nil {
		return false
	}
	sorted, _ := json.Marshal(obj)
	mac.Write(sorted)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

package paystack

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
)

type WebhookEvent struct {
	Event string      `json:"event"`
	Data  WebhookData `json:"data"`
}

type WebhookData struct {
	ID              int64    `json:"id"`
	Reference       string   `json:"reference"`
	Amount          int64    `json:"amount"`
	Status          string   `json:"status"`
	PaidAt          string   `json:"paid_at"`
	Channel         string   `json:"channel"`
	Currency        string   `json:"currency"`
	Customer        Customer `json:"customer"`
	GatewayResponse string   `json:"gateway_response"`
}

type Customer struct {
	Email string `json:"email"`
}

func ValidateWebhookSignature(r *http.Request, secretKey string) ([]byte, bool) {
	signature := r.Header.Get("X-Paystack-Signature")
	if signature == "" {
		return nil, false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, false
	}

	hash := hmac.New(sha512.New, []byte(secretKey))
	hash.Write(body)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))

	return body, hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

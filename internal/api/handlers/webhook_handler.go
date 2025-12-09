package handlers

import (
	"net/http"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
	"github.com/BerylCAtieno/paystack-wallet/internal/paystack"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"
	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	walletService  *wallet.Service
	paystackSecret string
}

func NewWebhookHandler(walletService *wallet.Service, paystackSecret string) *WebhookHandler {
	return &WebhookHandler{
		walletService:  walletService,
		paystackSecret: paystackSecret,
	}
}

// HandlePaystackWebhook now accepts *gin.Context
func (h *WebhookHandler) HandlePaystackWebhook(c *gin.Context) {
	body, valid := paystack.ValidateWebhookSignature(c.Request, h.paystackSecret)
	if !valid {
		utils.RespondError(c, http.StatusUnauthorized, "invalid signature")
		return
	}

	// Parse webhook event
	event, err := paystack.ParseWebhookEvent(body)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	// Handle charge.success event
	if event.Event == "charge.success" {
		if event.Data.Status == "success" {
			// Complete deposit
			if err := h.walletService.CompleteDeposit(event.Data.Reference); err != nil {
				utils.RespondError(c, http.StatusInternalServerError, "failed to complete deposit")
				return
			}
		}
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"status": true,
	})
}

// GetDepositStatus now accepts *gin.Context
func (h *WebhookHandler) GetDepositStatus(c *gin.Context) {
	reference := c.Query("reference")
	if reference == "" {
		utils.RespondError(c, http.StatusBadRequest, "reference is required")
		return
	}

	// This is a read-only endpoint - does not credit wallet
	// Only the webhook credits wallets
	utils.RespondSuccess(c, map[string]interface{}{
		"message": "Use webhook for real-time updates. This endpoint is for manual verification only.",
	})
}

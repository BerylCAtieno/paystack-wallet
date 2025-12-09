package handlers

import (
	"github.com/BerylCAtieno/paystack-wallet/internal/api/middleware"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"
	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	authService *auth.Service
}

func NewAPIKeyHandler(authService *auth.Service) *APIKeyHandler {
	return &APIKeyHandler{authService: authService}
}

type CreateAPIKeyRequest struct {
	Name        string              `json:"name"`
	Permissions []auth.Permission   `json:"permissions"`
	Expiry      auth.ExpiryDuration `json:"expiry"`
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		utils.RespondError(c, 401, "user not authenticated")
		return
	}

	var req CreateAPIKeyRequest
	if err := c.BindJSON(&req); err != nil {
		utils.RespondError(c, 400, "invalid request body")
		return
	}

	// Validate expiry
	validExpiry := map[auth.ExpiryDuration]bool{
		auth.Expiry1Hour:  true,
		auth.Expiry1Day:   true,
		auth.Expiry1Month: true,
		auth.Expiry1Year:  true,
	}
	if !validExpiry[req.Expiry] {
		utils.RespondError(c, 400, "invalid expiry duration. Use: 1H, 1D, 1M, 1Y")
		return
	}

	apiKey, rawKey, err := h.authService.CreateAPIKey(userID, req.Name, req.Permissions, req.Expiry)
	if err != nil {
		if err == auth.ErrMaxAPIKeysReached {
			utils.RespondError(c, 400, err.Error())
			return
		}
		utils.RespondError(c, 500, "failed to create API key")
		return
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"api_key":    rawKey,
		"expires_at": apiKey.ExpiresAt,
	})
}

type RolloverAPIKeyRequest struct {
	ExpiredKeyID string              `json:"expired_key_id"`
	Expiry       auth.ExpiryDuration `json:"expiry"`
}

func (h *APIKeyHandler) RolloverAPIKey(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		utils.RespondError(c, 401, "user not authenticated")
		return
	}

	var req RolloverAPIKeyRequest
	if err := c.BindJSON(&req); err != nil {
		utils.RespondError(c, 400, "invalid request body")
		return
	}

	apiKey, rawKey, err := h.authService.RolloverAPIKey(userID, req.ExpiredKeyID, req.Expiry)
	if err != nil {
		if err == auth.ErrKeyNotExpired {
			utils.RespondError(c, 400, "key is not expired")
			return
		}
		utils.RespondError(c, 400, err.Error())
		return
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"api_key":    rawKey,
		"expires_at": apiKey.ExpiresAt,
	})
}

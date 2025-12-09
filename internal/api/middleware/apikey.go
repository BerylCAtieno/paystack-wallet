package middleware

import (
	"net/http"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"
	"github.com/gin-gonic/gin"
)

const (
	APIKeyKey       = "api_key"
	APIKeyUserIDKey = "api_key_user_id"
)

// APIKeyAuth is a Gin middleware for API key authentication
func APIKeyAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			utils.RespondError(c, http.StatusUnauthorized, "missing api key")
			c.Abort()
			return
		}

		key, err := authService.ValidateAPIKey(apiKey)
		if err != nil {
			utils.RespondError(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set(APIKeyKey, key)
		c.Set(APIKeyUserIDKey, key.UserID)
		c.Next()
	}
}

// GetAPIKey retrieves the APIKey object from Gin context
func GetAPIKey(c *gin.Context) *auth.APIKey {
	val, exists := c.Get(APIKeyKey)
	if !exists {
		return nil
	}
	key, ok := val.(*auth.APIKey)
	if !ok {
		return nil
	}
	return key
}

// GetAPIKeyUserID retrieves the authenticated user ID from Gin context
func GetAPIKeyUserID(c *gin.Context) string {
	val, exists := c.Get(APIKeyUserIDKey)
	if !exists {
		return ""
	}
	userID, ok := val.(string)
	if !ok {
		return ""
	}
	return userID
}

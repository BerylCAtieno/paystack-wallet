package middleware

import (
	"net/http"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"
	"github.com/gin-gonic/gin"
)

// FlexibleAuthGin allows either JWT or API Key authentication for Gin
func FlexibleAuthGin(jwtSecret string, authService *auth.Service, requiredPermission ...auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT first
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			JWTAuth(jwtSecret)(c)
			if c.IsAborted() {
				return
			}
			c.Next()
			return
		}

		// Try API Key
		if apiKey := c.GetHeader("x-api-key"); apiKey != "" {
			key, err := authService.ValidateAPIKey(apiKey)
			if err != nil {
				utils.RespondError(c, http.StatusUnauthorized, err.Error())
				c.Abort()
				return
			}

			// Check required permissions
			for _, perm := range requiredPermission {
				if !key.HasPermission(perm) {
					utils.RespondError(c, http.StatusForbidden, "insufficient permissions")
					c.Abort()
					return
				}
			}

			c.Set(APIKeyKey, key)
			c.Set(APIKeyUserIDKey, key.UserID)
			c.Next()
			return
		}

		utils.RespondError(c, http.StatusUnauthorized, "authentication required")
		c.Abort()
	}
}

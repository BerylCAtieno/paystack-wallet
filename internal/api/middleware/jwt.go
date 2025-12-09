package middleware

import (
	"net/http"
	"strings"

	"github.com/BerylCAtieno/paystack-wallet/internal/security"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"
	"github.com/gin-gonic/gin"
)

const (
	UserIDKey    = "user_id"
	UserEmailKey = "user_email"
)

// JWTAuth is a Gin middleware for JWT authentication
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondError(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondError(c, http.StatusUnauthorized, "invalid authorization header")
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := security.ValidateJWT(token, jwtSecret)
		if err != nil {
			utils.RespondError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from Gin context
func GetUserID(c *gin.Context) string {
	val, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	userID, ok := val.(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserEmail retrieves the authenticated user email from Gin context
func GetUserEmail(c *gin.Context) string {
	val, exists := c.Get(UserEmailKey)
	if !exists {
		return ""
	}
	email, ok := val.(string)
	if !ok {
		return ""
	}
	return email
}

package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/user"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
	"github.com/BerylCAtieno/paystack-wallet/internal/repository"
	"github.com/BerylCAtieno/paystack-wallet/internal/security"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	userRepo      *repository.UserRepository
	walletService *wallet.Service
	oauthConfig   *oauth2.Config
	jwtSecret     string
}

func NewAuthHandler(userRepo *repository.UserRepository, walletService *wallet.Service, googleClientID, googleClientSecret, redirectURL, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:      userRepo,
		walletService: walletService,
		jwtSecret:     jwtSecret,
		oauthConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Println("Google OAuth redirect URL:", url)
	c.Redirect(307, url)
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		utils.RespondError(c, 400, "code not found")
		return
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		utils.RespondError(c, 500, "failed to exchange token")
		return
	}

	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		utils.RespondError(c, 500, "failed to get user info")
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		utils.RespondError(c, 500, "failed to decode user info")
		return
	}

	// Get or create user
	u, err := h.userRepo.GetByGoogleID(googleUser.ID)
	if err != nil {
		u = &user.User{
			ID:        security.GenerateID(),
			Email:     googleUser.Email,
			Name:      googleUser.Name,
			GoogleID:  googleUser.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := h.userRepo.Create(u); err != nil {
			utils.RespondError(c, 500, "failed to create user")
			return
		}

		if _, err := h.walletService.CreateWallet(u.ID); err != nil {
			utils.RespondError(c, 500, "failed to create wallet")
			return
		}
	}

	jwtToken, err := security.GenerateJWT(u.ID, u.Email, h.jwtSecret)
	if err != nil {
		utils.RespondError(c, 500, "failed to generate token")
		return
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"token": jwtToken,
		"user":  u,
	})
}

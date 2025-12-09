package router

import (
	"net/http"

	"github.com/BerylCAtieno/paystack-wallet/internal/api/handlers"
	"github.com/BerylCAtieno/paystack-wallet/internal/api/middleware"
	"github.com/BerylCAtieno/paystack-wallet/internal/config"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
	"github.com/BerylCAtieno/paystack-wallet/internal/paystack"
	"github.com/BerylCAtieno/paystack-wallet/internal/repository"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Engine        *gin.Engine
	cfg           *config.Config
	authService   *auth.Service
	walletService *wallet.Service
	walletRepo    *repository.WalletRepository
	userRepo      *repository.UserRepository
}

func NewRouter(
	cfg *config.Config,
	authService *auth.Service,
	walletService *wallet.Service,
	walletRepo *repository.WalletRepository,
	userRepo *repository.UserRepository,
) *Router {
	engine := gin.Default()
	r := &Router{
		Engine:        engine,
		cfg:           cfg,
		authService:   authService,
		walletService: walletService,
		walletRepo:    walletRepo,
		userRepo:      userRepo,
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// Health check
	r.Engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Auth routes
	authHandler := handlers.NewAuthHandler(
		r.userRepo,
		r.walletService,
		r.cfg.GoogleClientID,
		r.cfg.GoogleClientSecret,
		r.cfg.GoogleRedirectURL,
		r.cfg.JWTSecret,
	)
	r.Engine.GET("/auth/google", authHandler.GoogleLogin)
	r.Engine.GET("/auth/google/callback", authHandler.GoogleCallback)

	// API Key routes (requires JWT)
	apiKeyHandler := handlers.NewAPIKeyHandler(r.authService)
	keysGroup := r.Engine.Group("/keys")
	keysGroup.Use(middleware.JWTAuth(r.cfg.JWTSecret))
	keysGroup.POST("/create", apiKeyHandler.CreateAPIKey)
	keysGroup.POST("/rollover", apiKeyHandler.RolloverAPIKey)

	// Wallet routes (JWT or API Key)
	paystackClient := paystack.NewClient(r.cfg.PaystackSecretKey)
	walletHandler := handlers.NewWalletHandler(r.walletService, r.walletRepo, paystackClient)
	walletGroup := r.Engine.Group("/wallet")
	walletGroup.Use(middleware.FlexibleAuthGin(r.cfg.JWTSecret, r.authService))
	walletGroup.POST("/deposit", walletHandler.InitiateDeposit)
	walletGroup.GET("/balance", walletHandler.GetBalance)
	walletGroup.POST("/transfer", walletHandler.Transfer)
	walletGroup.GET("/transactions", walletHandler.GetTransactions)

	// Webhook routes (no auth - validated by signature)
	webhookHandler := handlers.NewWebhookHandler(r.walletService, r.cfg.PaystackSecretKey)
	r.Engine.POST("/wallet/paystack/webhook", webhookHandler.HandlePaystackWebhook)
	r.Engine.GET("/wallet/deposit/:reference/status", webhookHandler.GetDepositStatus)
}

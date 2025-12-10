package router

import (
	"net/http"

	"github.com/BerylCAtieno/paystack-wallet/docs"
	"github.com/BerylCAtieno/paystack-wallet/internal/api/handlers"
	"github.com/BerylCAtieno/paystack-wallet/internal/api/middleware"
	"github.com/BerylCAtieno/paystack-wallet/internal/config"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
	"github.com/BerylCAtieno/paystack-wallet/internal/paystack"
	"github.com/BerylCAtieno/paystack-wallet/internal/repository"
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// CORS Configuration
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080/docs", "https://paystack-wallet.fly.dev/docs", "https://paystack-wallet-beryl-673dde33fda9.herokuapp.com/"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "x-api-key", "x-paystack-signature"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

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
	r.Engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the Wallet Service")
	})

	r.Engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "Service is running")
	})

	// SWAGGER DOCUMENTATION
	r.Engine.GET("/swagger.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", docs.SwaggerYAML)
	})

	// If the file doesn't exist, return a helpful error
	r.Engine.GET("/swagger-check", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Swagger file should be at ./docs/swagger.yaml",
			"url":     "/swagger.yaml",
		})
	})

	// Swagger UI handler
	r.Engine.GET("/docs/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger.yaml"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	// AUTH ROUTES
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

	// API KEY ROUTES (JWT)
	apiKeyHandler := handlers.NewAPIKeyHandler(r.authService)

	keysGroup := r.Engine.Group("/keys")
	keysGroup.Use(middleware.JWTAuth(r.cfg.JWTSecret))
	{
		keysGroup.POST("/create", apiKeyHandler.CreateAPIKey)
		keysGroup.POST("/rollover", apiKeyHandler.RolloverAPIKey)
	}

	// WALLET ROUTES (JWT/API KEY)
	paystackClient := paystack.NewClient(r.cfg.PaystackSecretKey)
	walletHandler := handlers.NewWalletHandler(r.walletService, r.walletRepo, paystackClient)

	walletGroup := r.Engine.Group("/wallet")
	{
		walletGroup.POST(
			"/deposit",
			middleware.FlexibleAuth(r.cfg.JWTSecret, r.authService, auth.PermissionDeposit),
			walletHandler.InitiateDeposit,
		)

		walletGroup.GET(
			"/balance",
			middleware.FlexibleAuth(r.cfg.JWTSecret, r.authService, auth.PermissionRead),
			walletHandler.GetBalance,
		)

		walletGroup.POST(
			"/transfer",
			middleware.FlexibleAuth(r.cfg.JWTSecret, r.authService, auth.PermissionTransfer),
			walletHandler.Transfer,
		)

		walletGroup.GET(
			"/transactions",
			middleware.FlexibleAuth(r.cfg.JWTSecret, r.authService, auth.PermissionRead),
			walletHandler.GetTransactions,
		)
	}

	// WEBHOOK
	webhookHandler := handlers.NewWebhookHandler(r.walletService, r.cfg.PaystackSecretKey)

	r.Engine.POST("/wallet/paystack/webhook", webhookHandler.HandlePaystackWebhook)
	r.Engine.GET("/wallet/deposit/:reference/status", webhookHandler.GetDepositStatus)
}

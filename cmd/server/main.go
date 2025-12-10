package main

import (
	"fmt"
	"log"

	"github.com/BerylCAtieno/paystack-wallet/internal/api/router"
	"github.com/BerylCAtieno/paystack-wallet/internal/config"
	"github.com/BerylCAtieno/paystack-wallet/internal/database"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
	"github.com/BerylCAtieno/paystack-wallet/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewSQLite(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, ""); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database initialized successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	// Initialize services
	walletService := wallet.NewService(walletRepo, transactionRepo)
	authService := auth.NewService(apiKeyRepo)

	// Initialize router
	r := router.NewRouter(cfg, authService, walletService, walletRepo, userRepo)

	// Start server
	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Engine.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

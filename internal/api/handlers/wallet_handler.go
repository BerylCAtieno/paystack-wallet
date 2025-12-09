package handlers

import (
	"fmt"

	"github.com/BerylCAtieno/paystack-wallet/internal/api/middleware"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/auth"
	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
	"github.com/BerylCAtieno/paystack-wallet/internal/paystack"
	"github.com/BerylCAtieno/paystack-wallet/internal/repository"
	"github.com/BerylCAtieno/paystack-wallet/internal/utils"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService  *wallet.Service
	walletRepo     *repository.WalletRepository
	paystackClient *paystack.Client
}

func NewWalletHandler(walletService *wallet.Service, walletRepo *repository.WalletRepository, paystackClient *paystack.Client) *WalletHandler {
	return &WalletHandler{
		walletService:  walletService,
		walletRepo:     walletRepo,
		paystackClient: paystackClient,
	}
}

func (h *WalletHandler) getUserID(c *gin.Context) string {
	// Try JWT first
	if userID := middleware.GetUserID(c); userID != "" {
		return userID
	}
	// Try API Key
	if userID := middleware.GetAPIKeyUserID(c); userID != "" {
		return userID
	}
	return ""
}

func (h *WalletHandler) checkPermission(c *gin.Context, perm auth.Permission) bool {
	// JWT users have all permissions
	if middleware.GetUserID(c) != "" {
		return true
	}

	// Check API key permissions
	if apiKey := middleware.GetAPIKey(c); apiKey != nil {
		return apiKey.HasPermission(perm)
	}

	return false
}

type DepositRequest struct {
	Amount int64 `json:"amount"`
}

func (h *WalletHandler) InitiateDeposit(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		utils.RespondError(c, 401, "user not authenticated")
		return
	}

	if !h.checkPermission(c, auth.PermissionDeposit) {
		utils.RespondError(c, 403, "insufficient permissions")
		return
	}

	var req DepositRequest
	if err := c.BindJSON(&req); err != nil {
		utils.RespondError(c, 400, "invalid request body")
		return
	}

	if req.Amount <= 0 {
		utils.RespondError(c, 400, "amount must be greater than 0")
		return
	}

	userWallet, err := h.walletRepo.GetByUserID(userID)
	if err != nil {
		utils.RespondError(c, 500, "wallet not found")
		return
	}

	reference := fmt.Sprintf("DEP_%s_%d", userID, req.Amount)

	email := middleware.GetUserEmail(c)
	if email == "" {
		email = "user@example.com" // fallback for API key auth
	}

	paystackResp, err := h.paystackClient.InitializeTransaction(email, req.Amount, reference)
	if err != nil {
		utils.RespondError(c, 500, "failed to initialize payment")
		return
	}

	_, err = h.walletService.InitiateDeposit(userWallet.ID, req.Amount, reference)
	if err != nil {
		utils.RespondError(c, 500, "failed to create transaction")
		return
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"reference":         paystackResp.Data.Reference,
		"authorization_url": paystackResp.Data.AuthorizationURL,
	})
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		utils.RespondError(c, 401, "user not authenticated")
		return
	}

	if !h.checkPermission(c, auth.PermissionRead) {
		utils.RespondError(c, 403, "insufficient permissions")
		return
	}

	balance, err := h.walletService.GetBalance(userID)
	if err != nil {
		utils.RespondError(c, 500, "failed to get balance")
		return
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"balance": balance,
	})
}

type TransferRequest struct {
	WalletNumber string `json:"wallet_number"`
	Amount       int64  `json:"amount"`
}

func (h *WalletHandler) Transfer(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		utils.RespondError(c, 401, "user not authenticated")
		return
	}

	if !h.checkPermission(c, auth.PermissionTransfer) {
		utils.RespondError(c, 403, "insufficient permissions")
		return
	}

	var req TransferRequest
	if err := c.BindJSON(&req); err != nil {
		utils.RespondError(c, 400, "invalid request body")
		return
	}

	senderWallet, err := h.walletRepo.GetByUserID(userID)
	if err != nil {
		utils.RespondError(c, 500, "wallet not found")
		return
	}

	if err := h.walletService.Transfer(senderWallet.ID, req.WalletNumber, req.Amount); err != nil {
		if err == wallet.ErrInsufficientBalance {
			utils.RespondError(c, 400, "insufficient balance")
			return
		}
		if err == wallet.ErrWalletNotFound {
			utils.RespondError(c, 400, "recipient wallet not found")
			return
		}
		utils.RespondError(c, 500, "transfer failed")
		return
	}

	utils.RespondSuccess(c, map[string]interface{}{
		"status":  "success",
		"message": "Transfer completed",
	})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		utils.RespondError(c, 401, "user not authenticated")
		return
	}

	if !h.checkPermission(c, auth.PermissionRead) {
		utils.RespondError(c, 403, "insufficient permissions")
		return
	}

	userWallet, err := h.walletRepo.GetByUserID(userID)
	if err != nil {
		utils.RespondError(c, 500, "wallet not found")
		return
	}

	transactions, err := h.walletService.GetTransactions(userWallet.ID)
	if err != nil {
		utils.RespondError(c, 500, "failed to get transactions")
		return
	}

	utils.RespondSuccess(c, transactions)
}

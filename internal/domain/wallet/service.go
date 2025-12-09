package wallet

import (
	"fmt"
	"time"

	"github.com/BerylCAtieno/paystack-wallet/internal/security"
)

type Service struct {
	walletRepo      WalletRepository
	transactionRepo TransactionRepository
}

func NewService(walletRepo WalletRepository, transactionRepo TransactionRepository) *Service {
	return &Service{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
	}
}

type WalletRepository interface {
	Create(wallet *Wallet) error
	GetByUserID(userID string) (*Wallet, error)
	GetByWalletNumber(walletNumber string) (*Wallet, error)
	UpdateBalance(walletID string, amount int64) error
	BeginTx() (Transaction, error)
}

type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByReference(reference string) (*Transaction, error)
	Update(tx *Transaction) error
	ListByWalletID(walletID string) ([]*Transaction, error)
}

type TransactionInterface interface {
	Commit() error
	Rollback() error
	UpdateWalletBalance(walletID string, newBalance int64) error
	GetWallet(walletID string) (*Wallet, error)
}

func (s *Service) CreateWallet(userID string) (*Wallet, error) {
	walletNumber := generateWalletNumber()
	wallet := &Wallet{
		ID:           security.GenerateID(),
		UserID:       userID,
		WalletNumber: walletNumber,
		Balance:      0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.walletRepo.Create(wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *Service) GetOrCreateWallet(userID string) (*Wallet, error) {
	wallet, err := s.walletRepo.GetByUserID(userID)
	if err == nil {
		return wallet, nil
	}

	return s.CreateWallet(userID)
}

func (s *Service) InitiateDeposit(walletID string, amount int64, reference string) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	tx := &Transaction{
		ID:        security.GenerateID(),
		WalletID:  walletID,
		Type:      TransactionTypeDeposit,
		Amount:    amount,
		Status:    TransactionStatusPending,
		Reference: reference,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *Service) CompleteDeposit(reference string) error {
	tx, err := s.transactionRepo.GetByReference(reference)
	if err != nil {
		return err
	}

	if tx.Status == TransactionStatusSuccess {
		return nil // Already processed (idempotency)
	}

	// Update transaction status
	tx.Status = TransactionStatusSuccess
	tx.UpdatedAt = time.Now()

	if err := s.transactionRepo.Update(tx); err != nil {
		return err
	}

	// Update wallet balance
	if err := s.walletRepo.UpdateBalance(tx.WalletID, tx.Amount); err != nil {
		return err
	}

	return nil
}

func (s *Service) Transfer(senderWalletID, recipientWalletNumber string, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	// Get sender wallet
	senderWallet, err := s.walletRepo.GetByUserID("") // We'll fix this
	if err != nil {
		return err
	}

	if senderWallet.Balance < amount {
		return ErrInsufficientBalance
	}

	// Get recipient wallet
	recipientWallet, err := s.walletRepo.GetByWalletNumber(recipientWalletNumber)
	if err != nil {
		return ErrWalletNotFound
	}

	// Create debit transaction for sender
	debitTx := &Transaction{
		ID:              security.GenerateID(),
		WalletID:        senderWallet.ID,
		Type:            TransactionTypeTransfer,
		Amount:          -amount,
		Status:          TransactionStatusSuccess,
		Reference:       generateReference(),
		RecipientWallet: recipientWalletNumber,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create credit transaction for recipient
	creditTx := &Transaction{
		ID:              security.GenerateID(),
		WalletID:        recipientWallet.ID,
		Type:            TransactionTypeReceived,
		Amount:          amount,
		Status:          TransactionStatusSuccess,
		Reference:       debitTx.Reference,
		RecipientWallet: senderWallet.WalletNumber,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Update balances
	if err := s.walletRepo.UpdateBalance(senderWallet.ID, -amount); err != nil {
		return err
	}

	if err := s.walletRepo.UpdateBalance(recipientWallet.ID, amount); err != nil {
		return err
	}

	// Save transactions
	if err := s.transactionRepo.Create(debitTx); err != nil {
		return err
	}

	if err := s.transactionRepo.Create(creditTx); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetBalance(userID string) (int64, error) {
	wallet, err := s.walletRepo.GetByUserID(userID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

func (s *Service) GetTransactions(walletID string) ([]*Transaction, error) {
	return s.transactionRepo.ListByWalletID(walletID)
}

func generateWalletNumber() string {
	return fmt.Sprintf("%013d", time.Now().UnixNano()%10000000000000)
}

func generateReference() string {
	return fmt.Sprintf("TXN_%d", time.Now().UnixNano())
}

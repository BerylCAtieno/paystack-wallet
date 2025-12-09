package repository

import (
	"database/sql"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(w *wallet.Wallet) error {
	query := `INSERT INTO wallets (id, user_id, wallet_number, balance, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, w.ID, w.UserID, w.WalletNumber, w.Balance, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WalletRepository) GetByUserID(userID string) (*wallet.Wallet, error) {
	query := `SELECT id, user_id, wallet_number, balance, created_at, updated_at 
		FROM wallets WHERE user_id = ?`

	w := &wallet.Wallet{}
	err := r.db.QueryRow(query, userID).Scan(
		&w.ID, &w.UserID, &w.WalletNumber, &w.Balance, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WalletRepository) GetByWalletNumber(walletNumber string) (*wallet.Wallet, error) {
	query := `SELECT id, user_id, wallet_number, balance, created_at, updated_at 
		FROM wallets WHERE wallet_number = ?`

	w := &wallet.Wallet{}
	err := r.db.QueryRow(query, walletNumber).Scan(
		&w.ID, &w.UserID, &w.WalletNumber, &w.Balance, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WalletRepository) GetByID(id string) (*wallet.Wallet, error) {
	query := `SELECT id, user_id, wallet_number, balance, created_at, updated_at 
		FROM wallets WHERE id = ?`

	w := &wallet.Wallet{}
	err := r.db.QueryRow(query, id).Scan(
		&w.ID, &w.UserID, &w.WalletNumber, &w.Balance, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WalletRepository) UpdateBalance(walletID string, amount int64) error {
	query := `UPDATE wallets SET balance = balance + ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`

	_, err := r.db.Exec(query, amount, walletID)
	return err
}

func (r *WalletRepository) BeginTx() (wallet.Transaction, error) {
	// For simplicity, we're not implementing full transaction support in this version
	return wallet.Transaction{}, nil
}

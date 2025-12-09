package repository

import (
	"database/sql"

	"github.com/BerylCAtieno/paystack-wallet/internal/domain/wallet"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *wallet.Transaction) error {
	query := `INSERT INTO transactions (id, wallet_id, type, amount, status, reference, recipient_wallet, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		tx.ID, tx.WalletID, tx.Type, tx.Amount, tx.Status,
		tx.Reference, tx.RecipientWallet, tx.Metadata, tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

func (r *TransactionRepository) GetByReference(reference string) (*wallet.Transaction, error) {
	query := `SELECT id, wallet_id, type, amount, status, reference, recipient_wallet, metadata, created_at, updated_at
		FROM transactions WHERE reference = ?`

	tx := &wallet.Transaction{}
	err := r.db.QueryRow(query, reference).Scan(
		&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.Status,
		&tx.Reference, &tx.RecipientWallet, &tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *TransactionRepository) Update(tx *wallet.Transaction) error {
	query := `UPDATE transactions SET status = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.Exec(query, tx.Status, tx.UpdatedAt, tx.ID)
	return err
}

func (r *TransactionRepository) ListByWalletID(walletID string) ([]*wallet.Transaction, error) {
	query := `SELECT id, wallet_id, type, amount, status, reference, recipient_wallet, metadata, created_at, updated_at
		FROM transactions WHERE wallet_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*wallet.Transaction
	for rows.Next() {
		tx := &wallet.Transaction{}
		err := rows.Scan(
			&tx.ID, &tx.WalletID, &tx.Type, &tx.Amount, &tx.Status,
			&tx.Reference, &tx.RecipientWallet, &tx.Metadata, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

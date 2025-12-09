package wallet

import "time"

type Wallet struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	WalletNumber string    `json:"wallet_number"`
	Balance      int64     `json:"balance"` // in cents
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Transaction struct {
	ID              string            `json:"id"`
	WalletID        string            `json:"wallet_id"`
	Type            TransactionType   `json:"type"`
	Amount          int64             `json:"amount"`
	Status          TransactionStatus `json:"status"`
	Reference       string            `json:"reference"`
	RecipientWallet string            `json:"recipient_wallet,omitempty"`
	Metadata        string            `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type TransactionType string

const (
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypeReceived TransactionType = "received"
)

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

package wallet

import "errors"

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrDuplicateReference  = errors.New("duplicate transaction reference")
)

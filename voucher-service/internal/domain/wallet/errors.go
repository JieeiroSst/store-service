package wallet

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient wallet funds")
	ErrInvalidAmount     = errors.New("invalid wallet amount")
	ErrVersionConflict   = errors.New("wallet version conflict")
	ErrWalletNotFound    = errors.New("wallet not found")
)

package domain

import "errors"

var (
	ErrWalletNotFound     = errors.New("wallet not found")
	ErrWalletInactive     = errors.New("wallet is not active")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrVersionConflict    = errors.New("optimistic lock version conflict, please retry")

	ErrCardNotFound       = errors.New("card not found")
	ErrCardExpired        = errors.New("card is expired")
	ErrCardBlocked        = errors.New("card is blocked")

	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrTransactionNotPending = errors.New("transaction is not in pending/authorized state")
	ErrTransactionNotCaptured = errors.New("transaction is not in captured state")

	ErrBatchNotFound      = errors.New("settlement batch not found")
	ErrNetworkNotSupported = errors.New("card network not supported")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
)

package port

import "errors"

var (
	ErrNotFound                 = errors.New("resource not found")
	ErrWalletAlreadyExists      = errors.New("wallet already exists for user")
	ErrWalletNotActive          = errors.New("wallet is not active")
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrInvalidAmount            = errors.New("amount must be positive")
	ErrCurrencyMismatch         = errors.New("currency mismatch")
	ErrSameWallet               = errors.New("cannot transfer to the same wallet")
	ErrInvalidPaymentMethodType = errors.New("invalid payment method type")
	ErrInvalidCardNumber        = errors.New("card account_number must be a valid 13-19 digit card number")

	ErrTransactionNotReversible = errors.New("transaction cannot be reversed")
	ErrTransferNotReversible    = errors.New("transfer cannot be reversed")

	ErrInvalidWalletStatusTransition = errors.New("invalid wallet status transition")
	ErrWalletNotEmpty                = errors.New("wallet must have a zero balance to close")
	ErrPerTransactionLimitExceeded   = errors.New("amount exceeds the per-transaction limit")
	ErrDailyLimitExceeded            = errors.New("amount exceeds the daily limit")

	ErrPaymentRequestNotPending    = errors.New("payment request is not pending")
	ErrPaymentRequestExpired       = errors.New("payment request has expired")
	ErrPaymentRequestPayerMismatch = errors.New("wallet is not the designated payer for this request")
)

package model

import "errors"

type ErrorKind int

const (
	KindInternal ErrorKind = iota
	KindInvalid
	KindNotFound
	KindConflict
	KindUnprocessable
	KindForbidden
	KindLocked
	KindUpstream
)

type Error struct {
	Kind    ErrorKind
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func newError(kind ErrorKind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func Invalid(message string) error {
	return newError(KindInvalid, "INVALID_ARGUMENT", message)
}

func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return newError(KindInternal, "INTERNAL", "an internal error occurred")
}

var (
	ErrWalletNotFound  = newError(KindNotFound, "WALLET_NOT_FOUND", "wallet not found")
	ErrWalletExists    = newError(KindConflict, "WALLET_EXISTS", "user already has a wallet in this currency")
	ErrWalletInactive  = newError(KindUnprocessable, "WALLET_INACTIVE", "wallet is not active")
	ErrWalletClosed    = newError(KindUnprocessable, "WALLET_CLOSED", "wallet is closed")
	ErrWalletNotEmpty  = newError(KindUnprocessable, "WALLET_NOT_EMPTY", "wallet must have zero balance and no held funds to be closed")
	ErrNotWalletOwner  = newError(KindForbidden, "FORBIDDEN", "wallet does not belong to this user")
	ErrVersionConflict = newError(KindConflict, "VERSION_CONFLICT", "wallet was modified concurrently, please retry")
	ErrInvalidTier     = newError(KindInvalid, "INVALID_TIER", "unknown wallet tier")

	ErrInsufficientFunds   = newError(KindUnprocessable, "INSUFFICIENT_FUNDS", "insufficient funds")
	ErrInvalidAmount       = newError(KindInvalid, "INVALID_AMOUNT", "amount must be greater than zero")
	ErrAmountPrecision     = newError(KindInvalid, "INVALID_AMOUNT", "amount has more decimal places than the currency allows")
	ErrUnsupportedCurrency = newError(KindInvalid, "UNSUPPORTED_CURRENCY", "currency is not supported")
	ErrCurrencyMismatch    = newError(KindUnprocessable, "CURRENCY_MISMATCH", "wallets must use the same currency")
	ErrSameWallet          = newError(KindInvalid, "SAME_WALLET", "source and destination wallet must differ")
	ErrPerTransactionLimit = newError(KindUnprocessable, "LIMIT_PER_TRANSACTION", "amount exceeds the per-transaction limit for this wallet tier")
	ErrDailyLimit          = newError(KindUnprocessable, "LIMIT_DAILY", "amount exceeds the remaining daily limit for this wallet tier")

	ErrPinNotSet  = newError(KindUnprocessable, "PIN_NOT_SET", "a transaction PIN must be set before moving money out of the wallet")
	ErrPinInvalid = newError(KindForbidden, "PIN_INVALID", "transaction PIN is incorrect")
	ErrPinLocked  = newError(KindLocked, "PIN_LOCKED", "too many incorrect PIN attempts, please try again later")
	ErrPinFormat  = newError(KindInvalid, "PIN_FORMAT", "PIN must be exactly 6 digits")

	ErrCardNotFound    = newError(KindNotFound, "CARD_NOT_FOUND", "card not found")
	ErrCardExpired     = newError(KindUnprocessable, "CARD_DECLINED", "card is expired")
	ErrCardBlocked     = newError(KindUnprocessable, "CARD_DECLINED", "card is blocked")
	ErrCardNotActive   = newError(KindUnprocessable, "CARD_NOT_BLOCKED", "only blocked cards can be unblocked")
	ErrCardNetwork     = newError(KindInvalid, "NETWORK_NOT_SUPPORTED", "card network not supported")
	ErrCardNumberTaken = newError(KindConflict, "CARD_NUMBER_TAKEN", "generated card number already exists")
	ErrCardIssuing     = newError(KindUnprocessable, "CARD_ISSUING_DISABLED", "card issuing is not configured")

	ErrBankNotFound      = newError(KindNotFound, "BANK_NOT_FOUND", "bank not found")
	ErrMerchantNotFound  = newError(KindNotFound, "MERCHANT_NOT_FOUND", "merchant not found")
	ErrBankAccountFound  = newError(KindConflict, "BANK_ACCOUNT_EXISTS", "bank account is already linked")
	ErrBankAccountAbsent = newError(KindNotFound, "BANK_ACCOUNT_NOT_FOUND", "bank account not found")
	ErrBankAccountNumber = newError(KindInvalid, "INVALID_ACCOUNT_NUMBER", "account number must be 6 to 20 digits")
	ErrBankDeclined      = newError(KindUpstream, "BANK_DECLINED", "the bank declined the transfer")

	ErrTransactionNotFound   = newError(KindNotFound, "TRANSACTION_NOT_FOUND", "transaction not found")
	ErrTransactionState      = newError(KindUnprocessable, "INVALID_TRANSACTION_STATE", "transaction is not in a state that allows this operation")
	ErrNotRefundable         = newError(KindUnprocessable, "NOT_REFUNDABLE", "only completed wallet payments can be refunded")
	ErrRefundExceeds         = newError(KindUnprocessable, "REFUND_EXCEEDS_REMAINING", "refund amount exceeds the remaining refundable amount")
	ErrDuplicateTransaction  = newError(KindConflict, "DUPLICATE_TRANSACTION", "duplicate transaction")
	ErrIdempotencyKeyReused  = newError(KindConflict, "IDEMPOTENCY_KEY_REUSED", "idempotency key was already used for a different operation")
	ErrIdempotencyKeyMissing = newError(KindInvalid, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required for this endpoint")

	ErrBatchNotFound   = newError(KindNotFound, "BATCH_NOT_FOUND", "settlement batch not found")
	ErrBatchState      = newError(KindUnprocessable, "INVALID_BATCH_STATE", "settlement batch is not in a state that allows this operation")
	ErrNothingToSettle = newError(KindUnprocessable, "NOTHING_TO_SETTLE", "no captured card transactions to settle for this merchant")

	ErrPaymentRequestNotFound = newError(KindNotFound, "PAYMENT_REQUEST_NOT_FOUND", "payment request not found")
	ErrPaymentRequestState    = newError(KindUnprocessable, "PAYMENT_REQUEST_NOT_PENDING", "payment request is no longer pending")
	ErrPaymentRequestExpired  = newError(KindUnprocessable, "PAYMENT_REQUEST_EXPIRED", "payment request has expired")
	ErrPaymentRequestSelf     = newError(KindInvalid, "PAYMENT_REQUEST_SELF", "cannot request money from yourself")
	ErrNotPaymentRequestParty = newError(KindForbidden, "FORBIDDEN", "user is not allowed to act on this payment request")
)

package port

import "errors"

var (
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthenticated = errors.New("a valid bearer token is required")
	ErrUnauthorized    = errors.New("admin token required")
	ErrForbidden       = errors.New("not allowed to modify this resource")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidUser     = errors.New("user_id is required")
	ErrAlreadyExists   = errors.New("resource already exists")

	ErrInvalidOutcome      = errors.New("outcome must be yes or no")
	ErrInvalidPrice        = errors.New("price is outside the market's tick range")
	ErrInvalidSize         = errors.New("size is below the market minimum or not positive")
	ErrMarketNotTradable   = errors.New("market is not open for trading")
	ErrInsufficientBalance = errors.New("insufficient trading balance")
	ErrInsufficientShares  = errors.New("not enough free shares in position")
	ErrNotFilled           = errors.New("order could not be filled completely (fill-or-kill)")
	ErrOrderNotOpen        = errors.New("order is not open")

	ErrInvalidTransition  = errors.New("market is not in a state that allows this action")
	ErrDisputeWindowOpen  = errors.New("dispute window is still open")
	ErrNegRisk            = errors.New("neg-risk events are resolved as a whole: use the event endpoints")
	ErrNotNegRisk         = errors.New("event is not a neg-risk event")
	ErrReferralNotAllowed = errors.New("referral codes can only be redeemed by new users who were not referred before")
	ErrDisputeWindowShut  = errors.New("dispute window has closed")

	ErrUpstream          = errors.New("upstream service request failed")
	ErrWalletNotFound    = errors.New("wallet not found for user")
	ErrInsufficientFunds = errors.New("insufficient wallet balance")
	ErrWalletRejected    = errors.New("wallet service rejected the transfer")
	ErrWalletUnavailable = errors.New("wallet service request failed")
)

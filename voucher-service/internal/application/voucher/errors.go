package voucher

import "errors"

var (
	ErrRedeemInProgress           = errors.New("another redeem for this voucher is already in progress")
	ErrLockUnavailable            = errors.New("locking backend unavailable")
	ErrDuplicateRequestInProgress = errors.New("duplicate request already in progress")
	ErrProviderTimeout            = errors.New("merchant provider timed out")
	ErrProviderRejected           = errors.New("merchant provider rejected the request")
	ErrTransactionFailed          = errors.New("transaction commit failed")
)

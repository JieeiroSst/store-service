package distribution

import "errors"

var (
	errAlreadyClaimed   = errors.New("claim token already used")
	errClaimExpired     = errors.New("claim token expired")
	errNoVoucherForClaim = errors.New("claim has no associated voucher")
)

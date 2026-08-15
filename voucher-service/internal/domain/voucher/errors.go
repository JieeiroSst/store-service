package voucher

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid voucher status transition")
	ErrVoucherExpired    = errors.New("voucher expired")
	ErrAlreadyRedeemed   = errors.New("voucher already redeemed")
	ErrInvalidPIN        = errors.New("invalid voucher pin")
	ErrInsufficientValue = errors.New("redeem amount exceeds voucher value")
	ErrVersionConflict   = errors.New("voucher version conflict")
	ErrVoucherNotFound   = errors.New("voucher not found")
	ErrInvalidVoucher    = errors.New("invalid voucher")
)

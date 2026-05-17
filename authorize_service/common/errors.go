package common

import "errors"

// Domain errors — use errors.Is() for comparison, never string matching.
var (
	ErrNotFound        = errors.New("record not found")
	ErrDBFailed        = errors.New("database operation failed")
	ErrEnforcerFailed  = errors.New("casbin enforcer failed")
	ErrNotAllowed      = errors.New("access denied")
	ErrOTPFailed       = errors.New("OTP authorization failed")
	ErrOTPLimit        = errors.New("OTP creation limit exceeded")
	ErrInvalidField    = errors.New("invalid update field")
)

package common

import "errors"

// Domain errors — use errors.Is() for comparison, never string matching.
var (
	ErrNotFound       = errors.New("record not found")
	ErrDBFailed       = errors.New("database operation failed")
	ErrEnforcerFailed = errors.New("casbin enforcer failed")
	ErrNotAllowed     = errors.New("access denied")
	ErrOTPFailed      = errors.New("OTP authorization failed")
	ErrOTPLimit       = errors.New("OTP creation limit exceeded")
	ErrInvalidField   = errors.New("invalid update field")

	ErrInvalidUserID     = errors.New("user_id is required")
	ErrUnknownRole       = errors.New("unknown role")
	ErrRoleNotAssignable = errors.New("role can only be granted through bootstrap config")

	ErrMFAInvalidCode     = errors.New("authentication code is invalid or already used")
	ErrMFANotEnabled      = errors.New("two-factor authentication is not enabled")
	ErrMFAAlreadyEnabled  = errors.New("two-factor authentication is already enabled")
	ErrMFANotSetUp        = errors.New("two-factor authentication has not been set up")
	ErrMFATooManyAttempts = errors.New("too many invalid authentication codes, please try again later")
)

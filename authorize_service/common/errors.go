package common

import "errors"

var (
	NotFound       = errors.New("Not Found")
	OTPFailed      = errors.New("There was an error in sending the OTP. Please enter a valid email id or contact site user")
	NotAllow       = errors.New("can't allow users")
	FailedDB       = errors.New("failed to load policy from DB")
	Failedenforcer = errors.New("failed to create casbin enforcer")
	OTPLimmit      = errors.New("pass 5 token generation attempts wait 24 hours to be able to call token")
)

package common

import "errors"

var (
	NotFound  = errors.New("Not Found")
	OTPFailed = errors.New("There was an error in sending the OTP. Please enter a valid email id or contact site user")
)

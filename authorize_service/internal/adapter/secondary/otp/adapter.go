package otp

import (
	"time"

	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	pkgotp "github.com/JieeiroSst/authorize-service/pkg/otp"
)

type otpAdapter struct {
	otp pkgotp.OTP
}

func NewOTPAdapter(secret string) port.OTPPort {
	return &otpAdapter{otp: pkgotp.NewOtp(secret)}
}

func (a *otpAdapter) GenerateOTP(username string) (string, time.Time, error) {
	result, err := a.otp.CreateOtpByUser(username)
	if err != nil {
		return "", time.Time{}, err
	}
	return result.OTP, result.ExpiresAt, nil
}

func (a *otpAdapter) ValidateOTP(otpCode, username string) error {
	return a.otp.Authorize(otpCode, username)
}

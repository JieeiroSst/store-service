package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/zap"
)

const (
	keyOTPCount    = "otp:count:%s" // per-user attempt counter key
	otpMaxAttempts = 5
)

type otpService struct {
	otp   port.OTPPort
	cache port.CachePort
}

func NewOTPService(otp port.OTPPort, cache port.CachePort) port.OTPUsecase {
	return &otpService{otp: otp, cache: cache}
}

func (s *otpService) CreateOtpByUser(ctx context.Context, username string) (string, time.Time, error) {
	lg := logger.WithContext(ctx)
	key := fmt.Sprintf(keyOTPCount, username)

	count, _ := s.cache.GetInt(ctx, key)
	if count >= otpMaxAttempts {
		lg.Debug("CreateOtpByUser: rate limit hit", zap.String("username", username))
		return "", time.Time{}, common.ErrOTPLimit
	}

	if err := s.cache.SetInt(ctx, key, count+1); err != nil {
		lg.Error("CreateOtpByUser: SetInt failed", zap.Error(err))
	}

	token, expiresAt, err := s.otp.GenerateOTP(username)
	if err != nil {
		lg.Error("CreateOtpByUser: GenerateOTP failed", zap.Error(err))
		return "", time.Time{}, common.ErrOTPFailed
	}
	return token, expiresAt, nil
}

func (s *otpService) Authorize(ctx context.Context, otpCode string, username string) error {
	lg := logger.WithContext(ctx)
	if err := s.otp.ValidateOTP(otpCode, username); err != nil {
		lg.Error("Authorize OTP", zap.Error(err))
		return common.ErrOTPFailed
	}
	return nil
}

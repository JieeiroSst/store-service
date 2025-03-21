package usecase

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/dto"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
	"go.uber.org/zap"
)

type Otps interface {
	CreateOtpByUser(ctx context.Context, username string) (*dto.OTP, error)
	Authorize(ctx context.Context, otp string, username string) error
}

type OTPUsecase struct {
	otp         otp.OTP
	cacheHelper cache.CacheHelper
}

var (
	KeyCountCreateToken = "count_create_token_%v"
)

func NewOTPUsecase(otp otp.OTP, cacheHelper cache.CacheHelper) *OTPUsecase {
	return &OTPUsecase{
		otp:         otp,
		cacheHelper: cacheHelper,
	}
}

func (u *OTPUsecase) CreateOtpByUser(ctx context.Context, username string) (*dto.OTP, error) {
	lg := logger.WithContext(ctx)
	key := fmt.Sprintf(KeyCountCreateToken, username)
	count, _ := u.cacheHelper.GetInt(ctx, key)
	if count >= 5 {
		lg.Debug("", zap.Any("limit otp", common.OTPLimmit))
		return nil, common.OTPLimmit
	}
	if err := u.cacheHelper.SetInt(ctx, key, count+1); err != nil {
		lg.Error("CreateOtpByUser", zap.Any("SetInt", err))
	}

	otp, err := u.otp.CreateOtpByUser(username)
	if err != nil {
		lg.Error("CreateOtpByUser", zap.Error(err))
		return nil, err
	}

	return otp, nil
}

func (u *OTPUsecase) Authorize(ctx context.Context, otp string, username string) error {
	lg := logger.WithContext(ctx)
	if err := u.otp.Authorize(otp, username); err != nil {
		lg.Error("Authorize", zap.Error(err))
		return err
	}
	return nil
}

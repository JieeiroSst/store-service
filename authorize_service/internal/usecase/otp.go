package usecase

import (
	"context"
	"fmt"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/dto"
	"github.com/JieeiroSst/authorize-service/pkg/cache"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
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
	key := fmt.Sprintf(KeyCountCreateToken, username)
	count, _ := u.cacheHelper.GetInt(ctx, key)
	if count >= 5 {
		return nil, common.OTPLimmit
	}
	u.cacheHelper.SetInt(ctx, key, count+1)

	otp, err := u.otp.CreateOtpByUser(username)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return otp, nil
}

func (u *OTPUsecase) Authorize(ctx context.Context, otp string, username string) error {
	if err := u.otp.Authorize(otp, username); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

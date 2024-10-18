package usecase

import (
	"context"

	"github.com/JieeiroSst/authorize-service/dto"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
)

type Otps interface {
	CreateOtpByUser(ctx context.Context, username string) (*dto.OTP, error)
	Authorize(ctx context.Context, otp string, username string) error
}

type OTPUsecase struct {
	otp otp.OTP
}

func NewOTPUsecase(otp otp.OTP) *OTPUsecase {
	return &OTPUsecase{
		otp: otp,
	}
}

func (u *OTPUsecase) CreateOtpByUser(ctx context.Context, username string) (*dto.OTP, error) {
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

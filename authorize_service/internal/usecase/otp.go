package usecase

import (
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/otp"
)

type Otps interface {
	CreateOtpByUser(username string) (*model.OTP, error)
	Authorize(otp string, username string) error
}

type OTPUsecase struct {
	otp otp.OTP
}

func NewOTPUsecase(otp otp.OTP) *OTPUsecase {
	return &OTPUsecase{
		otp: otp,
	}
}

func (u *OTPUsecase) CreateOtpByUser(username string) (*model.OTP, error) {
	otp, err := u.otp.CreateOtpByUser(username)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return otp, nil
}

func (u *OTPUsecase) Authorize(otp string, username string) error {
	if err := u.otp.Authorize(otp, username); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

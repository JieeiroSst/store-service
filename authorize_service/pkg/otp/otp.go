package otp

import (
	"fmt"
	"strings"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/jltorresm/otpgo"
	"github.com/jltorresm/otpgo/config"
)

type otp struct {
	serect string
}

type OTP interface {
	generate(username string) otpgo.TOTP
	CreateOtpByUser(username string) (*model.OTP, error)
	Authorize(otp string, username string) error
}

func NewOtp(serect string) OTP {
	return &otp{
		serect: serect,
	}
}

func (o *otp) generate(username string) otpgo.TOTP {
	serect := fmt.Sprintf("%s%s", o.serect, strings.ToUpper(username))
	return otpgo.TOTP{
		Key:       serect,
		Period:    30,
		Delay:     1,
		Algorithm: config.HmacSHA1,
		Length:    6,
	}
}

func (o *otp) CreateOtpByUser(username string) (*model.OTP, error) {
	totp := o.generate(username)
	token, err := totp.Generate()
	if err != nil {
		return nil, err
	}
	return &model.OTP{
		OTP: token,
	}, nil
}

func (o *otp) Authorize(otp string, username string) error {
	totp := o.generate(username)
	ok, err := totp.Validate(otp)
	if err != nil {
		return err
	}
	if !ok {
		return common.OTPFailed
	}
	return nil
}

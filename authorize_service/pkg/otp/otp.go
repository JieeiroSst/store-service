package otp

import (
	"fmt"
	"strings"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
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
		log.Error(err.Error())
		return nil, err
	}
	log.Info(token)
	return &model.OTP{
		OTP: token,
	}, nil
}

func (o *otp) Authorize(otp string, username string) error {
	totp := o.generate(username)
	ok, err := totp.Validate(otp)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if !ok {
		log.Warn(fmt.Sprintf("Token authorize faild %v", otp))
		return common.OTPFailed
	}
	log.Info("Token authorize success")
	return nil
}

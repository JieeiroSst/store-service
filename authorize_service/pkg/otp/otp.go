package otp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JieeiroSst/authorize-service/dto"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/jltorresm/otpgo"
	"github.com/jltorresm/otpgo/config"
)

var ErrInvalid = errors.New("otp is invalid or expired")

const (
	period = 30
	delay  = 1
)

type otp struct {
	serect string
}

type OTP interface {
	generate(username string) otpgo.TOTP
	CreateOtpByUser(username string) (*dto.OTP, error)
	Authorize(otp string, username string) error
}

func NewOtp(serect string) OTP {
	return &otp{
		serect: serect,
	}
}

func (o *otp) generate(username string) otpgo.TOTP {
	m := hmac.New(sha256.New, []byte(o.serect))
	m.Write([]byte("otp:" + strings.ToUpper(username)))
	return otpgo.TOTP{
		Key:       base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(m.Sum(nil)),
		Period:    period,
		Delay:     delay,
		Algorithm: config.HmacSHA1,
		Length:    6,
	}
}

func (o *otp) CreateOtpByUser(username string) (*dto.OTP, error) {
	totp := o.generate(username)
	token, err := totp.Generate()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &dto.OTP{
		OTP:       token,
		ExpiresAt: ExpiresAt(time.Now()),
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
		log.Warn(fmt.Sprintf("otp rejected for user %s", username))
		return ErrInvalid
	}
	return nil
}

func ExpiresAt(now time.Time) time.Time {
	start := now.Unix() - now.Unix()%period
	return time.Unix(start+period*(delay+1), 0)
}

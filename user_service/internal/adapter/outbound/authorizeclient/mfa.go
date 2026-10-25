package authorizeclient

import (
	"context"
	"strconv"
	"time"

	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/authorizeclient/mfapb"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

const authorizeErrorDomain = "authorize-service"

var mfaReasons = map[string]error{
	"MFA_INVALID_CODE":      domain.ErrMFAInvalidCode,
	"MFA_NOT_ENABLED":       domain.ErrMFANotEnabled,
	"MFA_ALREADY_ENABLED":   domain.ErrMFAAlreadyEnabled,
	"MFA_NOT_SET_UP":        domain.ErrMFANotSetUp,
	"MFA_TOO_MANY_ATTEMPTS": domain.ErrTooManyAttempts,
	"INVALID_USER_ID":       domain.ErrInvalidRequest,
}

type MFAClient struct {
	api     mfapb.MfaServiceClient
	timeout time.Duration
}

func (c *Client) MFA() *MFAClient {
	return &MFAClient{api: mfapb.NewMfaServiceClient(c.conn), timeout: c.timeout}
}

func (c *MFAClient) Status(ctx context.Context, userID int) (domain.MFAStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.GetTOTPStatus(ctx, &mfapb.GetTOTPStatusRequest{UserId: strconv.Itoa(userID)})
	if err != nil {
		return domain.MFAStatus{}, mfaError(err)
	}
	out := domain.MFAStatus{Enabled: res.GetEnabled(), RecoveryCodesRemaining: int(res.GetRecoveryCodesRemaining())}
	if res.GetEnabledAt() > 0 {
		at := time.Unix(res.GetEnabledAt(), 0).UTC()
		out.EnabledAt = &at
	}
	return out, nil
}

func (c *MFAClient) SetupTOTP(ctx context.Context, userID int, account string) (domain.TOTPSetup, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.SetupTOTP(ctx, &mfapb.SetupTOTPRequest{UserId: strconv.Itoa(userID), AccountName: account})
	if err != nil {
		return domain.TOTPSetup{}, mfaError(err)
	}
	return domain.TOTPSetup{
		Secret:     res.GetSecret(),
		OtpauthURI: res.GetOtpauthUri(),
		QRCodePNG:  res.GetQrCodePng(),
		Issuer:     res.GetIssuer(),
		Account:    res.GetAccount(),
		Digits:     int(res.GetDigits()),
		Period:     int(res.GetPeriod()),
		Algorithm:  res.GetAlgorithm(),
	}, nil
}

func (c *MFAClient) EnableTOTP(ctx context.Context, userID int, code string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.EnableTOTP(ctx, &mfapb.TOTPCodeRequest{UserId: strconv.Itoa(userID), Code: code})
	if err != nil {
		return nil, mfaError(err)
	}
	return res.GetRecoveryCodes(), nil
}

func (c *MFAClient) DisableTOTP(ctx context.Context, userID int, code string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.api.DisableTOTP(ctx, &mfapb.TOTPCodeRequest{UserId: strconv.Itoa(userID), Code: code})
	return mfaError(err)
}

func (c *MFAClient) RegenerateRecoveryCodes(ctx context.Context, userID int, code string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.RegenerateRecoveryCodes(ctx, &mfapb.TOTPCodeRequest{UserId: strconv.Itoa(userID), Code: code})
	if err != nil {
		return nil, mfaError(err)
	}
	return res.GetRecoveryCodes(), nil
}

func (c *MFAClient) VerifyTOTP(ctx context.Context, userID int, code string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.api.VerifyTOTP(ctx, &mfapb.VerifyTOTPRequest{UserId: strconv.Itoa(userID), Code: code, AllowRecoveryCode: true})
	return mfaError(err)
}

func mfaError(err error) error {
	if err == nil {
		return nil
	}
	for _, d := range status.Convert(err).Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok && info.GetDomain() == authorizeErrorDomain {
			if mapped, ok := mfaReasons[info.GetReason()]; ok {
				return mapped
			}
		}
	}
	return err
}

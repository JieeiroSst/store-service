package grpc

import (
	"context"
	"fmt"
	"testing"

	mfapb "github.com/JieeiroSst/authorize-service/api/mfa/pb"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubMFA struct{ err error }

func (s stubMFA) Status(context.Context, string) (model.TOTPStatus, error) {
	return model.TOTPStatus{}, s.err
}
func (s stubMFA) SetupTOTP(context.Context, string, string) (model.TOTPSetup, error) {
	return model.TOTPSetup{}, s.err
}
func (s stubMFA) EnableTOTP(context.Context, string, string) ([]string, error) { return nil, s.err }
func (s stubMFA) DisableTOTP(context.Context, string, string) error           { return s.err }
func (s stubMFA) RegenerateRecoveryCodes(context.Context, string, string) ([]string, error) {
	return nil, s.err
}
func (s stubMFA) VerifyTOTP(context.Context, string, string, bool) (model.TOTPVerification, error) {
	return model.TOTPVerification{UsedRecoveryCode: true}, s.err
}

func TestMfaHandlerErrorReasons(t *testing.T) {
	cases := []struct {
		err    error
		code   codes.Code
		reason string
	}{
		{common.ErrMFAInvalidCode, codes.Unauthenticated, "MFA_INVALID_CODE"},
		{fmt.Errorf("wrap: %w", common.ErrMFATooManyAttempts), codes.ResourceExhausted, "MFA_TOO_MANY_ATTEMPTS"},
		{common.ErrMFANotEnabled, codes.FailedPrecondition, "MFA_NOT_ENABLED"},
		{common.ErrInvalidUserID, codes.InvalidArgument, "INVALID_USER_ID"},
	}
	for _, c := range cases {
		_, err := NewMfaHandler(stubMFA{err: c.err}).VerifyTOTP(context.Background(), &mfapb.VerifyTOTPRequest{UserId: "7", Code: "1"})
		st := status.Convert(err)
		if st.Code() != c.code {
			t.Errorf("%v: code = %v, want %v", c.err, st.Code(), c.code)
		}
		var reason string
		for _, d := range st.Details() {
			if info, ok := d.(*errdetails.ErrorInfo); ok && info.GetDomain() == errorDomain {
				reason = info.GetReason()
			}
		}
		if reason != c.reason {
			t.Errorf("%v: reason = %q, want %q", c.err, reason, c.reason)
		}
	}

	_, err := NewMfaHandler(stubMFA{err: fmt.Errorf("pq: boom")}).SetupTOTP(context.Background(), &mfapb.SetupTOTPRequest{UserId: "7"})
	if st := status.Convert(err); st.Code() != codes.Internal || st.Message() != "internal error" {
		t.Errorf("unexpected internal mapping %v", err)
	}

	res, err := NewMfaHandler(stubMFA{}).VerifyTOTP(context.Background(), &mfapb.VerifyTOTPRequest{UserId: "7", Code: "1"})
	if err != nil || !res.GetValid() || !res.GetUsedRecoveryCode() {
		t.Errorf("success res=%v err=%v", res, err)
	}
}

package grpc

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/trace_id"
	mfapb "github.com/JieeiroSst/authorize-service/api/mfa/pb"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errorDomain = "authorize-service"

type MfaHandler struct {
	mfa port.MFAUsecase
	mfapb.UnimplementedMfaServiceServer
}

func NewMfaHandler(mfa port.MFAUsecase) *MfaHandler {
	return &MfaHandler{mfa: mfa}
}

func (h *MfaHandler) GetTOTPStatus(ctx context.Context, in *mfapb.GetTOTPStatusRequest) (*mfapb.TOTPStatus, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	st, err := h.mfa.Status(ctx, in.GetUserId())
	if err != nil {
		return nil, mfaStatus(ctx, "GetTOTPStatus", err)
	}
	out := &mfapb.TOTPStatus{Enabled: st.Enabled, RecoveryCodesRemaining: int32(st.RecoveryCodesRemaining)}
	if st.EnabledAt != nil {
		out.EnabledAt = st.EnabledAt.Unix()
	}
	return out, nil
}

func (h *MfaHandler) SetupTOTP(ctx context.Context, in *mfapb.SetupTOTPRequest) (*mfapb.SetupTOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("SetupTOTP", zap.String("user_id", in.GetUserId()))
	setup, err := h.mfa.SetupTOTP(ctx, in.GetUserId(), in.GetAccountName())
	if err != nil {
		return nil, mfaStatus(ctx, "SetupTOTP", err)
	}
	return &mfapb.SetupTOTPResponse{
		Secret:     setup.Secret,
		OtpauthUri: setup.OtpauthURI,
		QrCodePng:  setup.QRCodePNG,
		Issuer:     setup.Issuer,
		Account:    setup.Account,
		Digits:     int32(setup.Digits),
		Period:     int32(setup.Period),
		Algorithm:  setup.Algorithm,
	}, nil
}

func (h *MfaHandler) EnableTOTP(ctx context.Context, in *mfapb.TOTPCodeRequest) (*mfapb.RecoveryCodesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("EnableTOTP", zap.String("user_id", in.GetUserId()))
	codes, err := h.mfa.EnableTOTP(ctx, in.GetUserId(), in.GetCode())
	if err != nil {
		return nil, mfaStatus(ctx, "EnableTOTP", err)
	}
	return &mfapb.RecoveryCodesResponse{RecoveryCodes: codes}, nil
}

func (h *MfaHandler) DisableTOTP(ctx context.Context, in *mfapb.TOTPCodeRequest) (*mfapb.DisableTOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("DisableTOTP", zap.String("user_id", in.GetUserId()))
	if err := h.mfa.DisableTOTP(ctx, in.GetUserId(), in.GetCode()); err != nil {
		return nil, mfaStatus(ctx, "DisableTOTP", err)
	}
	return &mfapb.DisableTOTPResponse{Success: true}, nil
}

func (h *MfaHandler) RegenerateRecoveryCodes(ctx context.Context, in *mfapb.TOTPCodeRequest) (*mfapb.RecoveryCodesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("RegenerateRecoveryCodes", zap.String("user_id", in.GetUserId()))
	codes, err := h.mfa.RegenerateRecoveryCodes(ctx, in.GetUserId(), in.GetCode())
	if err != nil {
		return nil, mfaStatus(ctx, "RegenerateRecoveryCodes", err)
	}
	return &mfapb.RecoveryCodesResponse{RecoveryCodes: codes}, nil
}

func (h *MfaHandler) VerifyTOTP(ctx context.Context, in *mfapb.VerifyTOTPRequest) (*mfapb.VerifyTOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	res, err := h.mfa.VerifyTOTP(ctx, in.GetUserId(), in.GetCode(), in.GetAllowRecoveryCode())
	if err != nil {
		return nil, mfaStatus(ctx, "VerifyTOTP", err)
	}
	return &mfapb.VerifyTOTPResponse{Valid: true, UsedRecoveryCode: res.UsedRecoveryCode}, nil
}

var mfaErrors = []struct {
	err    error
	code   codes.Code
	reason string
}{
	{common.ErrInvalidUserID, codes.InvalidArgument, "INVALID_USER_ID"},
	{common.ErrMFAInvalidCode, codes.Unauthenticated, "MFA_INVALID_CODE"},
	{common.ErrMFANotEnabled, codes.FailedPrecondition, "MFA_NOT_ENABLED"},
	{common.ErrMFAAlreadyEnabled, codes.FailedPrecondition, "MFA_ALREADY_ENABLED"},
	{common.ErrMFANotSetUp, codes.FailedPrecondition, "MFA_NOT_SET_UP"},
	{common.ErrMFATooManyAttempts, codes.ResourceExhausted, "MFA_TOO_MANY_ATTEMPTS"},
}

func mfaStatus(ctx context.Context, method string, err error) error {
	for _, m := range mfaErrors {
		if errors.Is(err, m.err) {
			st, detErr := status.New(m.code, m.err.Error()).WithDetails(&errdetails.ErrorInfo{Reason: m.reason, Domain: errorDomain})
			if detErr != nil {
				return status.Error(m.code, m.err.Error())
			}
			return st.Err()
		}
	}
	logger.WithContext(ctx).Error(method, zap.Error(err))
	return status.Error(codes.Internal, "internal error")
}
